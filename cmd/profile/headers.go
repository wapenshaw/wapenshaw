package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	role  = "sr technology automation engineer"
	years = "13+ yrs"
)

var options = []struct {
	key, name string
	build     func() string
}{
	{"a", "autopilot", autopilot},
	{"b", "agent", agent},
	{"c", "polyglot", polyglot},
}

func writeHeaders(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, o := range options {
		svg := o.build()
		for _, th := range themes {
			path := filepath.Join(dir, fmt.Sprintf("%s-%s-%s.svg", o.key, o.name, th.name))
			if err := writeText(filepath.Dir(dir), path, th.t.paint(svg)); err != nil {
				return err
			}
		}
	}
	return nil
}

func frame(h int, body, defs, css, label string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1000 %d" width="1000" height="%d" role="img" aria-label="%s">
  <style>
    .mono { font-family: %s; }
    .sans { font-family: %s; }
    .fade { animation: fade 0.7s ease-out both; }
    .blink { animation: blink 1s steps(1) infinite; }
    @keyframes fade { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: none; } }
    @keyframes blink { 50%% { opacity: 0; } }
%s  </style>
  <defs>
    <pattern id="dots" width="22" height="22" patternUnits="userSpaceOnUse">
      <circle cx="1.5" cy="1.5" r="1.2" fill="{dot}"/>
    </pattern>
    <radialGradient id="glow" cx="0.5" cy="0.5" r="0.5">
      <stop offset="0" stop-color="{glow}" stop-opacity="0.22"/>
      <stop offset="1" stop-color="{glow}" stop-opacity="0"/>
    </radialGradient>
    <linearGradient id="name" x1="0" x2="1">
      <stop offset="0" stop-color="{text}"/>
      <stop offset="0.5" stop-color="{text}"/>
      <stop offset="1" stop-color="{accent}"/>
    </linearGradient>
%s  </defs>
  <rect x="0.75" y="0.75" width="998.5" height="%s" rx="14" fill="{bg}" stroke="{border}" stroke-width="1.5"/>
  <rect x="1" y="1" width="998" height="%d" rx="14" fill="url(#dots)"/>
%s</svg>
`, h, h, esc(label), mono, sans, css, defs, n(float64(h)-1.5), h-2, body)
}

// typed is a mono line that types itself out, with a cursor that follows and keeps blinking.
func typed(x, y int, text string, size int, color string, begin, dur float64, cid string) (defs, body string) {
	ch := float64(size) * 0.6
	count := utf8.RuneCountInString(text)
	length := round1(float64(count) * ch)
	var clip, cur, kt []string
	for i := 0; i <= count; i++ {
		s := round1(float64(i) * ch)
		clip = append(clip, n(s))
		cur = append(cur, n(round1(float64(x)+s+2)))
		kt = append(kt, fmt.Sprintf("%.4f", float64(i)/float64(count)))
	}
	keys := strings.Join(kt, ";")
	fy, fs := float64(y), float64(size)
	defs = fmt.Sprintf(`    <clipPath id="%s">
      <rect x="%d" y="%s" height="%s" width="%s">
        <animate attributeName="width" values="%s" keyTimes="%s" calcMode="discrete" dur="%ss" begin="%ss" fill="freeze"/>
      </rect>
    </clipPath>
`, cid, x, n(fy-fs*1.2), n(fs*1.8), n(length+20), strings.Join(clip, ";"), keys, n(dur), n(begin))
	body = fmt.Sprintf(`  <g clip-path="url(#%s)">
    <text x="%d" y="%d" class="mono" font-size="%d" fill="%s" textLength="%s" lengthAdjust="spacing">%s</text>
  </g>
  <rect x="%s" y="%s" width="%s" height="%s" fill="%s" class="blink">
    <animate attributeName="x" values="%s" keyTimes="%s" calcMode="discrete" dur="%ss" begin="%ss" fill="freeze"/>
  </rect>
