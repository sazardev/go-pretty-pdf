package render

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// coverPxPerIn matches the "px" convention config.ParseCSSUnit already uses
// for CSS lengths (1in == 96px), so a cover image ends up printed at its
// native pixel size rather than being rescaled to fit a paper size.
const coverPxPerIn = 96.0

// coverImageDimensionsIn decodes imagePath's pixel dimensions (without
// loading the full pixel data) and converts them to inches for use as an
// exact PrintToPDF paper size.
func coverImageDimensionsIn(imagePath string) (widthIn, heightIn float64, err error) {
	switch strings.ToLower(filepath.Ext(imagePath)) {
	case ".png", ".jpg", ".jpeg":
	default:
		return 0, 0, fmt.Errorf("cover image %s: unsupported format (expected .png, .jpg, or .jpeg)", imagePath)
	}

	f, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("opening cover image: %w", err)
	}
	defer func() { _ = f.Close() }()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("decoding cover image %s: %w", imagePath, err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return 0, 0, fmt.Errorf("cover image %s has no usable dimensions", imagePath)
	}

	return float64(cfg.Width) / coverPxPerIn, float64(cfg.Height) / coverPxPerIn, nil
}

// coverImageMIMEType maps a cover image's extension to its MIME type for a
// data: URI. Callers must have already validated the extension via
// coverImageDimensionsIn.
func coverImageMIMEType(imagePath string) string {
	if strings.ToLower(filepath.Ext(imagePath)) == ".png" {
		return "image/png"
	}
	return "image/jpeg"
}

// coverPDFTasks builds the chromedp actions that print imagePath, full
// bleed and edge to edge, as a standalone single page sized to exactly
// widthIn x heightIn, writing the resulting PDF bytes into *out once the
// returned tasks run inside a chromedp.Run call. The returned cleanup func
// removes the temporary HTML file the tasks navigate to and must be called
// after that Run completes (defer it right after a nil error check).
func coverPDFTasks(imagePath string, widthIn, heightIn float64, out *[]byte) (chromedp.Tasks, func(), error) {
	imgBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading cover image: %w", err)
	}
	// Embedded as a data: URI rather than referenced by a file:// src:
	// RenderToPDFWithAudit blocks all network requests by default
	// (Options.NetworkAccess == false) via Network.setBlockedURLs("*://*/*"),
	// which — perhaps surprisingly — also matches and blocks file://
	// subresource fetches, not just remote http(s) ones. The top-level
	// Navigate to this cover page's own temp file still works (navigation
	// isn't blocked the same way), but an <img src="file://..."> inside it
	// would silently fail to load under that same default. A data: URI
	// carries the pixels inline, so no subresource fetch — blocked or
	// not — ever happens.
	dataURI := "data:" + coverImageMIMEType(imagePath) + ";base64," + base64.StdEncoding.EncodeToString(imgBytes)

	// No theme CSS, header/footer, or margin here on purpose: this page is
	// the image, exactly as-is, at exactly its own size — nothing else
	// should compete for pixels on it.
	html := fmt.Sprintf(
		`<!DOCTYPE html><html><head><meta charset="UTF-8">`+
			`<style>html,body{margin:0;padding:0;}img{display:block;width:100%%;height:100%%;}</style>`+
			`</head><body><img src="%s"></body></html>`,
		dataURI,
	)

	navURL, cleanup, err := navigationURLFor(html)
	if err != nil {
		return nil, nil, err
	}

	tasks := chromedp.Tasks{
		chromedp.Navigate(navURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			*out, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithPaperWidth(widthIn).
				WithPaperHeight(heightIn).
				Do(ctx)
			return err
		}),
	}
	return tasks, cleanup, nil
}

// mergeCoverAndBody concatenates coverPDF in front of bodyPDF into a single
// PDF written to outputPath, with each side keeping its own page size.
// Page.printToPDF only accepts one paperWidth/paperHeight per call, so a
// cover page sized to a custom image can never share a print pass with the
// rest of the document — the two are printed separately and stitched back
// together here instead.
func mergeCoverAndBody(coverPDF, bodyPDF []byte, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output PDF: %w", err)
	}
	defer func() { _ = out.Close() }()

	rsc := []io.ReadSeeker{bytes.NewReader(coverPDF), bytes.NewReader(bodyPDF)}
	if err := api.MergeRaw(rsc, out, false, nil); err != nil {
		return fmt.Errorf("merging cover and body PDFs: %w", err)
	}
	return nil
}
