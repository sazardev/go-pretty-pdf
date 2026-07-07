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

	"github.com/sazardev/go-pretty-pdf/theme"
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
	PageNumbers   bool
	ShowHeader    bool
	// ChromeExecPath, when non-empty, pins chromedp to this specific
	// Chrome/Chromium/chrome-headless-shell binary instead of letting it
	// search the system's default install locations. Callers resolving a
	// browser via chromemgr.EnsureChrome pass its result straight through
	// here; leaving it empty preserves the previous default-discovery
	// behavior exactly.
	ChromeExecPath string
}

func DefaultOptions() Options {
	return Options{
		Timeout:     60 * time.Second,
		HeaderTitle: "Document",
		// Left/right are 0 so the page background can bleed edge to edge —
		// the visual reading margin instead comes from CSS padding on
		// <body> (see theme/assets/base.css), which the theme's own
		// background color paints straight through. Top/bottom stay as a
		// real Chrome print margin (not CSS padding): that's the only
		// space the running header/page-number footer can render into,
		// and — unlike body padding, which only applies to the first/last
		// page per the CSS Fragmentation spec — a native margin repeats
		// identically on every page. RenderToPDF colors that strip to
		// match the theme, so it doesn't show up as a white band either.
		MarginTop:    0.6,
		MarginBottom: 0.6,
		MarginLeft:   0,
		MarginRight:  0,
		PaperWidth:   8.27,
		PaperHeight:  11.69,
		// NetworkAccess defaults to false: the rendered HTML is a
		// self-contained data URI, so outbound network requests are
		// blocked unless explicitly enabled.
		NetworkAccess: false,
		PageNumbers:   true,
		ShowHeader:    true,
	}
}

// RenderToPDF composes htmlContent into a PDF at outputPath. It never
// returns audit findings — use RenderToPDFWithAudit for that — but always
// keeps this signature exactly as-is for API stability.
func RenderToPDF(htmlContent string, outputPath string, opts Options) error {
	_, err := RenderToPDFWithAudit(htmlContent, outputPath, opts)
	return err
}

