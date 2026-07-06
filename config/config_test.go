package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Source != defaultSource {
		t.Errorf("expected source 'book', got %q", cfg.Source)
	}
	if cfg.Output != defaultOutput {
		t.Errorf("expected output 'out.pdf', got %q", cfg.Output)
	}
	if cfg.Title != "Document" {
		t.Errorf("expected title 'Document', got %q", cfg.Title)
	}
	if len(cfg.Lint.RequireFrontmatter) != 2 {
		t.Errorf("expected 2 required frontmatter fields, got %d", len(cfg.Lint.RequireFrontmatter))
	}
	if !cfg.Lint.NoDuplicateIDs {
		t.Error("expected NoDuplicateIDs to be true")
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go-pretty-pdf.yml")
	content := `title: "My Book"
subtitle: "A Guide"
author: "Jane Doe"
source: docs
output: mybook.pdf
theme: minimal
vars:
  api_version: "v2.1"
  company: "Acme Corp"
lint:
  require_frontmatter: [id, title, subtitle]
  require_id_format: "[X.Y.Z]"
  no_duplicate_ids: true
  max_heading_depth: 2
render:
  timeout: 30s
  paper: letter
  margin_top: 10mm
  margin_bottom: 10mm
  header_title: "{{title}}"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Title != "My Book" {
		t.Errorf("expected title 'My Book', got %q", cfg.Title)
	}
	if cfg.Subtitle != "A Guide" {
		t.Errorf("expected subtitle 'A Guide', got %q", cfg.Subtitle)
	}
	if cfg.Author != "Jane Doe" {
		t.Errorf("expected author 'Jane Doe', got %q", cfg.Author)
	}
	if cfg.Source != "docs" {
		t.Errorf("expected source 'docs', got %q", cfg.Source)
	}
	if cfg.Output != "mybook.pdf" {
		t.Errorf("expected output 'mybook.pdf', got %q", cfg.Output)
	}
	if cfg.Theme != "minimal" {
		t.Errorf("expected theme 'minimal', got %q", cfg.Theme)
	}
	if cfg.Vars["api_version"] != "v2.1" {
		t.Errorf("expected var api_version=v2.1, got %q", cfg.Vars["api_version"])
	}
	if cfg.Vars["company"] != "Acme Corp" {
		t.Errorf("expected var company='Acme Corp', got %q", cfg.Vars["company"])
	}
	if len(cfg.Lint.RequireFrontmatter) != 3 {
		t.Errorf("expected 3 required frontmatter fields, got %d", len(cfg.Lint.RequireFrontmatter))
	}
	if cfg.Lint.MaxHeadingDepth != 2 {
		t.Errorf("expected max heading depth 2, got %d", cfg.Lint.MaxHeadingDepth)
	}
	if cfg.Render.Timeout != "30s" {
		t.Errorf("expected timeout '30s', got %q", cfg.Render.Timeout)
	}
	if cfg.Render.Paper != PaperLetter {
		t.Errorf("expected paper 'letter', got %q", cfg.Render.Paper)
	}
	if cfg.Render.MarginTop != "10mm" {
		t.Errorf("expected margin_top '10mm', got %q", cfg.Render.MarginTop)
	}
}

func TestLoadDefaultsOnMissingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go-pretty-pdf.yml")
	content := `title: "Just Title"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Title != "Just Title" {
		t.Errorf("expected title 'Just Title', got %q", cfg.Title)
	}
	if cfg.Source != defaultSource {
		t.Errorf("expected default source 'book', got %q", cfg.Source)
	}
	if cfg.Output != defaultOutput {
		t.Errorf("expected default output 'out.pdf', got %q", cfg.Output)
	}
}

func TestFindConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go-pretty-pdf.yml")
	if err := os.WriteFile(path, []byte("title: test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	oldDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatal(err)
		}
	}()

	found, err := FindConfig()
	if err != nil {
		t.Fatal(err)
	}
	if found == "" {
		t.Error("expected to find config file")
	}
}
