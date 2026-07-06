package prettypdf

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sazardev/go-pretty-pdf/config"
	"github.com/sazardev/go-pretty-pdf/mdx"
	"github.com/sazardev/go-pretty-pdf/render"
	"github.com/sazardev/go-pretty-pdf/theme"
)

const testSourceDir = "src"

func writeFixtureMDX(t *testing.T, dir, filename, id, title string) string {
	t.Helper()
	content := "---\nid: \"" + id + "\"\ntitle: " + title + "\n---\n\n# " + title + "\n"
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	fn()

	_ = w.Close()
	os.Stderr = orig
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestNewDefaults(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.sourceDir != "book" {
		t.Errorf("expected default source dir 'book', got %q", p.sourceDir)
	}
	if p.outputFile != "out.pdf" {
		t.Errorf("expected default output file 'out.pdf', got %q", p.outputFile)
	}
	if p.composeOpts.Title != "Document" {
		t.Errorf("expected default title 'Document', got %q", p.composeOpts.Title)
	}
	if p.renderOpts.HeaderTitle != p.composeOpts.Title {
		t.Errorf("expected renderOpts.HeaderTitle to mirror composeOpts.Title, got %q vs %q", p.renderOpts.HeaderTitle, p.composeOpts.Title)
	}
	if p.renderOpts.NetworkAccess {
		t.Error("expected network access to default to false")
	}
}

func TestOptionsMutateFields(t *testing.T) {
	p, err := New(
		WithSourceDir("docs"),
		WithOutputFile("book.pdf"),
		WithTitle("T"),
		WithSubtitle("S"),
		WithAuthor("A"),
		WithCSS("css{}"),
		WithTemplate("<html></html>"),
		WithTimeout(5*time.Second),
		WithHeaderTitle("Header"),
		WithVerbose(true),
		WithRenderMargins(1, 2, 3, 4),
		WithPaperSize(8, 10),
		WithNetworkAccess(true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := map[string]bool{
		"sourceDir":     p.sourceDir == "docs",
		"outputFile":    p.outputFile == "book.pdf",
		"title":         p.composeOpts.Title == "T",
		"subtitle":      p.composeOpts.Subtitle == "S",
		"author":        p.composeOpts.Author == "A",
		"css":           p.composeOpts.CSS == "css{}",
		"template":      p.composeOpts.Template == "<html></html>",
		"timeout":       p.renderOpts.Timeout == 5*time.Second,
		"headerTitle":   p.renderOpts.HeaderTitle == "Header",
		"verbose":       p.verbose,
		"marginTop":     p.renderOpts.MarginTop == 1,
		"marginBottom":  p.renderOpts.MarginBottom == 2,
		"marginLeft":    p.renderOpts.MarginLeft == 3,
		"marginRight":   p.renderOpts.MarginRight == 4,
		"paperWidth":    p.renderOpts.PaperWidth == 8,
		"paperHeight":   p.renderOpts.PaperHeight == 10,
		"networkAccess": p.renderOpts.NetworkAccess,
	}
	for name, ok := range checks {
		if !ok {
			t.Errorf("option did not apply expected value: %s", name)
		}
	}
}

func TestWithThemeAndConflict(t *testing.T) {
	minimal, ok := theme.Get("minimal")
	if !ok {
		t.Fatal("expected builtin theme 'minimal' to be registered")
	}

	p, err := New(WithTheme(minimal))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.composeOpts.CSS != minimal.CSS {
		t.Error("expected WithTheme to set composeOpts.CSS to the theme's CSS")
	}

	// Last CSS-setting option wins.
	p2, err := New(WithTheme(minimal), WithCSS("custom"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p2.composeOpts.CSS != "custom" {
		t.Error("expected the later WithCSS to override the earlier WithTheme")
	}
}

func TestWithThemeName(t *testing.T) {
	p, err := New(WithThemeName("minimal", theme.Options{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(p.composeOpts.CSS, "minimal") {
		t.Errorf("expected resolved CSS to include the minimal theme's marker comment, got: %.100s...", p.composeOpts.CSS)
	}
	if !p.composeOpts.ShowCover || !p.composeOpts.ShowTOC {
		t.Error("expected minimal theme to default all sections on")
	}
	if !p.renderOpts.PageNumbers || !p.renderOpts.ShowHeader {
		t.Error("expected minimal theme to default page numbers/header on")
	}

	p2, err := New(WithThemeName("corporate", theme.Options{
		Sections: theme.Sections{Cover: theme.BoolPtr(false), PageNumbers: theme.BoolPtr(false)},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p2.composeOpts.ShowCover {
		t.Error("expected --no-cover equivalent to disable ShowCover")
	}
	if p2.renderOpts.PageNumbers {
		t.Error("expected page numbers to be disabled")
	}
	if !p2.composeOpts.ShowTOC || !p2.renderOpts.ShowHeader {
		t.Error("expected untouched sections to stay enabled")
	}
}

func TestWithThemeNameUnknownWarns(t *testing.T) {
	out := captureStderr(t, func() {
		_, err := New(WithVerbose(true), WithThemeName("does-not-exist", theme.Options{}))
		if err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "does-not-exist") {
		t.Errorf("expected warning about unknown theme, got: %q", out)
	}
}

func TestWithComponentRegisters(t *testing.T) {
	p, err := New(WithComponent("Custom", func(attrs map[string]string, inner string) string {
		return "<div>" + inner + "</div>"
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "test.mdx")
	content := "---\nid: \"[1.0.0]\"\ntitle: Test\n---\n\n<Custom>hi</Custom>\n"
	if writeErr := os.WriteFile(path, []byte(content), 0644); writeErr != nil {
		t.Fatal(writeErr)
	}
	doc, err := p.parser.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.HTML, "<div>hi</div>") {
		t.Errorf("expected registered component to transpile, got: %s", doc.HTML)
	}
}

func TestWithConfig(t *testing.T) {
	cfg := &config.Config{
		Source:   testSourceDir,
		Output:   "o.pdf",
		Title:    "CfgTitle",
		Subtitle: "CfgSub",
		Author:   "CfgAuthor",
	}
	p, err := New(WithConfig(cfg))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.sourceDir != testSourceDir || p.outputFile != "o.pdf" {
		t.Errorf("expected source/output from config, got %q/%q", p.sourceDir, p.outputFile)
	}
	if p.composeOpts.Title != "CfgTitle" || p.composeOpts.Subtitle != "CfgSub" || p.composeOpts.Author != "CfgAuthor" {
		t.Error("expected title/subtitle/author from config")
	}
}

func TestWithConfigCSSAndTemplateReadsFiles(t *testing.T) {
	dir := t.TempDir()
	cssPath := filepath.Join(dir, "c.css")
	tmplPath := filepath.Join(dir, "t.html")
	if err := os.WriteFile(cssPath, []byte("body{color:red}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmplPath, []byte("<html>{{.Body}}</html>"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{CSS: cssPath, Template: tmplPath}
	p, err := New(WithConfigCSSAndTemplate(cfg))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.composeOpts.CSS != "body{color:red}" {
		t.Errorf("expected CSS to be loaded from file, got %q", p.composeOpts.CSS)
	}
	if p.composeOpts.Template != "<html>{{.Body}}</html>" {
		t.Errorf("expected template to be loaded from file, got %q", p.composeOpts.Template)
	}
}

// Regression test: warnings from WithConfigCSSAndTemplate must be printed
// regardless of whether WithVerbose is applied before or after it, since
// New() flushes buffered warnings only after every option has run.
func TestWithConfigCSSAndTemplateWarningsOrderIndependent(t *testing.T) {
	cfg := &config.Config{CSS: "/nonexistent/does-not-exist.css"}

	t.Run("verbose before config option", func(t *testing.T) {
		out := captureStderr(t, func() {
			_, err := New(WithVerbose(true), WithConfigCSSAndTemplate(cfg))
			if err != nil {
				t.Fatal(err)
			}
		})
		if !strings.Contains(out, "does-not-exist.css") {
			t.Errorf("expected warning to be printed, got: %q", out)
		}
	})

	t.Run("verbose after config option", func(t *testing.T) {
		out := captureStderr(t, func() {
			_, err := New(WithConfigCSSAndTemplate(cfg), WithVerbose(true))
			if err != nil {
				t.Fatal(err)
			}
		})
		if !strings.Contains(out, "does-not-exist.css") {
			t.Errorf("expected warning to be printed even when WithVerbose comes after, got: %q", out)
		}
	})
}

func TestWithFullConfig(t *testing.T) {
	cfg := &config.Config{
		Source: testSourceDir,
		Vars:   map[string]string{"name": "World"},
		Render: config.RenderConfig{
			Timeout:     "5s",
			Paper:       config.PaperLetter,
			MarginTop:   "10mm",
			HeaderTitle: "Full Config Header for {{name}}",
		},
	}

	p, err := New(WithFullConfig(cfg))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.sourceDir != testSourceDir {
		t.Errorf("expected source from config, got %q", p.sourceDir)
	}
	if p.renderOpts.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", p.renderOpts.Timeout)
	}
	if p.renderOpts.PaperWidth != 8.5 || p.renderOpts.PaperHeight != 11 {
		t.Errorf("expected letter paper size, got %vx%v", p.renderOpts.PaperWidth, p.renderOpts.PaperHeight)
	}
	if p.renderOpts.HeaderTitle != "Full Config Header for World" {
		t.Errorf("expected header title from config with vars substituted, got %q", p.renderOpts.HeaderTitle)
	}
	defOpts := render.DefaultOptions()
	if p.renderOpts.MarginBottom != defOpts.MarginBottom {
		t.Errorf("expected unset margin to fall back to default, got %v", p.renderOpts.MarginBottom)
	}

	dir := t.TempDir()
	path := writeFixtureMDX(t, dir, "test.mdx", "[1.0.0]", "Hello {{name}}")
	doc, err := p.parser.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.Title(), "World") {
		t.Errorf("expected vars from config to be substituted, got title %q", doc.Title())
	}
}

func TestPDFParseDirAndComposeHTML(t *testing.T) {
	dir := t.TempDir()
	writeFixtureMDX(t, dir, "a.mdx", "[1.0.0]", "Chapter One")

	p, err := New(WithSourceDir(dir))
	if err != nil {
		t.Fatal(err)
	}

	docs, err := p.ParseDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(docs))
	}

	html, err := p.ComposeHTML(docs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Chapter One") {
		t.Error("expected composed HTML to contain the document content")
	}
}

func TestPDFValidateNoValidator(t *testing.T) {
	dir := t.TempDir()
	writeFixtureMDX(t, dir, "a.mdx", "[1.0.0]", "Chapter One")

	p, err := New(WithSourceDir(dir))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := p.Validate(context.Background()); err == nil {
		t.Fatal("expected error when no validator is configured")
	}

	if errs := p.ValidateDoc(&mdx.Document{}); errs != nil {
		t.Error("expected ValidateDoc to return nil without a configured validator")
	}
	if errs := p.ValidateAll(nil); errs != nil {
		t.Error("expected ValidateAll to return nil without a configured validator")
	}
}

func TestPDFValidateWithValidator(t *testing.T) {
	dir := t.TempDir()
	// Missing title triggers a required-frontmatter validation error.
	content := "---\nid: \"[1.0.0]\"\n---\n\n# No title field\n"
	if err := os.WriteFile(filepath.Join(dir, "a.mdx"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := New(WithSourceDir(dir), WithValidator(mdx.NewDefaultValidator()))
	if err != nil {
		t.Fatal(err)
	}

	errs, err := p.Validate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Error("expected validation errors for a document missing the title field")
	}
}

func TestPDFBuildParseError(t *testing.T) {
	dir := t.TempDir() // empty: no .mdx files

	p, err := New(WithSourceDir(dir))
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Build(context.Background()); err == nil {
		t.Fatal("expected Build to fail when the source directory has no .mdx files")
	}
}

func TestPDFBuildRenderSmoke(t *testing.T) {
	if err := render.CheckChromeAvailable(); err != nil {
		t.Skipf("Chrome/Chromium not available: %v", err)
	}

	dir := t.TempDir()
	writeFixtureMDX(t, dir, "a.mdx", "[1.0.0]", "Chapter One")

	outPath := filepath.Join(t.TempDir(), "out.pdf")
	p, err := New(WithSourceDir(dir), WithOutputFile(outPath))
	if err != nil {
		t.Fatal(err)
	}

	if buildErr := p.Build(context.Background()); buildErr != nil {
		t.Fatalf("Build failed: %v", buildErr)
	}

	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		t.Fatal("expected a non-empty PDF to be produced")
	}
}
