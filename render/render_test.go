package render

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func requireChrome(t *testing.T) {
	t.Helper()
	if err := CheckChromeAvailable(); err != nil {
		t.Skipf("Chrome/Chromium not available: %v", err)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Timeout != 60*time.Second {
		t.Errorf("expected 60s timeout, got %v", opts.Timeout)
	}
	if opts.NetworkAccess {
		t.Error("expected NetworkAccess to default to false (network blocked)")
	}
	if opts.PaperWidth == 0 || opts.PaperHeight == 0 {
		t.Error("expected non-zero default paper size")
	}
	if !opts.PageNumbers {
		t.Error("expected PageNumbers to default to true")
	}
	if !opts.ShowHeader {
		t.Error("expected ShowHeader to default to true")
	}
	// Left/right must stay 0: that's what lets a dark theme's background
	// bleed to the physical page edge instead of leaving a white gutter
	// Chrome's print margin never paints (see pageChrome/base.css).
	if opts.MarginLeft != 0 || opts.MarginRight != 0 {
		t.Errorf("expected zero left/right margin for edge-to-edge background, got left=%v right=%v", opts.MarginLeft, opts.MarginRight)
	}
	if opts.MarginTop <= 0 || opts.MarginBottom <= 0 {
		t.Error("expected a positive top/bottom margin — it's the only space the header/footer template can render into")
	}
}

func TestPageChrome(t *testing.T) {
	bg, muted := pageChrome(`<style>:root{--pdf-bg:#282828;--pdf-muted:#a89984;}</style>`)
	if bg != "#282828" {
		t.Errorf("pageChrome() bg = %q, want #282828", bg)
	}
	if muted != "#a89984" {
		t.Errorf("pageChrome() muted = %q, want #a89984", muted)
	}

	// No theme vars at all (e.g. a raw hand-written CSS file) must fall
	// back to legible light-theme defaults, not empty strings that would
	// produce a broken "color:;" declaration in the header/footer style.
	bg, muted = pageChrome(`<style>body{color:black;}</style>`)
	if bg == "" || muted == "" {
		t.Errorf("pageChrome() with no --pdf- vars returned empty values: bg=%q muted=%q", bg, muted)
	}
}

// TestPageChromeIgnoresVarsOutsideStyleBlock guards against pageChrome
// picking up a `--pdf-x: value;`-shaped string from document body content
// (e.g. a chapter's code sample documenting theme variables) and letting it
// override the real theme's header/footer colors declared in <style>.
func TestPageChromeIgnoresVarsOutsideStyleBlock(t *testing.T) {
	html := `<html><head><style>:root{--pdf-bg:#282828;--pdf-muted:#a89984;}</style></head>` +
		`<body><pre>--pdf-bg: #ff0000;</pre></body></html>`

	bg, muted := pageChrome(html)
	if bg != "#282828" {
		t.Errorf("pageChrome() bg = %q, want #282828 (body content must not override it)", bg)
	}
	if muted != "#a89984" {
		t.Errorf("pageChrome() muted = %q, want #a89984", muted)
	}
}

func TestRenderToPDFPageNumbersAndHeaderDisabled(t *testing.T) {
	requireChrome(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.pdf")

	opts := DefaultOptions()
	opts.PageNumbers = false
	opts.ShowHeader = false
	opts.HeaderTitle = "Should Not Appear"

	if err := RenderToPDF(`<html><body><h1>Content</h1></body></html>`, outPath, opts); err != nil {
		t.Fatalf("RenderToPDF failed: %v", err)
	}
	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		t.Fatal("expected a non-empty PDF even with header/footer disabled")
	}
}

func TestRenderToPDFProducesFile(t *testing.T) {
	requireChrome(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.pdf")

	html := `<html><body><h1>Hello, PDF</h1></body></html>`
	if err := RenderToPDF(html, outPath, DefaultOptions()); err != nil {
		t.Fatalf("RenderToPDF failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF output")
	}
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		t.Error("expected output to start with a %PDF magic header")
	}
}

func TestRenderToPDFCreatesOutputDir(t *testing.T) {
	requireChrome(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "nested", "deeper", "out.pdf")

	if err := RenderToPDF(`<html><body>x</body></html>`, outPath, DefaultOptions()); err != nil {
		t.Fatalf("RenderToPDF failed: %v", err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected nested output directory to be created: %v", err)
	}
}

func TestRenderToPDFNetworkBlockedByDefaultStillRenders(t *testing.T) {
	requireChrome(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.pdf")

	// Content that would trigger an outbound request if network access
	// were allowed; with NetworkAccess left at its default (false), this
	// must still render successfully.
	html := `<html><body><img src="https://example.invalid/nonexistent.png"><h1>Local content</h1></body></html>`
	opts := DefaultOptions()

	if err := RenderToPDF(html, outPath, opts); err != nil {
		t.Fatalf("expected rendering to succeed even with network blocked: %v", err)
	}
	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		t.Fatal("expected a non-empty PDF despite the blocked remote image")
	}
}

// TestRenderToPDFWithAuditContextStopsOnCancellation guards a real bug:
// RenderToPDFWithAudit used to root its Chrome allocator in
// context.Background() regardless of what the caller passed in, so
// canceling a caller's context (client disconnect, SIGINT wired to context
// cancellation) never actually stopped an in-flight render — it kept
// running until opts.Timeout regardless. RenderToPDFWithAuditContext must
// tear down well before that timeout once ctx is canceled.
func TestRenderToPDFWithAuditContextStopsOnCancellation(t *testing.T) {
	requireChrome(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.pdf")

	opts := DefaultOptions()
	opts.Timeout = 60 * time.Second // deliberately long: cancellation, not the timeout, must be what stops this

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled before the render even starts

	start := time.Now()
	_, err := RenderToPDFWithAuditContext(ctx, `<html><body>x</body></html>`, outPath, opts)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error for a render started with an already-canceled context")
	}
	if elapsed > 15*time.Second {
		t.Errorf("expected cancellation to stop the render quickly, took %v (opts.Timeout was %v)", elapsed, opts.Timeout)
	}
}

func TestNavigationURLForReturnsFileURLAndCleansUp(t *testing.T) {
	navURL, cleanup, err := navigationURLFor("<html><body>hi</body></html>")
	if err != nil {
		t.Fatalf("navigationURLFor: %v", err)
	}
	if !strings.HasPrefix(navURL, "file://") {
		t.Errorf("expected a file:// URL, got %q", navURL)
	}

	path := strings.TrimPrefix(navURL, "file://")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected temp html file to exist at %q: %v", path, err)
	}

	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected cleanup() to remove the temp file, stat err = %v", err)
	}
}