// RenderToPDFWithAudit does exactly what RenderToPDF does, and additionally
// runs a best-effort visual/structural audit (see audit.go) of the
// composed document, returning its findings alongside any render error.
// The audit never turns a successful render into a failure: an audit
// finding is a warning about the *output*, not a reason to reject it.
func RenderToPDFWithAudit(htmlContent string, outputPath string, opts Options) (*AuditReport, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.Flag("disable-dev-shm-usage", true),
		// Chrome can take longer than chromedp's 20s default to print its
		// DevTools websocket URL on a cold/loaded CI runner (e.g. right
		// after a fresh install); give it more room to avoid a spurious
		// "websocket url timeout reached" before the browser even starts.
		chromedp.WSURLReadTimeout(45*time.Second),
	)
	if opts.ChromeExecPath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(opts.ChromeExecPath))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	encoded := base64.StdEncoding.EncodeToString([]byte(htmlContent))
	dataURI := "data:text/html;charset=utf-8;base64," + encoded

	// Chrome renders the header/footer template inside the top/bottom
	// margin strip it was given — a strip that's otherwise always blank
	// white, regardless of the page's own background. Painting that strip
	// with the theme's own --pdf-bg (and using --pdf-muted for the text
	// instead of a hardcoded gray) is what makes a dark theme's PDF look
	// like one continuous dark page instead of dark content floating in a
	// white border.
	//
	// The wrapper div must be given an explicit height equal to the
	// margin, in the same absolute unit (inches) passed to
	// WithMarginTop/Bottom — height:100% is unreliable here since Chrome
	// doesn't reliably give this template's ancestor chain a resolvable
	// height, and without *some* explicit height the div only grows as
	// tall as its one line of text, leaving the rest of the margin strip
	// (most of it) unpainted white.
	bg, mutedText := pageChrome(htmlContent)
	wrap := func(marginIn float64, alignItems string) string {
		return fmt.Sprintf(
			`width:100%%;height:%.3fin;box-sizing:border-box;margin:0;background:%s;`+
				`-webkit-print-color-adjust:exact;print-color-adjust:exact;`+
				`display:flex;align-items:%s;`,
			marginIn, bg, alignItems,
		)
	}
	headerWrap := wrap(opts.MarginTop, "flex-end")      // text sits just above the content
	footerWrap := wrap(opts.MarginBottom, "flex-start") // text sits just below the content
	textStyle := fmt.Sprintf(
		// 15mm left/right matches <body>'s own reading margin (see
		// theme/assets/base.css) so header/footer text lines up with the
		// content edge instead of the physical page edge.
		`width:100%%;box-sizing:border-box;font-size:8pt;font-family:system-ui,sans-serif;color:%s;padding:2pt 15mm;`,
		mutedText,
	)

	// Chrome renders this template HTML into its own little document with
	// the usual user-agent default (a few pixels of <body> margin) —
	// enough to leave a sliver of the true page-edge unpainted above the
	// header even though our own div's height matches the margin exactly.
	// Reset it explicitly rather than guess at an offset.
	resetStyle := `<style>html,body{margin:0;padding:0;}</style>`

	headerTpl := fmt.Sprintf(`%s<div style="%s"><div style="%s">&nbsp;</div></div>`, resetStyle, headerWrap, textStyle)
	if opts.ShowHeader && opts.HeaderTitle != "" {
		headerTpl = fmt.Sprintf(
			`%s<div style="%s"><div style="%s">%s</div></div>`,
			resetStyle, headerWrap, textStyle, opts.HeaderTitle,
		)
	}

	footerTpl := fmt.Sprintf(`%s<div style="%s"><div style="%s">&nbsp;</div></div>`, resetStyle, footerWrap, textStyle)
	if opts.PageNumbers {
		footerTpl = fmt.Sprintf(
			`%s<div style="%s"><div style="%s"><span class="title"></span><span style="float:right;">Page <span class="pageNumber"></span> of <span class="totalPages"></span></span></div></div>`,
			resetStyle, footerWrap, textStyle,
		)
	}

	var pdfBuf []byte
	var domIssues []Issue

	tasks := chromedp.Tasks{}
	if !opts.NetworkAccess {
		tasks = append(tasks,
			network.Enable(),
			network.SetBlockedURLs().WithURLPatterns([]*network.BlockPattern{
				{URLPattern: "*://*/*", Block: true},
			}),
		)
	}
	// Chrome reserves a small fixed inset (~0.2in, empirically measured —
	// independent of the margin size and not overridable via the template's
	// own CSS) at the very top/bottom of the page whenever header/footer
	// templates are displayed at all, even fully empty ones. There's no way
	// to make that sliver match the theme's background short of not paying
	// for the header/footer mechanism in the first place, so skip it
	// entirely when the caller wants neither a header nor page numbers —
	// pair with MarginTop/Bottom: 0 for a genuinely gap-free page.
	needsHeaderFooter := opts.ShowHeader || opts.PageNumbers

	tasks = append(tasks,
		chromedp.Navigate(dataURI),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Runs against the fully-loaded document, before PrintToPDF
			// hands it to Chrome's print engine — see audit.go for what it
			// checks and why it has to happen here rather than after.
			domIssues = runDOMAudit(ctx, needsHeaderFooter)
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithDisplayHeaderFooter(needsHeaderFooter).
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
		return nil, fmt.Errorf("chromedp render: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, pdfBuf, 0644); err != nil {
		return nil, fmt.Errorf("writing PDF: %w", err)
	}

	report := &AuditReport{Issues: append(domIssues, auditPDFBytes(pdfBuf)...)}
	return report, nil
}

// pageChrome pulls the page background and muted-text colors out of the
// composed document's own CSS, so the native header/footer strip can be
// painted to match whatever theme is active instead of hardcoded
// white/gray. Falls back to sensible light-theme defaults if a theme
// doesn't declare one (shouldn't happen for any builtin theme — see
// theme.TestBuiltinThemesProduceSiteVars — but custom/raw CSS files are
// user-authored and may omit either).
func pageChrome(htmlContent string) (bg, mutedText string) {
	vars := theme.ExtractCSSVars(htmlContent)
	bg = vars["bg"]
	if bg == "" {
		bg = "#ffffff"
	}
	mutedText = vars["muted"]
	if mutedText == "" {
		mutedText = "#666666"
	}
	return bg, mutedText
}

func CheckChromeAvailable() error {
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WSURLReadTimeout(15*time.Second),
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	return chromedp.Run(ctx)
}
