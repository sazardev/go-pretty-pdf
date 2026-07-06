package compose

import (
	"strings"
	"testing"

	"github.com/sazardev/go-pretty-pdf/mdx"
)

func docWithID(id, title, html string) *mdx.Document {
	return &mdx.Document{
		Frontmatter: map[string]interface{}{
			"id":    id,
			"title": title,
		},
		HTML: html,
	}
}

func TestComposeHTMLDefault(t *testing.T) {
	docs := []*mdx.Document{
		docWithID("[1.0.0]", "Chapter One", "<h1>Chapter One</h1><p>Body</p>"),
	}
	opts := DefaultOptions()
	opts.Title = "My Book"

	html, err := ComposeHTML(docs, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "My Book") {
		t.Error("expected title in composed HTML")
	}
	if !strings.Contains(html, "Chapter One") {
		t.Error("expected document content in composed HTML")
	}
	if !strings.Contains(html, `class="toc"`) {
		t.Error("expected table of contents in composed HTML")
	}
}

func TestComposeHTMLEmptyDocs(t *testing.T) {
	html, err := ComposeHTML(nil, DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error for empty doc list: %v", err)
	}
	if !strings.Contains(html, `class="toc"`) {
		t.Error("expected an (empty) table of contents to still render")
	}
}

func TestComposeHTMLBadTemplate(t *testing.T) {
	opts := DefaultOptions()
	opts.Template = `{{ .Title `

	_, err := ComposeHTML(nil, opts)
	if err == nil {
		t.Fatal("expected error for malformed template")
	}
	if !strings.Contains(err.Error(), "parsing template") {
		t.Errorf("expected 'parsing template' error, got: %v", err)
	}
}

func TestComposeHTMLTemplateExecFailure(t *testing.T) {
	opts := DefaultOptions()
	opts.Template = `{{ .NoSuchField }}`

	_, err := ComposeHTML(nil, opts)
	if err == nil {
		t.Fatal("expected error for template referencing an unknown field")
	}
	if !strings.Contains(err.Error(), "executing template") {
		t.Errorf("expected 'executing template' error, got: %v", err)
	}
}

func TestCollectKeywordsDedupSort(t *testing.T) {
	docs := []*mdx.Document{
		{Frontmatter: map[string]interface{}{"tags": []interface{}{"zebra", "apple"}}},
		{Frontmatter: map[string]interface{}{"tags": []interface{}{"apple", "mango"}}},
	}

	got := collectKeywords(docs)
	want := "apple, mango, zebra"
	if got != want {
		t.Errorf("collectKeywords() = %q, want %q", got, want)
	}
}