// TestRenderToPDFLargeDocumentPastOldDataURILimit is a regression test for
// a real bug: this package used to navigate Chrome to a
// "data:text/html;base64,..." URI, which silently fails
// ("net::ERR_ABORTED") once the encoded payload crosses roughly 2MB. A
// book-length document (hundreds of thousands of words of prose and code)
// crosses that threshold easily. This generates content well past the old
// limit and confirms it still renders — the fix (navigating to a temp
// file via file://) has no such ceiling.
func TestRenderToPDFLargeDocumentPastOldDataURILimit(t *testing.T) {
	requireChrome(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.pdf")

	var body strings.Builder
	body.WriteString("<html><body>")
	// ~120 bytes per paragraph * ~30000 = ~3.6MB of raw HTML, comfortably
	// past the ~2MB (post-base64) size that used to break data URI
	// navigation, even before base64's own ~33% size inflation is applied.
	for i := 0; i < 30000; i++ {
		fmt.Fprintf(&body, "<p>Paragraph number %d with some filler text to pad it out a bit.</p>", i)
	}
	body.WriteString("</body></html>")
	html := body.String()

	if len(html) < 2*1024*1024 {
		t.Fatalf("test fixture too small to exercise the old data URI limit: %d bytes", len(html))
	}

	if err := RenderToPDF(html, outPath, DefaultOptions()); err != nil {
		t.Fatalf("expected large document to render successfully, got: %v", err)
	}
	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		t.Fatal("expected a non-empty PDF for the large document")
	}
}
