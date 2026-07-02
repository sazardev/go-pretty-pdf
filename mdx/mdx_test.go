package mdx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComponentTranspileDeepDive(t *testing.T) {
	r := NewComponentRegistry()
	r.Register("DeepDive", deepDiveHandler)

	input := `<DeepDive title="Test Title">
This is **deep** content with ` + "`code`" + ` here.
</DeepDive>`

	result := r.Transpile(input)

	if !contains(result, `class="component-deep-dive"`) {
		t.Error("expected component-deep-dive class")
	}
	if !contains(result, `class="component-deep-dive-title"`) {
		t.Error("expected component-deep-dive-title class")
	}
	if !contains(result, "Test Title") {
		t.Error("expected title")
	}
	if !contains(result, "<strong>deep</strong>") {
		t.Error("expected strong tag for bold")
	}
	if !contains(result, "<code>code</code>") {
		t.Error("expected code tag")
	}
	if contains(result, "DeepDive") {
		t.Error("expected no raw component tags")
	}
}

func TestComponentTranspileWarning(t *testing.T) {
	r := NewComponentRegistry()

	input := `<Warning title="Heads Up">
Be careful with this.
</Warning>`

	result := r.Transpile(input)

	if !contains(result, `class="component-warning"`) {
		t.Error("expected component-warning class")
	}
	if !contains(result, "Heads Up") {
		t.Error("expected title")
	}
	if !contains(result, "Be careful") {
		t.Error("expected content")
	}
}

func TestComponentTranspileAxiom(t *testing.T) {
	r := NewComponentRegistry()

	input := `<Axiom>
Simplicity is prerequisite for reliability.
</Axiom>`

	result := r.Transpile(input)

	if !contains(result, `class="component-axiom"`) {
		t.Error("expected component-axiom class")
	}
	if !contains(result, "Simplicity is prerequisite for reliability.") {
		t.Error("expected content")
	}
}

func TestComponentTranspileMultipleComponents(t *testing.T) {
	r := NewComponentRegistry()

	input := `<DeepDive title="One">
First component.
</DeepDive>

<Warning title="Two">
Second component.
</Warning>`

	result := r.Transpile(input)

	if !contains(result, "First component") {
		t.Error("expected first component content")
	}
	if !contains(result, "Second component") {
		t.Error("expected second component content")
	}
}

func TestComponentTranspileWithoutTitle(t *testing.T) {
	r := NewComponentRegistry()

	input := `<Axiom>
No title here.
</Axiom>`

	result := r.Transpile(input)

	if !contains(result, `class="component-axiom"`) {
		t.Error("expected component-axiom class")
	}
	if contains(result, `class="component-deep-dive-title"`) {
		t.Error("unexpected title div")
	}
	if !contains(result, "No title here") {
		t.Error("expected content")
	}
}

func TestCustomComponent(t *testing.T) {
	r := NewComponentRegistry()
	r.Register("Callout", func(attrs map[string]string, inner string) string {
		return `<div class="callout callout-` + attrs["title"] + `">` + inner + `</div>`
	})

	input := `<Callout title="info">
Custom component content.
</Callout>`

	result := r.Transpile(input)

	if !contains(result, `class="callout callout-info"`) {
		t.Error("expected custom component class")
	}
	if !contains(result, "Custom component content") {
		t.Error("expected custom component content")
	}
}

