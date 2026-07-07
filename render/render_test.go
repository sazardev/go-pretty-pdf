package render

import (
	"os"
	"path/filepath"
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
	// must still render successfully from the self-contained data URI.
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
