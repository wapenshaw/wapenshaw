package main

import (
	"strconv"
	"strings"
)

// Theme maps color names to values. SVG templates refer to them as {name} and paint resolves them.
type Theme map[string]string

var themes = []struct {
	name string
	t    Theme
}{
	{"dark", Theme{
		"bg": "#0d1117", "border": "#30363d", "dot": "#21262d", "text": "#e6edf3", "muted": "#7d8590",
		"accent": "#58a6ff", "accent2": "#d2a8ff", "ok": "#3fb950", "okbg": "#12261e", "node": "#161b22", "glow": "#1f6feb",
		"kw": "#ff7b72", "str": "#a5d6ff", "fn": "#d2a8ff", "typ": "#ffa657",
	}},
	{"light", Theme{
		"bg": "#ffffff", "border": "#d0d7de", "dot": "#e6e9ed", "text": "#1f2328", "muted": "#59636e",
		"accent": "#0969da", "accent2": "#8250df", "ok": "#1a7f37", "okbg": "#dafbe1", "node": "#f6f8fa", "glow": "#54aeff",
		"kw": "#cf222e", "str": "#0a3069", "fn": "#8250df", "typ": "#953800",
	}},
}

func (t Theme) paint(svg string) string {
	pairs := make([]string, 0, 2*len(t))
	for k, v := range t {
		pairs = append(pairs, "{"+k+"}", v)
	}
	return strings.NewReplacer(pairs...).Replace(svg)
}

const (
	mono = `ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace`
	sans = `-apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif`
)

var esc = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace

// n formats a coordinate with the fewest digits that round-trip.
func n(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// round1 rounds to one decimal place on the exact binary value.
func round1(v float64) float64 {
	r, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'f', 1, 64), 64)
	return r
}
