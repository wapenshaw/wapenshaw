package main

// refresh rebuilds the data-driven parts of the profile from public GitHub activity:
//
//	assets/stats-{dark,light}.svg  last 12 months: contributions, upstream PRs, repos, languages
//	assets/pr-*-{dark,light}.svg   cards for the latest upstream PRs (merged or in review)
//	the <!-- upstream --> block in README.md and options/*.md
//
// Only public repositories count, and anything owned by excludeOwners is ignored.
// It authenticates with GH_TOKEN and falls back to `gh auth token`.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var excludeOwners = map[string]bool{"assurant": true}

const cards = 4

// overrides holds hand-written one-liners and corrections for cards, keyed by "owner/repo#number".
// Cards without an entry fall back to the repo description.
var overrides = map[string]struct{ blurb, lang, color string }{
	"kryptamine/herdr-auto-title#120": {blurb: `Windows tabs stop being named Files\PowerShell\7\pwsh.exe`},
	"qintmb/herdr-theme-picker#8":     {blurb: "Go port for Windows, plus themes grown from 20 to 112"},
	"webosbrew/youtube-webos#487":     {blurb: "Autostart on LG TVs waits for the network, not an error page"},
	"TheRealDevKat/farm-check#9":      {blurb: "Spot used Seagate drives sold as new, now on Windows", lang: "PowerShell", color: "#5391FE"},
}

type repo struct {
	NameWithOwner   string
	Description     string
	IsPrivate       bool
	IsFork          bool
	Owner           struct{ Login string }
	PrimaryLanguage *struct{ Name, Color string }
}

func (r repo) public() bool {
	return !r.IsPrivate && !excludeOwners[strings.ToLower(r.Owner.Login)]
}

type byRepo []struct {
	Repository    repo
	Contributions struct{ TotalCount int }
}

type contributions struct {
	CommitContributionsByRepository            byRepo
	PullRequestContributionsByRepository       byRepo
	IssueContributionsByRepository             byRepo
	PullRequestReviewContributionsByRepository byRepo
	RepositoryContributions                    struct{ Nodes []struct{ Repository repo } }
}

type pullRequest struct {
	Number              int
	Title, URL, State   string
	Merged              bool
	MergedAt, CreatedAt time.Time
	Additions           int
	Deletions           int
	Repository          repo
	at                  time.Time
}

const repoFields = "nameWithOwner isPrivate isFork owner { login } primaryLanguage { name color }"

var contributionsQuery = strings.ReplaceAll(`
query($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    contributionsCollection(from: $from, to: $to) {
      commitContributionsByRepository(maxRepositories: 100) { repository { REPO } contributions { totalCount } }
      pullRequestContributionsByRepository(maxRepositories: 100) { repository { REPO } contributions { totalCount } }
      issueContributionsByRepository(maxRepositories: 100) { repository { REPO } contributions { totalCount } }
      pullRequestReviewContributionsByRepository(maxRepositories: 100) { repository { REPO } contributions { totalCount } }
      repositoryContributions(first: 100) { nodes { repository { REPO } } }
    }
  }
}`, "REPO", repoFields)

const upstreamQuery = `
query($q: String!) {
  search(query: $q, type: ISSUE, first: 50) {
    nodes {
      ... on PullRequest {
        number title url state merged mergedAt createdAt additions deletions
        repository { nameWithOwner description isPrivate owner { login } primaryLanguage { name color } }
      }
    }
  }
}`

func token() (string, error) {
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return t, nil
	}
	out, err := exec.Command("gh", "auth", "token").Output()
	return strings.TrimSpace(string(out)), err
}

func graphql(query string, vars map[string]any, into any) error {
	tok, err := token()
	if err != nil {
		return fmt.Errorf("no GH_TOKEN and `gh auth token` failed: %w", err)
	}
	payload, _ := json.Marshal(map[string]any{"query": query, "variables": vars})
	req, _ := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var body struct {
		Data   json.RawMessage
		Errors []struct{ Message string }
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return fmt.Errorf("graphql: %s: %w", res.Status, err)
	}
	if len(body.Errors) > 0 {
		return fmt.Errorf("graphql: %s", body.Errors[0].Message)
	}
	return json.Unmarshal(body.Data, into)
}

