package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindCustomAndListCustom(t *testing.T) {
	cwd := t.TempDir()
	themesDir := ProjectThemesDir(cwd)
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeThemeFile(t, themesDir, "mine.theme.yml", "name: mine\nextends: default\n")
	writeThemeFile(t, themesDir, "other.theme.yml", "name: other\nextends: minimal\n")
	// Not a theme file — must be ignored.
	writeThemeFile(t, themesDir, "notes.txt", "irrelevant")

	if _, ok := FindCustom(cwd, "does-not-exist"); ok {
		t.Error("expected FindCustom to report false for a missing theme")
	}
	path, ok := FindCustom(cwd, "mine")
	if !ok {
		t.Fatal("expected FindCustom to locate 'mine'")
	}
	if filepath.Base(path) != "mine.theme.yml" {
		t.Errorf("expected path ending in mine.theme.yml, got %s", path)
	}

	list, err := ListCustom(cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 custom themes, got %d: %+v", len(list), list)
	}
	if list[0].Name != "mine" || list[1].Name != "other" {
		t.Errorf("expected sorted names [mine, other], got [%s, %s]", list[0].Name, list[1].Name)
	}
	for _, c := range list {
		if c.Global {
			t.Errorf("expected project-local theme %q to not be marked global", c.Name)
		}
	}
}

func TestListCustomEmptyDirDoesNotError(t *testing.T) {
	cwd := t.TempDir()
	list, err := ListCustom(cwd)
	if err != nil {
		t.Fatalf("unexpected error for a project with no themes/ dir: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected no custom themes, got %+v", list)
	}
}

func TestScaffoldYAMLIsValidAndLoadable(t *testing.T) {
	base, _ := Get("classic")
	yamlContent := ScaffoldYAML("my-theme", base)
	if !strings.Contains(yamlContent, "extends: classic") {
		t.Error("expected scaffold to extend the given base theme")
	}

	dir := t.TempDir()
	path := writeThemeFile(t, dir, "my-theme.theme.yml", yamlContent)
	ct, err := LoadCustomTheme(path)
	if err != nil {
		t.Fatalf("expected scaffolded YAML to parse cleanly, got: %v", err)
	}
	if ct.Name != "my-theme" || ct.Extends != "classic" {
		t.Errorf("unexpected parsed scaffold: %+v", ct)
	}
	if _, _, err := ct.Resolve(Options{}); err != nil {
		t.Fatalf("expected scaffolded theme to resolve cleanly, got: %v", err)
	}
}

func TestResolveByNameBuiltinCustomAndUnknown(t *testing.T) {
	cwd := t.TempDir()
	themesDir := ProjectThemesDir(cwd)
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeThemeFile(t, themesDir, "mine.theme.yml", "name: mine\nextends: dark\n")

	if _, _, err := ResolveByName("modern", Options{}, cwd); err != nil {
		t.Errorf("expected builtin theme to resolve: %v", err)
	}
	if css, _, err := ResolveByName("mine", Options{}, cwd); err != nil {
		t.Errorf("expected custom theme to resolve: %v", err)
	} else if !strings.Contains(css, "dark") {
		t.Error("expected resolved CSS to include the extended (dark) theme's CSS")
	}
	if _, _, err := ResolveByName("totally-unknown", Options{}, cwd); err == nil {
		t.Error("expected an error for an unknown theme name")
	}
	if _, _, err := ResolveByName("", Options{}, cwd); err != nil {
		t.Errorf("expected empty name to fall back to the default theme, got error: %v", err)
	}
}

func TestResolveByNameDirectCSSPath(t *testing.T) {
	dir := t.TempDir()
	path := writeThemeFile(t, dir, "raw.css", "body { color: purple; }")

	css, sections, err := ResolveByName(path, Options{}, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(css, "body { color: purple; }") {
		t.Error("expected the raw CSS file's content to be included")
	}
	if !sections.Cover || !sections.TOC {
		t.Error("expected default sections to be on for a raw CSS file")
	}
}

func TestResolveByNameDirectThemeYAMLPath(t *testing.T) {
	dir := t.TempDir()
	path := writeThemeFile(t, dir, "custom.theme.yml", "name: custom\nextends: academic\n")

	css, _, err := ResolveByName(path, Options{}, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(css, "academic") {
		t.Error("expected resolved CSS to include the extended (academic) theme's CSS")
	}
}