`, cid, x, y, size, color, n(length), esc(text),
		n(float64(x)+length+2), n(fy-fs*0.85), n(fs*0.5), n(fs*1.1), color, strings.Join(cur, ";"), keys, n(dur), n(begin))
	return defs, body
}

func pill(x, y, w int, text, class, style string) string {
	cy := y + 17
	return fmt.Sprintf(`  <g class="%s" style="%s">
    <rect x="%d" y="%d" width="%d" height="34" rx="17" fill="{okbg}" stroke="{ok}" stroke-opacity="0.45"/>
    <circle cx="%d" cy="%d" r="9" fill="{ok}" opacity="0.3">
      <animate attributeName="r" values="4;10;4" dur="2s" repeatCount="indefinite"/>
      <animate attributeName="opacity" values="0.5;0;0.5" dur="2s" repeatCount="indefinite"/>
    </circle>
    <circle cx="%d" cy="%d" r="4.5" fill="{ok}"/>
    <text x="%d" y="%d" class="mono" font-size="14" fill="{ok}">%s</text>
  </g>
`, class, style, x, y, w, x+20, cy, x+20, cy, x+36, cy+5, esc(text))
}

func prompt(x, y int, cmd string) string {
	return fmt.Sprintf(`  <text x="%d" y="%d" class="mono fade" font-size="16"><tspan fill="{ok}">➜</tspan> `+
		`<tspan fill="{accent}">~/wapenshaw</tspan> <tspan fill="{muted}">%s</tspan></text>`+"\n", x, y, esc(cmd))
}

// --- option A: autopilot -------------------------------------------------------------------------
// Shared animation helpers for the infrastructure and application lanes in autopilot.go.

const run18 = 18 // seconds per run; every option-A animation uses this clock

func keyTimes(pts ...float64) string {
	out := make([]string, len(pts))
	for i, p := range pts {
		out[i] = fmt.Sprintf("%.3f", p)
	}
	return strings.Join(out, ";")
}

// window is visible only between fractions a and b of the run.
func window(a, b float64) string {
	return fmt.Sprintf(`<animate attributeName="opacity" dur="%ds" repeatCount="indefinite" values="0;0;1;1;0;0" keyTimes="%s"/>`,
		run18, keyTimes(0, a, a+.002, b, b+.002, 1))
}

func after(f float64) string { return window(f, .995) }

func until(f float64) string {
	return fmt.Sprintf(`<animate attributeName="opacity" dur="%ds" repeatCount="indefinite" values="1;1;0;0;1" keyTimes="%s"/>`,
		run18, keyTimes(0, f, f+.002, .995, 1))
}

func packet(d string, a, b float64, color string) string {
	return fmt.Sprintf(`    <circle r="4" fill="%s" opacity="0">%s<animateMotion dur="%ds" repeatCount="indefinite" path="%s" keyPoints="0;0;1;1" keyTimes="%s" calcMode="linear"/></circle>`+"\n",
		color, window(a, b), run18, d, keyTimes(0, a, b, 1))
}

func check(x, y int, label string, f float64, size int) string {
	return fmt.Sprintf(`    <text x="%d" y="%d" class="mono" font-size="%d" fill="{muted}">○ %s%s</text>`+"\n"+
		`    <text x="%d" y="%d" class="mono" font-size="%d" opacity="0"><tspan fill="{ok}">✓</tspan> <tspan fill="{text}">%s</tspan>%s</text>`+"\n",
		x, y, size, esc(label), until(f), x, y, size, esc(label), after(f))
}

// --- option B: agent run -------------------------------------------------------------------------

func agent() string {
	type seg struct{ text, color string }
	log := []struct {
		segs []seg
		wait float64 // seconds before this line appears
	}{
		{[]seg{{"$ ", "muted"}, {`agent run --goal "patch aks node images"`, "text"}}, 0.4},
		{[]seg{{"◆ planning", "accent2"}, {" · runbook + cluster state", "muted"}}, 0.6},
		{[]seg{{"✓ ", "ok"}, {"draft      ", "text"}, {"PR #2148 · bump node image", "muted"}}, 0.6},
		{[]seg{{"✓ ", "ok"}, {"ci         ", "text"}, {"lint · test · plan ~1", "muted"}}, 0.6},
		{[]seg{{"✓ ", "ok"}, {"policy     ", "text"}, {"opa + vuln scan passed", "muted"}}, 0.6},
		{[]seg{{"◇ ", "typ"}, {"waiting    ", "text"}, {"human approval required", "typ"}}, 0.6},
		{[]seg{{"✓ ", "ok"}, {"approved   ", "text"}, {"by on-call · merged", "muted"}}, 2.2},
		{[]seg{{"✓ ", "ok"}, {"rollout    ", "text"}, {"1 node at a time · pdbs held", "muted"}}, 0.6},
		{[]seg{{"◆ done", "accent2"}, {" · slos green · summary posted", "muted"}}, 0.6},
	}
	var lines []string
	delay := 0.0
	for i, l := range log {
		delay += l.wait
		var spans strings.Builder
		for _, s := range l.segs {
			fmt.Fprintf(&spans, `<tspan fill="{%s}">%s</tspan>`, s.color, esc(s.text))
		}
		lines = append(lines, fmt.Sprintf(`  <text x="532" y="%d" class="mono fade" font-size="14" xml:space="preserve" style="animation-delay:%.2fs">%s</text>`,
			88+i*29, delay, spans.String()))
	}
	last := 88 + len(log)*29
	delay += 0.6
	lines = append(lines, fmt.Sprintf(`  <g class="fade" style="animation-delay:%.2fs"><text x="532" y="%d" class="mono" font-size="14" fill="{muted}">$</text><rect x="548" y="%d" width="8" height="16" fill="{accent2}" class="blink"/></g>`,
		delay, last, last-12))

	body := `  <path d="M1,44 V15 A14,14 0 0 1 15,1 H985 A14,14 0 0 1 999,15 V44 Z" fill="{node}"/>
  <line x1="1" y1="44" x2="999" y2="44" stroke="{border}" stroke-width="1.5"/>
  <circle cx="28" cy="22" r="6.5" fill="#ff5f57"/>
  <circle cx="50" cy="22" r="6.5" fill="#febc2e"/>
  <circle cx="72" cy="22" r="6.5" fill="#28c840"/>
  <text x="500" y="27" text-anchor="middle" class="mono" font-size="13" fill="{muted}">siddharth@wapenshaw: ~/agents</text>
  <ellipse cx="230" cy="210" rx="240" ry="150" fill="url(#glow)"/>

  <text x="46" y="138" class="sans fade" font-size="60" font-weight="700" letter-spacing="-1.5" fill="{text}">Siddharth</text>
  <text x="46" y="202" class="sans fade" font-size="60" font-weight="700" letter-spacing="-1.5" fill="url(#name)" style="animation-delay:0.15s">Abbineni</text>
  <text x="50" y="246" class="mono fade" font-size="15" fill="{accent2}" style="animation-delay:0.3s">` + role + `</text>
  <text x="50" y="272" class="mono fade" font-size="15" fill="{muted}" style="animation-delay:0.4s">+ ai automation · ` + years + ` in prod</text>
