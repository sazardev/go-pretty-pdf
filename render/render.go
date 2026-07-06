package render

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Options struct {
	Timeout       time.Duration
	HeaderTitle   string
	MarginTop     float64
	MarginBottom  float64
	MarginLeft    float64
	MarginRight   float64
	PaperWidth    float64
	PaperHeight   float64
	NetworkAccess bool
}

func DefaultOptions() Options {
	return Options{
		Timeout:      60 * time.Second,
		HeaderTitle:  "Document",
		MarginTop:    0.8,
		MarginBottom: 0.8,
		MarginLeft:   0.6,
		MarginRight:  0.6,
		PaperWidth:   8.27,
		PaperHeight:  11.69,
		// NetworkAccess defaults to false: the rendered HTML is a
		// self-contained data URI, so outbound network requests are
		// blocked unless explicitly enabled.
		NetworkAccess: false,
	}
}

func RenderToPDF(htmlContent string, outputPath string, opts Options) error {
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			chromedp.NoSandbox,
			chromedp.Headless,
			chromedp.Flag("disable-dev-shm-usage", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	encoded := base64.StdEncoding.EncodeToString([]byte(htmlContent))
	dataURI := "data:text/html;charset=utf-8;base64," + encoded

	headerTpl := fmt.Sprintf(
		`<div style="font-size:8pt;font-family:system-ui,sans-serif;color:#666;padding-left:0.6in;padding-right:0.6in;">%s</div>`,
		opts.HeaderTitle,
	)

	footerTpl := `<div style="font-size:8pt;font-family:system-ui,sans-serif;color:#666;padding-left:0.6in;padding-right:0.6in;"><span class="title"></span><span style="float:right;">Page <span class="pageNumber"></span> of <span class="totalPages"></span></span></div>`

	var pdfBuf []byte

	tasks := chromedp.Tasks{}
	if !opts.NetworkAccess {
		tasks = append(tasks,
			network.Enable(),
			network.SetBlockedURLs().WithURLPatterns([]*network.BlockPattern{
				{URLPattern: "*://*/*", Block: true},
			}),
		)
	}
	tasks = append(tasks,
		chromedp.Navigate(dataURI),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithDisplayHeaderFooter(true).
				WithHeaderTemplate(headerTpl).
				WithFooterTemplate(footerTpl).
				WithGenerateDocumentOutline(true).
				WithGenerateTaggedPDF(true).
				WithMarginTop(opts.MarginTop).
				WithMarginBottom(opts.MarginBottom).
				WithMarginLeft(opts.MarginLeft).
				WithMarginRight(opts.MarginRight).
				WithPaperWidth(opts.PaperWidth).
				WithPaperHeight(opts.PaperHeight).
				Do(ctx)
			return err
		}),
	)

	if err := chromedp.Run(ctx, tasks...); err != nil {
		return fmt.Errorf("chromedp render: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, pdfBuf, 0644); err != nil {
		return fmt.Errorf("writing PDF: %w", err)
	}

	return nil
}

func CheckChromeAvailable() error {
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx)
}