func TestParserParseFile(t *testing.T) {
	dir := t.TempDir()

	content := `---
id: "[1.0.0]"
title: Test Chapter
subtitle: A test subtitle
tags: [test, example]
difficulty: Beginner
status: Draft
completeness: 50
depends_on: []
---

# Test Chapter

This is a test paragraph.

## Section One

More content here.

<DeepDive title="Test Deep Dive">
This is **bold** and ` + "`code`" + `.
</DeepDive>

| Col A | Col B |
|-------|-------|
| Val 1 | Val 2 |
`
	path := filepath.Join(dir, "test.mdx")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewParser()
	doc, err := p.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if doc.ID() != "[1.0.0]" {
		t.Errorf("expected ID [1.0.0], got %s", doc.ID())
	}
	if doc.Title() != "Test Chapter" {
		t.Errorf("expected title 'Test Chapter', got %s", doc.Title())
	}
	if len(doc.Tags()) != 2 {
		t.Errorf("expected 2 tags, got %d", len(doc.Tags()))
	}
	if doc.Difficulty() != "Beginner" {
		t.Errorf("expected Beginner, got %s", doc.Difficulty())
	}
	if doc.Status() != "Draft" {
		t.Errorf("expected Draft, got %s", doc.Status())
	}
	if doc.Completeness() != 50 {
		t.Errorf("expected completeness 50, got %d", doc.Completeness())
	}
	if len(doc.DependsOn()) != 0 {
		t.Errorf("expected no dependencies, got %v", doc.DependsOn())
	}

	if !contains(doc.HTML, "Test Chapter") {
		t.Error("expected h1 with title in HTML")
	}
	if !contains(doc.HTML, `class="component-deep-dive"`) {
		t.Error("expected transpiled component in HTML")
	}
	if !contains(doc.HTML, "<table>") {
		t.Error("expected table in HTML")
	}
}

func TestParserParseDir(t *testing.T) {
	dir := t.TempDir()

	files := []struct {
		name    string
		content string
	}{
		{"02-section.mdx", `---
id: "[2.0.0]"
title: Second Section
subtitle: ""
tags: [go]
difficulty: Advanced
status: Draft
completeness: 0
depends_on: []
---

# Second Section
`},
		{"01-intro.mdx", `---
id: "[1.0.0]"
title: Introduction
subtitle: ""
tags: [intro]
difficulty: Beginner
status: Draft
completeness: 0
depends_on: []
---

# Introduction
`},
		{"03-advanced.mdx", `---
id: "[1.1.0]"
title: Advanced Topics
subtitle: ""
tags: [advanced]
difficulty: Advanced
status: Draft
completeness: 0
depends_on: ["[1.0.0]"]
---

# Advanced Topics
`},
	}

	for _, f := range files {
		path := filepath.Join(dir, f.name)
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	p := NewParser()
	docs, err := p.ParseDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(docs) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(docs))
	}

	if docs[0].ID() != "[1.0.0]" {
		t.Errorf("expected first doc [1.0.0], got %s", docs[0].ID())
	}
	if docs[1].ID() != "[1.1.0]" {
		t.Errorf("expected second doc [1.1.0], got %s", docs[1].ID())
	}
	if docs[2].ID() != "[2.0.0]" {
		t.Errorf("expected third doc [2.0.0], got %s", docs[2].ID())
	}
}

func TestDocumentSortKey(t *testing.T) {
	tests := []struct {
		id     string
		expect [3]int
	}{
		{"[1.0.0]", [3]int{1, 0, 0}},
		{"[3.2.1]", [3]int{3, 2, 1}},
		{"[10.20.30]", [3]int{10, 20, 30}},
		{"invalid", [3]int{0, 0, 0}},
	}

	for _, tt := range tests {
		d := &Document{Frontmatter: map[string]interface{}{"id": tt.id}}
		got := d.SortKey()
		if got != tt.expect {
			t.Errorf("SortKey(%q) = %v, want %v", tt.id, got, tt.expect)
		}
	}
}

func TestAnchorID(t *testing.T) {
	tests := []struct {
		id     string
		expect string
	}{
		{"[1.0.0]", "section-1.0.0"},
		{"[10.20.30]", "section-10.20.30"},
		{"[0.0.0]", "section-0.0.0"},
		{"1.0.0", "section-1.0.0"},
		{"", "section-"},
	}
	for _, tt := range tests {
		got := AnchorID(tt.id)
		if got != tt.expect {
			t.Errorf("AnchorID(%q) = %q, want %q", tt.id, got, tt.expect)
		}
	}
}

func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
