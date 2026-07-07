package main

import (
	"strings"
	"testing"

	"github.com/sazardev/go-pretty-pdf/theme"
)

func TestExtractThemeVars(t *testing.T) {
	css := `:root {
  --pdf-primary: #1c1c1c;
  --pdf-accent: #7a4a2b;
  --pdf-font-heading: 'Georgia', 'Iowan Old Style', serif;
}
.cover h1 { border-bottom: 2px solid var(--pdf-accent, #7a4a2b); }
`
	got := extractThemeVars(css)

	want := map[string]string{
		varPrimary:     "#1c1c1c",
		varAccent:      "#7a4a2b",
		varFontHeading: "'Georgia', 'Iowan Old Style', serif",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("extractThemeVars()[%q] = %q, want %q", k, got[k], v)
		}
	}
	// A var(--pdf-accent, ...) *usage* must never be mistaken for a
	// declaration — it has no trailing "-usage" key and must not silently
	// override the real "accent" declaration above.
	if len(got) != len(want) {
		t.Errorf("extractThemeVars() found %d vars, want exactly %d: %v", len(got), len(want), got)
	}
}

// TestBuiltinThemesProduceSiteVars is the regression test for the exact
// bug this file's approach replaced: every builtin theme's CSS must yield
// a usable --site-primary/--site-accent/--site-bg/--site-font-body, or the
// docs site would silently fall back to no color/font at all for it.
func TestBuiltinThemesProduceSiteVars(t *testing.T) {
	required := []string{varPrimary, varAccent, varText, varMuted, varBg, varFontHeading, varFontBody, varFontCode}

	for _, th := range theme.List() {
		vars := extractThemeVars(th.CSS)
		for _, name := range required {
			if vars[name] == "" {
				t.Errorf("theme %q: missing --pdf-%s (site would render with no fallback)", th.Name, name)
			}
		}
	}
}

func TestThemeCSSBlock(t *testing.T) {
	classic, ok := theme.Get(theme.NameClassic)
	if !ok {
		t.Fatal("theme.Get(classic) not found")
	}
	block := themeCSSBlock(classic)

	if !strings.HasPrefix(block, ":root,\n") {
		t.Error("classic block should double as the :root fallback")
	}
	if !strings.Contains(block, `--site-accent: #7a4a2b;`) {
		t.Errorf("classic block missing expected accent value: %s", block)
	}
	if !strings.Contains(block, "--site-flourish: var(--site-accent);") {
		t.Error("classic is Accented, flourish should use the accent color")
	}

	def, _ := theme.Get(theme.NameDefault)
	defBlock := themeCSSBlock(def)
	if strings.HasPrefix(defBlock, ":root") {
		t.Error("only classic should double as the :root fallback")
	}
	if !strings.Contains(defBlock, "--site-flourish: var(--site-text);") {
		t.Error("default is not Accented, flourish should use the neutral text color")
	}
}

func TestDisplayName(t *testing.T) {
	tests := map[string]string{
		"default":   "Default",
		"corporate": "Corporate",
		"":          "",
	}
	for in, want := range tests {
		if got := displayName(in); got != want {
			t.Errorf("displayName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSwatchGradientFallsBackOnMissingColors(t *testing.T) {
	g := swatchGradient(theme.Theme{CSS: "/* no vars */"})
	if !strings.Contains(g, "#ffffff") || !strings.Contains(g, "#888888") {
		t.Errorf("swatchGradient() with no colors = %q, want fallback bg/accent", g)
	}
}
