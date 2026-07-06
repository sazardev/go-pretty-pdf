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

func TestComposeHTMLShowCoverFalse(t *testing.T) {
	docs := []*mdx.Document{
		docWithID("[1.0.0]", "Chapter One", "<h1>Chapter One</h1>"),
	}
	opts := DefaultOptions()
	opts.ShowCover = false

	html, err := ComposeHTML(docs, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(html, `class="cover"`) {
		t.Error("expected cover to be omitted when ShowCover is false")
	}
	if !strings.Contains(html, `class="toc"`) {
		t.Error("expected TOC to still render")
	}
}

func TestComposeHTMLShowTOCFalse(t *testing.T) {
	docs := []*mdx.Document{
		docWithID("[1.0.0]", "Chapter One", "<h1>Chapter One</h1>"),
	}
	opts := DefaultOptions()
	opts.ShowTOC = false

	html, err := ComposeHTML(docs, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(html, `class="toc"`) {
		t.Error("expected TOC to be omitted when ShowTOC is false")
	}
	if !strings.Contains(html, `class="cover"`) {
		t.Error("expected cover to still render")
	}
}

func TestDefaultOptionsShowsEverything(t *testing.T) {
	opts := DefaultOptions()
	if !opts.ShowCover || !opts.ShowTOC {
		t.Error("expected DefaultOptions to show both cover and TOC")
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
