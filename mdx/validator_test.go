package mdx

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testContentHTML = "<p>content</p>"
	testTitle       = "Test"
	testPath        = "book/test.mdx"
	testOneHTML     = "<h1>One</h1>"
	testThreeHTML   = "<h1>One</h1><h2>Two</h2><h3>Three</h3>"
	testOneTwoHTML  = "<h1>One</h1><h2>Two</h2>"
	testNoHeadings  = "<p>no headings</p>"
)

func TestDefaultValidatorValidate(t *testing.T) {
	v := NewDefaultValidator()

	t.Run("valid document", func(t *testing.T) {
		doc := &Document{
			Path: testPath,
			Frontmatter: map[string]interface{}{
				"id":              defaultIDValue,
				defaultTitleField: "Test Chapter",
			},
			HTML: "<h1>Title</h1>" + testContentHTML,
		}
		errs := v.Validate(doc)
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
	})

	t.Run("missing id and title", func(t *testing.T) {
		doc := &Document{
			Path:        testPath,
			Frontmatter: map[string]interface{}{},
			HTML:        testContentHTML,
		}
		errs := v.Validate(doc)
		if len(errs) < 2 {
			t.Fatalf("expected at least 2 errors, got %d", len(errs))
		}
	})

	t.Run("invalid id format", func(t *testing.T) {
		doc := &Document{
			Path: testPath,
			Frontmatter: map[string]interface{}{
				"id":              "not-valid",
				defaultTitleField: testTitle,
			},
			HTML: testContentHTML,
		}
		errs := v.Validate(doc)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("max heading depth exceeded", func(t *testing.T) {
		v2 := NewDefaultValidator()
		v2.MaxHeadingDepth = 2
		doc := &Document{
			Path: testPath,
			Frontmatter: map[string]interface{}{
				"id":              "",
				defaultTitleField: testTitle,
			},
			HTML: testContentHTML,
		}
		errs := v2.Validate(doc)
		if len(errs) == 0 {
			t.Fatal("expected heading depth error")
		}
	})

	t.Run("heading depth within limit", func(t *testing.T) {
		v3 := NewDefaultValidator()
		v3.MaxHeadingDepth = 3
		doc := &Document{
			Path: testPath,
			Frontmatter: map[string]interface{}{
				"id":              defaultIDValue,
				defaultTitleField: testTitle,
			},
			HTML: testThreeHTML,
		}
		errs := v3.Validate(doc)
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d", len(errs))
		}
	})

	t.Run("empty id field", func(t *testing.T) {
		doc := &Document{
			Path: testPath,
			Frontmatter: map[string]interface{}{
				"id":              "",
				defaultTitleField: testTitle,
			},
			HTML: testThreeHTML,
		}
		errs := v.Validate(doc)
		if len(errs) != 2 {
			t.Fatalf("expected 2 errors for empty id (frontmatter + format), got %d: %v", len(errs), errs)
		}
	})
}

func TestDefaultValidatorValidateAll(t *testing.T) {
	v := NewDefaultValidator()

	docs := []*Document{
		{
			Path: "book/ch1.mdx",
			Frontmatter: map[string]interface{}{
				"id":              defaultIDValue,
				defaultTitleField: "Chapter 1",
			},
			HTML: testOneHTML,
		},
		{
			Path: "book/ch2.mdx",
			Frontmatter: map[string]interface{}{
				"id":              "[2.0.0]",
				defaultTitleField: "Chapter 2",
			},
			HTML: testOneHTML,
		},
	}

	errs := v.ValidateAll(docs)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}
}

func TestDefaultValidatorDuplicateIDs(t *testing.T) {
	v := NewDefaultValidator()

	docs := []*Document{
		{
			Path: "book/ch1.mdx",
			Frontmatter: map[string]interface{}{
				"id":              defaultIDValue,
				defaultTitleField: "Chapter 1",
			},
			HTML: testOneHTML,
		},
		{
			Path: "book/ch2.mdx",
			Frontmatter: map[string]interface{}{
				"id":              defaultIDValue,
				defaultTitleField: "Duplicate Chapter",
			},
			HTML: testOneHTML,
		},
	}

	errs := v.ValidateAll(docs)
	hasDup := false
	for _, e := range errs {
		if e.Field == "id" && containsStr(e.Message, "duplicate") {
			hasDup = true
			break
		}
	}
	if !hasDup {
		t.Fatal("expected duplicate ID error")
	}
}

func TestCountMaxHeadingDepth(t *testing.T) {
	tests := []struct {
		html   string
		expect int
	}{
		{testOneTwoHTML, 2},
		{"<h1>One</h1><h2>Two</h2><h3>Three</h3><h4>Four</h4>", 4},
		{testNoHeadings, 0},
		{testOneHTML, 1},
	}
	for _, tt := range tests {
		got := countMaxHeadingDepth(tt.html)
		if got != tt.expect {
			t.Errorf("countMaxHeadingDepth(%q) = %d, want %d", tt.html, got, tt.expect)
		}
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestVariableSubstitution(t *testing.T) {
	dir := t.TempDir()

	content := `---
id: "[1.0.0]"
title: "{{book_title}}"
subtitle: ""
tags: [test]
difficulty: Beginner
status: Draft
completeness: 0
depends_on: []
---

# {{book_title}}

Welcome to {{company_name}}. The API is at version {{api_version}}.
`
	path := filepath.Join(dir, "test.mdx")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewParser(WithVars(map[string]string{
		"book_title":   "My Awesome Book",
		"company_name": "Acme Corp",
		"api_version":  "v2.1",
	}))

	doc, err := p.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if doc.Title() != "My Awesome Book" {
		t.Errorf("expected title 'My Awesome Book', got %q", doc.Title())
	}
	if !contains(doc.HTML, "My Awesome Book") {
		t.Error("expected var substitution in body")
	}
	if !contains(doc.HTML, "Acme Corp") {
		t.Error("expected company_name substitution in body")
	}
	if !contains(doc.HTML, "v2.1") {
		t.Error("expected api_version substitution in body")
	}
}

func TestParserWithoutVars(t *testing.T) {
	dir := t.TempDir()

	content := `---
id: "[1.0.0]"
title: "My Book"
subtitle: ""
tags: [test]
difficulty: Beginner
status: Draft
completeness: 0
depends_on: []
---

# My Book

The text {{unset_var}} should remain verbatim.
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

	if !contains(doc.HTML, "{{unset_var}}") {
		t.Error("expected unset var to remain verbatim")
	}
}

func TestRegisterComponent(t *testing.T) {
	p := NewParser()

	called := false
	p.RegisterComponent("Test", func(attrs map[string]string, inner string) string {
		called = true
		return "<div class=\"test\">" + inner + "</div>"
	})

	dir := t.TempDir()
	content := `---
id: "[1.0.0]"
title: "Test"
subtitle: ""
tags: []
difficulty: Beginner
status: Draft
completeness: 0
depends_on: []
---

<Test>
Hello
</Test>
`
	path := filepath.Join(dir, "test.mdx")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := p.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Error("expected custom component handler to be called")
	}
}