func login() string {
	if owner := os.Getenv("GITHUB_REPOSITORY_OWNER"); owner != "" {
		return owner
	}
	return "wapenshaw"
}

// --- data ----------------------------------------------------------------------------------------

func fetch(now time.Time) (since time.Time, cc contributions, prs []pullRequest, err error) {
	since = now.AddDate(0, 0, -365)
	var user struct {
		User struct{ ContributionsCollection contributions }
	}
	err = graphql(contributionsQuery, map[string]any{
		"login": login(), "from": since.Format(time.RFC3339), "to": now.Format(time.RFC3339),
	}, &user)
	if err != nil {
		return
	}
	cc = user.User.ContributionsCollection

	var search struct{ Search struct{ Nodes []pullRequest } }
	q := fmt.Sprintf("is:pr is:public author:%s -user:%s sort:updated-desc", login(), login())
	if err = graphql(upstreamQuery, map[string]any{"q": q}, &search); err != nil {
		return
	}
	for _, p := range search.Search.Nodes {
		if p.Number == 0 || !p.Repository.public() || !(p.Merged || p.State == "OPEN") {
			continue
		}
		p.at = p.CreatedAt
		if p.Merged {
			p.at = p.MergedAt
		}
		prs = append(prs, p)
	}
	sort.SliceStable(prs, func(i, j int) bool { return prs[i].at.After(prs[j].at) })
	return
}

type language struct {
	name, color string
	votes       int
}

type stats struct {
	contributions, merged, repos, lines int
	languages                           []language
}

func summarize(since time.Time, cc contributions, prs []pullRequest) stats {
	var s stats
	repos := map[string]repo{}
	for _, list := range []byRepo{cc.CommitContributionsByRepository, cc.PullRequestContributionsByRepository,
		cc.IssueContributionsByRepository, cc.PullRequestReviewContributionsByRepository} {
		for _, e := range list {
			if e.Repository.public() {
				s.contributions += e.Contributions.TotalCount
				repos[e.Repository.NameWithOwner] = e.Repository
			}
		}
	}
	for _, node := range cc.RepositoryContributions.Nodes {
		if node.Repository.public() && !node.Repository.IsFork { // a fork isn't a contribution
			s.contributions++
			repos[node.Repository.NameWithOwner] = node.Repository
		}
	}
	s.repos = len(repos)

	// one vote per repo, so a single busy scripts repo doesn't drown out everything else
	byName := map[string]*language{}
	for _, r := range repos {
		if r.PrimaryLanguage == nil {
			continue
		}
		l, ok := byName[r.PrimaryLanguage.Name]
		if !ok {
			l = &language{name: r.PrimaryLanguage.Name, color: r.PrimaryLanguage.Color}
			byName[l.name] = l
		}
		l.votes++
	}
	for _, l := range byName {
		s.languages = append(s.languages, *l)
	}
	sort.Slice(s.languages, func(i, j int) bool {
		a, b := s.languages[i], s.languages[j]
		return a.votes > b.votes || a.votes == b.votes && a.name < b.name
	})

	for _, p := range prs {
		if p.Merged && !p.at.Before(since) {
			s.merged++
			s.lines += p.Additions
		}
	}
	return s
}

// --- rendering -----------------------------------------------------------------------------------

func short(v int) string {
	if v < 1000 {
		return strconv.Itoa(v)
	}
	return strings.Replace(fmt.Sprintf("%.1fk", float64(v)/1000), ".0k", "k", 1)
}