` + pill(50, 300, 178, "agents: online", "fade", "animation-delay:0.5s") + `
  <line x1="502" y1="70" x2="502" y2="352" stroke="{border}" stroke-width="1.5" stroke-dasharray="2 5"/>
` + strings.Join(lines, "\n") + "\n"
	return frame(380, body, "", "", "Siddharth Abbineni: an agent run that provisions, configures, deploys and verifies")
}

// --- option C: polyglot --------------------------------------------------------------------------

var snippets = []struct {
	file     string
	keywords []string
	code     []string
}{
	{"main.rs", []string{"use", "fn", "let", "mut"}, []string{
		"use wapenshaw::Engineer;",
		"",
		"fn main() {",
		`    let sid = Engineer::new("siddharth");`,
		"    sid.automate(Scope::Everything);",
		"}",
	}},
	{"main.go", []string{"package", "func", "import"}, []string{
		"package main",
		"",
		"func main() {",
		`    sid := engineer.New("siddharth")`,
		"    sid.Automate(ctx, engineer.Everything)",
		"}",
	}},
	{"agent.py", []string{"from", "import", "await", "async", "def"}, []string{
		"from wapenshaw import Engineer",
		"",
		`sid = Engineer("siddharth")`,
		`await sid.automate(scope="everything")`,
	}},
	{"Program.cs", []string{"using", "var", "new", "await"}, []string{
		"using Wapenshaw;",
		"",
		`var sid = new Engineer("siddharth");`,
		"await sid.AutomateAsync(Scope.Everything);",
	}},
	{"index.ts", []string{"import", "from", "const", "new", "await"}, []string{
		`import { Engineer } from "wapenshaw";`,
		"",
		`const sid = new Engineer("siddharth");`,
		`await sid.automate({ scope: "everything" });`,
	}},
}

var tokens = regexp.MustCompile(`"[^"]*"|\w+|\s+|.`)

