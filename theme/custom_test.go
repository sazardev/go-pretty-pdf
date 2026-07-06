package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeThemeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadCustomThemeAndResolve(t *testing.T) {
	dir := t.TempDir()
	path := writeThemeFile(t, dir, "mine.theme.yml", `
name: mine
description: "My theme"
extends: classic
colors:
  primary: "#123456"
sections:
  cover: false
css: |
  .marker { color: red; }
`)

	ct, err := LoadCustomTheme(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct.Extends != "classic" {
		t.Errorf("expected extends 'classic', got %q", ct.Extends)
	}

	css, sections, err := ct.Resolve(Options{})
	if err != nil {
		t.Fatalf("unexpected error resolving: %v", err)
	}
	if !strings.Contains(css, "--pdf-primary: #123456;") {
		t.Error("expected the custom theme's own color to be applied")
	}
	if !strings.Contains(css, ".marker { color: red; }") {
		t.Error("expected raw css to be appended")
	}
	if sections.Cover {
		t.Error("expected cover to be disabled per the custom theme's sections block")
	}
	if !strings.Contains(css, "classic") {
		t.Error("expected the extended base theme's CSS to be included")
	}
}

func TestCustomThemeCLIOptionsOverrideYAMLDefaults(t *testing.T) {
	ct := &CustomTheme{
		Name:    "mine",
		Extends: "default",
		Colors:  Colors{Primary: "#111111"},
		Sections: Sections{
			Cover: BoolPtr(false),
		},
	}

	// CLI-level opts should win over the custom theme's own YAML defaults.
	css, sections, err := ct.Resolve(Options{
		Colors:   Colors{Primary: "#222222"},
		Sections: Sections{Cover: BoolPtr(true)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(css, "--pdf-primary: #222222;") {
		t.Error("expected opts.Colors.Primary to override the custom theme's own color")
	}
	if !sections.Cover {
		t.Error("expected opts.Sections.Cover to override the custom theme's own sections block")
	}
}

func TestCustomThemeUnknownExtends(t *testing.T) {
	ct := &CustomTheme{Name: "mine", Extends: "does-not-exist"}
	if _, _, err := ct.Resolve(Options{}); err == nil {
		t.Error("expected an error for an unknown extends target")
	}
}

func TestLoadCustomThemeMissingFile(t *testing.T) {
	if _, err := LoadCustomTheme("/nonexistent/path.theme.yml"); err == nil {
		t.Error("expected an error for a missing theme file")
	}
}