func thousands(v int) string {
	s := strconv.Itoa(v)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func statsSVG(s stats) string {
	metrics := [][2]string{
		{strconv.Itoa(s.contributions), "public contributions"},
		{strconv.Itoa(s.merged), "PRs merged upstream"},
		{strconv.Itoa(s.repos), "repos contributed to"},
		{"+" + short(s.lines), "lines merged upstream"},
	}
	var cells, bar, legend strings.Builder
	for i, m := range metrics {
		fmt.Fprintf(&cells, `  <text x="%d" y="96" class="sans" font-size="40" font-weight="700" fill="{text}">%s</text>
  <text x="%d" y="122" class="mono" font-size="12" fill="{muted}">%s</text>
`, 40+i*236, esc(m[0]), 42+i*236, esc(m[1]))
	}
	votes := 0
	for _, l := range s.languages {
		votes += l.votes
	}
	if votes == 0 {
		votes = 1
	}
	x, lx := 40.0, 40.0
	for _, l := range s.languages {
		color := l.color
		if color == "" {
			color = "{muted}"
		}
		w := 920 * float64(l.votes) / float64(votes)
		fmt.Fprintf(&bar, `    <rect x="%.1f" y="152" width="%.1f" height="10" fill="%s"/>`+"\n", x, w, color)
		x += w
		fmt.Fprintf(&legend, `  <circle cx="%s" cy="188" r="5" fill="%s"/><text x="%s" y="192" class="mono" font-size="12" fill="{text}">%s <tspan fill="{muted}">%d</tspan></text>`+"\n",
			n(lx+5), color, n(lx+16), esc(l.name), l.votes)
		lx += 16 + float64(utf8.RuneCountInString(l.name)+3)*7.2 + 18
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1000 216" width="1000" height="216" role="img" aria-label="Last 12 months on public GitHub: %d contributions, %d PRs merged upstream, %d repos">
  <style>
    .mono { font-family: %s; }
    .sans { font-family: %s; }
  </style>
  <defs><clipPath id="bar"><rect x="40" y="152" width="920" height="10" rx="5"/></clipPath></defs>
  <rect x="0.75" y="0.75" width="998.5" height="214.5" rx="14" fill="{bg}" stroke="{border}" stroke-width="1.5"/>
  <text x="40" y="38" class="mono" font-size="12" fill="{text}"><tspan fill="{accent2}">◆</tspan> last 12 months <tspan fill="{muted}">· public repos only</tspan></text>
  <text x="960" y="38" text-anchor="end" class="mono" font-size="12" fill="{muted}">languages: one vote per repo</text>
%s  <g clip-path="url(#bar)">
    <rect x="40" y="152" width="920" height="10" fill="{border}"/>
%s  </g>
%s</svg>
`, s.contributions, s.merged, s.repos, mono, sans, cells.String(), bar.String(), legend.String())
}

func wrap(text string, width int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(text) {
		if line != "" && utf8.RuneCountInString(line)+1+utf8.RuneCountInString(word) > width {
			lines = append(lines, line)
			line = word
		} else {
			line = strings.TrimSpace(line + " " + word)
		}
	}
	return append(lines, line)
}

func card(p pullRequest) string {
	owner, name, _ := strings.Cut(p.Repository.NameWithOwner, "/")
	extra := overrides[fmt.Sprintf("%s#%d", p.Repository.NameWithOwner, p.Number)]
	lang, color := "Code", "{muted}"
	if l := p.Repository.PrimaryLanguage; l != nil {
		lang = l.Name
		if l.Color != "" {
			color = l.Color
		}
	}
	if extra.lang != "" {
		lang = extra.lang
	}
	if extra.color != "" {
		color = extra.color
	}
	blurb := extra.blurb
	if blurb == "" {
		blurb = p.Repository.Description
	}
	if r := []rune(blurb); len(r) > 62 {
		blurb = strings.TrimRight(string(r[:61]), " ") + "…"
	}
	pillFill, pillText := "#238636", "in review"
	if p.Merged {
		pillFill, pillText = "#8957e5", "merged"
	}
	pillW := float64(len(pillText))*7.2 + 24
	var title strings.Builder
	for i, l := range wrap(p.Title, 44) {
		if i == 2 {
			break
		}
		dy := 0
		if i > 0 {
			dy = 22
		}
		fmt.Fprintf(&title, `<tspan x="24" dy="%d">%s</tspan>`, dy, esc(l))
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 480 172" width="480" height="172" role="img" aria-label="%s #%d: %s">
  <style>
    .mono { font-family: %s; }
    .sans { font-family: %s; }
  </style>
  <rect x="0.75" y="0.75" width="478.5" height="170.5" rx="12" fill="{bg}" stroke="{border}" stroke-width="1.5"/>
  <g fill="none" stroke="{muted}" stroke-width="1.6" stroke-linecap="round">
    <circle cx="28" cy="25" r="3"/><circle cx="28" cy="41" r="3"/><line x1="28" y1="28" x2="28" y2="38"/>
    <circle cx="42" cy="41" r="3"/><path d="M42,38 V31 Q42,27 38,27 H34"/><path d="M36,24.5 L33.5,27 L36,29.5"/>
  </g>
  <text x="56" y="38" class="mono" font-size="13"><tspan fill="{muted}">%s/</tspan><tspan fill="{accent}" font-weight="700">%s</tspan></text>
  <rect x="%s" y="20" width="%s" height="24" rx="12" fill="%s"/>
  <text x="%s" y="36" text-anchor="middle" class="mono" font-size="12" font-weight="700" fill="#ffffff">%s</text>
  <text y="76" class="sans" font-size="16" font-weight="600" fill="{text}">%s</text>
  <text x="24" y="126" class="sans" font-size="13" fill="{muted}">%s</text>
  <line x1="24" y1="140" x2="456" y2="140" stroke="{border}"/>
  <circle cx="30" cy="156" r="5" fill="%s"/>
  <text x="42" y="160" class="mono" font-size="12" xml:space="preserve" fill="{muted}">%s  <tspan fill="{ok}">+%s</tspan> <tspan fill="{kw}">−%s</tspan>  #%d</text>
  <text x="456" y="160" text-anchor="end" class="mono" font-size="12" fill="{muted}">%s</text>
</svg>
`, esc(p.Repository.NameWithOwner), p.Number, esc(p.Title), mono, sans,
		esc(owner), esc(name), n(456-pillW), n(pillW), pillFill, n(456-pillW/2), pillText,
		title.String(), esc(blurb), color, esc(lang), thousands(p.Additions), thousands(p.Deletions), p.Number,
		strings.ToLower(p.at.Format("Jan 2006")))
}

func slug(p pullRequest) string {
	_, name, _ := strings.Cut(p.Repository.NameWithOwner, "/")
	return fmt.Sprintf("pr-%s-%d", name, p.Number)
}

var upstreamBlock = regexp.MustCompile(`(?s)(<!-- upstream:start -->\n).*?(<!-- upstream:end -->)`)

func refresh(root string) error {
	now := time.Now().UTC().Truncate(time.Second)
	since, cc, prs, err := fetch(now)
	if err != nil {
		return err
	}
	if len(prs) > cards {
		prs = prs[:cards]
	}
	assets := filepath.Join(root, "assets")

	stats := statsSVG(summarize(since, cc, prs))
	keep := map[string]bool{}
	for _, th := range themes {
		if err := writeText(root, filepath.Join(assets, "stats-"+th.name+".svg"), th.t.paint(stats)); err != nil {
			return err
		}
		for _, p := range prs {
			file := slug(p) + "-" + th.name + ".svg"
			keep[file] = true
			if err := writeText(root, filepath.Join(assets, file), th.t.paint(card(p))); err != nil {
				return err
			}
		}
	}
	old, _ := filepath.Glob(filepath.Join(assets, "pr-*.svg"))
	for _, path := range old {
		if !keep[filepath.Base(path)] {
			if err := os.Remove(path); err != nil {
				return err
			}
			fmt.Println("removed", "assets/"+filepath.Base(path))
		}
	}

	pages, _ := filepath.Glob(filepath.Join(root, "options", "*.md"))
	for _, page := range append([]string{filepath.Join(root, "README.md")}, pages...) {
		prefix := "../assets/"
		if filepath.Dir(page) == root {
			prefix = "assets/"
		}
		var block strings.Builder
		block.WriteString("<p>\n")
		for _, p := range prs {
			fmt.Fprintf(&block, `  <a href="%s"><picture><source media="(prefers-color-scheme: dark)" srcset="%s%s-dark.svg"><img alt="%s #%d: %s" src="%s%s-light.svg" width="49%%"></picture></a>`+"\n",
				p.URL, prefix, slug(p), esc(p.Repository.NameWithOwner), p.Number, esc(p.Title), prefix, slug(p))
		}
		block.WriteString("</p>\n")
		md, err := readText(page)
		if err != nil {
			return err
		}
		updated := upstreamBlock.ReplaceAllStringFunc(md, func(m string) string {
			parts := upstreamBlock.FindStringSubmatch(m)
			return parts[1] + block.String() + parts[2]
		})
		if err := writeText(root, page, updated); err != nil {
			return err
		}
	}
	return nil
}