func highlight(line string, keywords []string) string {
	var out strings.Builder
	toks := tokens.FindAllString(line, -1)
	for i, tok := range toks {
		next := ""
		if i+1 < len(toks) {
			next = toks[i+1]
		}
		first, _ := utf8.DecodeRuneInString(tok)
		color := "text"
		switch {
		case strings.HasPrefix(tok, `"`):
			color = "str"
		case contains(keywords, tok):
			color = "kw"
		case next == "(":
			color = "fn"
		case unicode.IsUpper(first):
			color = "typ"
		}
		fmt.Fprintf(&out, `<tspan fill="{%s}">%s</tspan>`, color, esc(tok))
	}
	return out.String()
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func polyglot() string {
	count := len(snippets)
	per := 2.6
	on := 100 / float64(count)
	css := fmt.Sprintf(`    .slide { opacity: 0; animation: show %ss infinite; }
    .s0 { opacity: 1; }
    @keyframes show { 0%% { opacity: 0; } 2%% { opacity: 1; } %.1f%% { opacity: 1; } %.1f%% { opacity: 0; } 100%% { opacity: 0; } }
`, n(float64(count)*per), on-2, on)
	cx, cy, cw, ch := 498, 40, 458, 262
	var tabs, slides []string
	tx := float64(cx + 14)
	for i, s := range snippets {
		delay := fmt.Sprintf("animation-delay:-%.1fs", float64((count-i)%count)*per)
		w := float64(utf8.RuneCountInString(s.file))*7.2 + 22
		tabs = append(tabs,
			fmt.Sprintf(`    <text x="%s" y="%d" class="mono" font-size="12" fill="{muted}">%s</text>`, n(tx+11), cy+25, s.file),
			fmt.Sprintf(`    <g class="slide s%d" style="%s"><rect x="%s" y="%d" width="%s" height="30" rx="6" fill="{bg}"/><text x="%s" y="%d" class="mono" font-size="12" fill="{text}">%s</text><rect x="%s" y="%d" width="%s" height="2" rx="1" fill="{accent}"/></g>`,
				i, delay, n(tx), cy+6, n(w), n(tx+11), cy+25, s.file, n(tx+6), cy+34, n(w-12)))
		tx += w + 4
		var rows []string
		for j, line := range s.code {
			y := cy + 80 + j*26
			rows = append(rows, fmt.Sprintf(`      <text x="%d" y="%d" class="mono" font-size="13" fill="{muted}" opacity="0.6" text-anchor="start">%d</text>`, cx+18, y, j+1))
			if line != "" {
				rows = append(rows, fmt.Sprintf(`      <text x="%d" y="%d" class="mono" font-size="14" xml:space="preserve">%s</text>`, cx+46, y, highlight(line, s.keywords)))
			}
		}
		slides = append(slides, fmt.Sprintf("    <g class=\"slide s%d\" style=\"%s\">\n%s\n    </g>", i, delay, strings.Join(rows, "\n")))
	}
	body := `  <ellipse cx="730" cy="170" rx="260" ry="170" fill="url(#glow)"/>
` + prompt(56, 84, "ls ./languages") + `  <text x="53" y="148" class="sans fade" font-size="56" font-weight="700" letter-spacing="-1.4" fill="{text}">Siddharth</text>
  <text x="53" y="206" class="sans fade" font-size="56" font-weight="700" letter-spacing="-1.4" fill="url(#name)" style="animation-delay:0.15s">Abbineni</text>
  <text x="56" y="248" class="mono fade" font-size="15" fill="{accent2}" style="animation-delay:0.3s">` + role + `</text>
  <text x="56" y="274" class="mono fade" font-size="15" fill="{muted}" style="animation-delay:0.4s">+ ai automation · ` + years + ` in prod</text>

` + fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" rx="10" fill="{node}" stroke="{border}" stroke-width="1.5"/>
  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="{border}"/>
`, cx, cy, cw, ch, cx, cy+42, cx+cw, cy+42) + strings.Join(tabs, "\n") + "\n" + strings.Join(slides, "\n") + "\n" +
		fmt.Sprintf(`  <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="{border}"/>
  <text x="%d" y="%d" class="mono" font-size="12" fill="{muted}"><tspan fill="{ok}">✓</tspan> builds  <tspan fill="{ok}">✓</tspan> tests pass  <tspan fill="{accent2}">◆</tspan> same engineer, five runtimes</text>
`, cx, cy+ch-34, cx+cw, cy+ch-34, cx+18, cy+ch-12)
	return frame(340, body, "", css, "Siddharth Abbineni: the same snippet in Rust, Go, Python, C# and TypeScript")
}
