package main

import (
	"fmt"
	"strings"
)

// The lanes run independently until deployment needs a ready target environment.
// Infra represents a plan/approve/provision cycle for each changed environment;
// the application lane promotes one image digest through those environments.
func autopilot() string {
	defs, roleLine := typed(56, 164, "code · cloud · ai automation · open source", 17, "{accent2}", 0.6, 2.4, "typing")
	defs += `    <marker id="arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto">
      <path d="M0,0 L8,4 L0,8 z" fill="{muted}"/>
    </marker>
    <marker id="arrow2" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto">
      <path d="M0,0 L8,4 L0,8 z" fill="{accent2}"/>
    </marker>
    <marker id="ready-arrow" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="7" markerHeight="7" orient="auto">
      <path d="M0,0 L8,4 L0,8 z" fill="{ok}"/>
    </marker>
`
	css := `    .wire { fill: none; stroke: {muted}; stroke-opacity: 0.65; stroke-width: 1.5; marker-end: url(#arrow); }
    .loop { fill: none; stroke: {accent2}; stroke-opacity: 0.7; stroke-width: 1.5; stroke-dasharray: 5 5; marker-end: url(#arrow2); }
    .dependency { fill: none; stroke: {ok}; stroke-width: 1.5; marker-end: url(#ready-arrow); }
    .box { fill: {bg}; stroke: {border}; stroke-width: 1.5; }
    .hot { fill: none; stroke: {accent}; stroke-width: 2; }
    .title { fill: {text}; font-size: 13px; font-weight: 700; }
`
	type step struct {
		title string
		rows  [3]string
		times [3]float64
		a, b  float64
	}
	infra := []step{
		{"IaC PR", [3]string{"modules / config", "lint / validate", "review + merge"}, [3]float64{.08, .12, .17}, .04, .18},
		{"change plan", [3]string{"target env", "diff + cost", "change scope"}, [3]float64{.20, .23, .26}, .19, .27},
		{"approve", [3]string{"policy checks", "AI risk summary", "human approval"}, [3]float64{.29, .31, .35}, .28, .36},
		{"provision", [3]string{"scoped identity", "serialize runs", "approved change"}, [3]float64{.38, .40, .44}, .37, .45},
		{"verify", [3]string{"health + access", "policy posture", "ready signal"}, [3]float64{.48, .51, .54}, .46, .55},
		{"env outputs", [3]string{"IDs + endpoints", "secret refs", "env contract"}, [3]float64{.57, .58, .60}, .56, .61},
	}
	app := []step{
		{"app PR / CI", [3]string{"code + tests", "lint / test", "security"}, [3]float64{.08, .12, .17}, .04, .18},
		{"review", [3]string{"AI review", "human approval", "merge"}, [3]float64{.21, .25, .28}, .19, .29},
		{"artifact", [3]string{"build + scan", "publish digest", "manifest PR"}, [3]float64{.33, .37, .40}, .30, .41},
		{"dev → stage", [3]string{"ready + PR gate", "deploy + sync", "smoke + e2e"}, [3]float64{.65, .69, .74}, .64, .75},
		{"production", [3]string{"ready + approve", "SLO gates", "abort on fail"}, [3]float64{.78, .83, .89}, .76, .91},
		{"observe", [3]string{"errors + p99", "infra drift", "AI triage"}, [3]float64{.81, .85, .94}, .78, .99},
	}
	var lanes strings.Builder
	for lane, steps := range [][]step{infra, app} {
		y := 330 + lane*210
		for i, s := range steps {
			x := 44 + i*158
			fmt.Fprintf(&lanes, `  <rect x="%d" y="%d" width="132" height="118" rx="8" class="box"/>
  <rect x="%d" y="%d" width="132" height="118" rx="8" class="hot" opacity="0">%s</rect>
  <text x="%d" y="%d" class="mono title">%s</text>
`, x, y, x, y, window(s.a, s.b), x+12, y+24, esc(s.title))
			for j, label := range s.rows {
				// Production uses a live traffic bar below its readiness gate.
				if lane == 1 && i == 4 && j > 0 {
					continue
				}
				lanes.WriteString(check(x+10, y+48+j*23, label, s.times[j], 10))
			}
			if i < len(steps)-1 {
				d := fmt.Sprintf("M%d,%d L%d,%d", x+132, y+60, x+158, y+60)
				fmt.Fprintf(&lanes, `  <path d="%s" class="wire"/>`+"\n", d)
				start, end := s.b, steps[i+1].a
				if lane == 1 && i == 4 {
					// Telemetry is active during the canary, not just after it.
					start, end = .78, .80
				}
				lanes.WriteString(packet(d, start, end, "{accent}"))
			}
		}
	}
	// The artifact is available while the changed infrastructure is still being verified.
	wait := fmt.Sprintf(`  <g opacity="0">%s
    <text x="360" y="676" class="mono" font-size="10" fill="{typ}">waiting for target env</text>
  </g>
`, window(.42, .63))
	// Each deployment checks the readiness contract for its own target environment.
	readyDev := "M900,448 L900,478 L584,478 L584,540"
	readyProd := "M742,478 L742,540"
	dependencies := fmt.Sprintf(`  <path d="%s" class="dependency"/>
  <path d="%s" class="dependency"/>
  <text x="44" y="475" class="mono" font-size="12" fill="{ok}">deploy gate: target environment ready</text>
  <text x="44" y="496" class="mono" font-size="10" fill="{muted}">reuse existing infra when dependencies are unchanged</text>
  <text x="602" y="469" class="mono" font-size="10" fill="{ok}">per-env readiness + outputs</text>
`, readyDev, readyProd) +
		packet(readyDev, .61, .64, "{ok}") +
		packet(readyProd, .73, .76, "{ok}")
	// Traffic advances only after the current canary's metric checks pass.
	canary := fmt.Sprintf(`  <rect x="686" y="602" width="112" height="6" rx="3" fill="{border}"/>
  <rect x="686" y="602" width="0" height="6" rx="3" fill="{ok}"><animate attributeName="width" dur="%ds" repeatCount="indefinite" calcMode="discrete" values="0;11.2;56;112;0" keyTimes="%s"/></rect>
  <text x="686" y="625" class="mono" font-size="10" fill="{muted}">10 → 50 → 100%%</text>
  <text x="686" y="646" class="mono" font-size="10" fill="{muted}">SLOs / auto-abort</text>
`, run18, keyTimes(0, .79, .84, .89, 1))
	dispatchInfra := "M44,272 L34,272 L34,390 L44,390"
	dispatchApp := "M44,272 L34,272 L34,600 L44,600"
	feedback := "M900,658 L900,716 L16,716 L16,258 L44,258"
	aiWires := fmt.Sprintf(`  <path d="%s" class="loop"/>
  <path d="%s" class="loop"/>
  <path d="%s" class="loop"/>
`, dispatchInfra, dispatchApp, feedback) +
		packet(dispatchInfra, .01, .04, "{accent2}") +
		packet(dispatchApp, .01, .04, "{accent2}") +
		packet(feedback, .95, .99, "{accent2}")
	body := `  <ellipse cx="500" cy="480" rx="480" ry="270" fill="url(#glow)"/>
` + prompt(56, 62, "whoami") + `  <text x="53" y="122" class="sans fade" font-size="54" font-weight="700" letter-spacing="-1.2" fill="url(#name)">Siddharth Abbineni</text>
` + roleLine + pill(788, 44, 168, "mode: building", "fade", "") + `
  <rect x="24" y="196" width="952" height="546" rx="12" fill="{node}" fill-opacity="0.75" stroke="{border}" stroke-width="1.5"/>
  <text x="44" y="220" class="mono" font-size="12" fill="{text}"><tspan fill="{accent2}">◆</tspan> how I think about automation</text>
  <text x="956" y="220" text-anchor="end" class="mono" font-size="12" fill="{muted}">ai coordinates · humans approve · pipelines enforce</text>
  <line x1="24" y1="232" x2="976" y2="232" stroke="{border}"/>

  <rect x="44" y="242" width="912" height="50" rx="8" fill="{bg}" stroke="{accent2}" stroke-opacity="0.5"/>
  <text x="60" y="263" class="mono title" style="fill:{accent2}">◆ ai automation</text>
  <text x="224" y="262" class="mono" font-size="11" fill="{text}">map dependencies · draft IaC / code / tests · explain plans · coordinate runs</text>
  <text x="224" y="281" class="mono" font-size="10" fill="{muted}">scoped tools + runbooks · investigate failures · propose fixes · summarize evidence</text>

  <text x="44" y="318" class="mono" font-size="12" fill="{accent}">01 infrastructure · when infra changes</text>
  <text x="956" y="318" text-anchor="end" class="mono" font-size="10" fill="{muted}">fresh plan + approval per env: dev → stage → prod</text>
  <text x="44" y="526" class="mono" font-size="12" fill="{accent}">02 application · build in parallel</text>
  <text x="956" y="676" text-anchor="end" class="mono" font-size="10" fill="{muted}">one digest · GitOps promotion</text>
` + aiWires + lanes.String() + dependencies + wait + canary + `
  <text x="44" y="696" class="mono" font-size="11" fill="{accent2}">ops → AI correlates alerts + drift → app / IaC fix PRs → same checks + approvals</text>
  <rect x="260" y="708" width="464" height="17" fill="{node}"/>
  <text x="492" y="720" text-anchor="middle" class="mono" font-size="10" fill="{muted}">approved runbooks automate routine ops · exceptions escalate</text>
`
	return frame(762, body, defs, css,
		"Siddharth Abbineni: code, cloud, AI automation and open source. An illustration of how I think about automation: infrastructure and application delivery connect through environment readiness, with AI assistance, human review, verification and feedback.")
}
