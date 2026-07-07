package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	prettypdf "github.com/sazardev/go-pretty-pdf"
	"github.com/sazardev/go-pretty-pdf/theme"
)

// docsPDFDefault is the canonical, stable download URL (used in the
// sitemap and as the href before any client-side JS runs). It mirrors the
// site's own default theme (classic).
const docsPDFDefault = "go-pretty-pdf-docs.pdf"

// docsPDFFilename returns the per-theme download artifact name. The site's
// theme switcher (site.js) rewrites the download button's href to match
// whichever of these the visitor currently has selected, so "download the
// docs" always matches what they're looking at.
func docsPDFFilename(themeID string) string {
	return "go-pretty-pdf-docs-" + themeID + ".pdf"
}

var readmeBadgesRe = regexp.MustCompile(`(?m)^\[!\[.*\n?`)

// generateDocsPDF renders README.md + docs/cli.md + CHANGELOG.md into one
// downloadable PDF per builtin theme, using the same code path a real
// user's `pretty-pdf build` would take — dogfooding the actual public
// library API (mdx parser, theme package, chromedp render pipeline), not a
// raw HTML screenshot. Best-effort: like generateRasterAssets, it must not
// break `go run ./scripts/docsgen` for contributors without Chrome
// installed locally.
func generateDocsPDF(outDir string, readme, cli, changelog []byte) {
	srcDir, err := os.MkdirTemp("", "go-pretty-pdf-docs-src-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: skipping docs PDF, could not create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(srcDir)

	// Badge images point at shields.io/pkg.go.dev and would just render as
	// broken-image glyphs: WithNetworkAccess defaults to false, matching
	// the CLI's own safe default for untrusted MDX sources.
	cleanReadme := readmeBadgesRe.ReplaceAll(readme, nil)

	docs := []struct {
		file, id, title string
		body            []byte
	}{
		{"01-docs.mdx", "[1.0.0]", "go-pretty-pdf", cleanReadme},
		{"02-cli.mdx", "[2.0.0]", "CLI Reference", cli},
		{"03-changelog.mdx", "[3.0.0]", "Changelog", changelog},
	}
	for _, d := range docs {
		content := fmt.Sprintf("---\nid: %q\ntitle: %q\n---\n\n%s", d.id, d.title, d.body)
		if err := os.WriteFile(filepath.Join(srcDir, d.file), []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping docs PDF, could not write %s: %v\n", d.file, err)
			return
		}
	}

	for _, t := range siteThemes {
		outPath := filepath.Join(outDir, docsPDFFilename(t.ID))
		if err := buildOneDocsPDF(srcDir, outPath, t.ID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping docs PDF (theme %s): %v\n", t.ID, err)
			continue
		}
		if t.ID == theme.NameClassic {
			if err := copyFile(outPath, filepath.Join(outDir, docsPDFDefault)); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not stage default docs PDF: %v\n", err)
			}
		}
	}
}

func buildOneDocsPDF(srcDir, outPath, themeID string) error {
	pdf, err := prettypdf.New(
		prettypdf.WithSourceDir(srcDir),
		prettypdf.WithOutputFile(outPath),
		prettypdf.WithTitle("go-pretty-pdf"),
		prettypdf.WithSubtitle("Write Markdown. Ship a book."),
		prettypdf.WithAuthor("sazardev"),
		prettypdf.WithHeaderTitle("go-pretty-pdf — Documentation"),
		prettypdf.WithThemeName(themeID, theme.Options{}),
		prettypdf.WithTimeout(90*time.Second),
	)
	if err != nil {
		return fmt.Errorf("configuring PDF build: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if err := pdf.Build(ctx); err != nil {
		return fmt.Errorf("building PDF: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
