package theme

import "testing"

func TestExtractCSSVars(t *testing.T) {
	css := `:root {
  --pdf-bg: #282828;
  --pdf-muted: #a89984;
}
.cover h1 { border-bottom: 2px solid var(--pdf-accent, #7a4a2b); }
:root { --pdf-bg: #ffffff; }
`
	got := ExtractCSSVars(css)

	if got["muted"] != "#a89984" {
		t.Errorf(`ExtractCSSVars()["muted"] = %q, want "#a89984"`, got["muted"])
	}
	// The second declaration of --pdf-bg must win, matching CSS cascade
	// order — this matters for Resolve()'s appended color-override block.
	if got["bg"] != "#ffffff" {
		t.Errorf(`ExtractCSSVars()["bg"] = %q, want "#ffffff" (last declaration should win)`, got["bg"])
	}
	// var(--pdf-accent, ...) is a usage, not a declaration — must not leak in.
	if _, ok := got["accent"]; ok {
		t.Errorf(`ExtractCSSVars() should not treat var(--pdf-accent, ...) usage as a declaration, got %q`, got["accent"])
	}
}
