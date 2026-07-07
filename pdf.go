package prettypdf

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sazardev/go-pretty-pdf/compose"
	"github.com/sazardev/go-pretty-pdf/config"
	"github.com/sazardev/go-pretty-pdf/mdx"
	"github.com/sazardev/go-pretty-pdf/render"
	"github.com/sazardev/go-pretty-pdf/theme"
)

type PDF struct {
	sourceDir       string
	outputFile      string
	parser          *mdx.Parser
	composeOpts     compose.Options
	renderOpts      render.Options
	validator       mdx.Validator
	verbose         bool
	pendingWarnings []string
	headerTitleSet  bool
}

type ComposeOptions = compose.Options

func DefaultComposeOptions() ComposeOptions {
	return compose.DefaultOptions()
}

func ComposeHTML(docs []*mdx.Document, opts ComposeOptions) (string, error) {
	return compose.ComposeHTML(docs, opts)
}

type Option func(*PDF)

func WithSourceDir(dir string) Option {
	return func(p *PDF) {
		p.sourceDir = dir
	}
}

func WithOutputFile(path string) Option {
	return func(p *PDF) {
		p.outputFile = path
	}
}

func WithTitle(title string) Option {
	return func(p *PDF) {
		p.composeOpts.Title = title
	}
}

func WithSubtitle(subtitle string) Option {
	return func(p *PDF) {
		p.composeOpts.Subtitle = subtitle
	}
}

func WithAuthor(author string) Option {
	return func(p *PDF) {
		p.composeOpts.Author = author
	}
}

func WithCSS(css string) Option {
	return func(p *PDF) {
		p.composeOpts.CSS = css
	}
}

func WithTemplate(html string) Option {
	return func(p *PDF) {
		p.composeOpts.Template = html
	}
}

// WithTheme applies a raw builtin/synthetic Theme's CSS as-is, with no
// customization (colors/fonts/sections/density) and no section toggles
// applied. It shares composeOpts.CSS with WithCSS and WithThemeName —
// whichever of these options is applied last wins, since New() applies
// options in the order they're passed. Most callers should prefer
// WithThemeName, which resolves section toggles (cover/TOC/page
// numbers/header) into composeOpts/renderOpts too.
func WithTheme(t theme.Theme) Option {
	return func(p *PDF) {
		if t.CSS != "" {
			p.composeOpts.CSS = t.CSS
		}
	}
}

// WithThemeName resolves a theme by name — a builtin ("default",
// "corporate", ...), a custom theme discovered in ./themes/ or the global
// themes directory, or a direct path to a .theme.yml/.css file — applies
// opts customization (colors, fonts, density, network fonts), and wires
// the resulting section toggles (cover, TOC, page numbers, header) into
// composeOpts/renderOpts.
func WithThemeName(name string, opts theme.Options) Option {
	return func(p *PDF) {
		cwd, _ := os.Getwd()
		css, sections, err := theme.ResolveByName(name, opts, cwd)
		if err != nil {
			p.pendingWarnings = append(p.pendingWarnings, fmt.Sprintf("theme %q: %v", name, err))
			return
		}
		p.composeOpts.CSS = css
		p.composeOpts.ShowCover = sections.Cover
		p.composeOpts.ShowTOC = sections.TOC
		p.renderOpts.PageNumbers = sections.PageNumbers
		p.renderOpts.ShowHeader = sections.Header
	}
}

func WithComponent(name string, handler mdx.ComponentHandler) Option {
	return func(p *PDF) {
		p.parser.RegisterComponent(name, handler)
	}
}

func WithValidator(v mdx.Validator) Option {
	return func(p *PDF) {
		p.validator = v
	}
}

func WithTimeout(d time.Duration) Option {
	return func(p *PDF) {
		p.renderOpts.Timeout = d
	}
}

// WithHeaderTitle sets the PDF page header text. If never called, New()
// defaults it to the document title (WithTitle/composeOpts.Title).
func WithHeaderTitle(title string) Option {
	return func(p *PDF) {
		p.renderOpts.HeaderTitle = title
		p.headerTitleSet = true
	}
}

func WithVerbose(v bool) Option {
	return func(p *PDF) {
		p.verbose = v
	}
}

func WithVars(vars map[string]string) Option {
	return func(p *PDF) {
		p.parser.SetVars(vars)
	}
}

func WithRenderMargins(top, bottom, left, right float64) Option {
	return func(p *PDF) {
		p.renderOpts.MarginTop = top
		p.renderOpts.MarginBottom = bottom
		p.renderOpts.MarginLeft = left
		p.renderOpts.MarginRight = right
	}
}

func WithPaperSize(width, height float64) Option {
	return func(p *PDF) {
		p.renderOpts.PaperWidth = width
		p.renderOpts.PaperHeight = height
	}
}

// WithNetworkAccess controls whether headless Chrome may make outbound
// network requests while rendering. It defaults to false: the composed
// HTML is a self-contained data URI, so network access is blocked to
// prevent SSRF/exfiltration from untrusted MDX content (e.g. a malicious
// <img> or <script> tag). Enable it only if your documents intentionally
// reference remote images, fonts, or other resources by URL.
func WithNetworkAccess(enabled bool) Option {
	return func(p *PDF) {
		p.renderOpts.NetworkAccess = enabled
	}
}

// WithChromeExecPath pins rendering to a specific Chrome/Chromium binary
// instead of chromedp's default system discovery. Leave unset (or pass "")
// to keep the default behavior. See the chromemgr package for resolving
// this automatically, including downloading a browser when none is
// installed.
func WithChromeExecPath(path string) Option {
	return func(p *PDF) {
		p.renderOpts.ChromeExecPath = path
	}
}

func WithConfig(cfg *config.Config) Option {
	return func(p *PDF) {
		if cfg.Source != "" {
			p.sourceDir = cfg.Source
		}
		if cfg.Output != "" {
			p.outputFile = cfg.Output
		}
		if cfg.Title != "" {
			p.composeOpts.Title = cfg.Title
		}
		if cfg.Subtitle != "" {
			p.composeOpts.Subtitle = cfg.Subtitle
		}
		if cfg.Author != "" {
			p.composeOpts.Author = cfg.Author
		}
	}
}

// themeOptionsFromConfig converts cfg.ThemeOptions (as loaded from
// go-pretty-pdf.yml or set by CLI flags) into theme.Options.
func themeOptionsFromConfig(cfg *config.Config) theme.Options {
	to := cfg.ThemeOptions
	return theme.Options{
		Colors: theme.Colors{
			Primary:    to.Colors.Primary,
			Accent:     to.Colors.Accent,
			Text:       to.Colors.Text,
			Muted:      to.Colors.Muted,
			Background: to.Colors.Background,
		},
		Fonts: theme.Fonts{
			Heading:       to.Fonts.Heading,
			Body:          to.Fonts.Body,
			Code:          to.Fonts.Code,
			GoogleImports: to.Fonts.GoogleFonts,
		},
		Sections: theme.Sections{
			Cover:       to.Sections.Cover,
			TOC:         to.Sections.TOC,
			PageNumbers: to.Sections.PageNumbers,
			Header:      to.Sections.Header,
		},
		Density:           theme.Density(to.Density),
		AllowNetworkFonts: to.AllowNetworkFonts,
	}
}

// WithConfigCSSAndTemplate resolves cfg.Theme (with cfg.ThemeOptions
// customization) and then loads CSS/template content from cfg.CSS/
// cfg.Template, which — being explicit file overrides — take priority over
// the theme and replace its CSS/template outright. Read/resolve failures
// are recorded as warnings and flushed to stderr by New() once all options
// have been applied, so ordering relative to WithVerbose does not matter.
func WithConfigCSSAndTemplate(cfg *config.Config) Option {
	return func(p *PDF) {
		if cfg.Theme != "" {
			cwd, _ := os.Getwd()
			css, sections, err := theme.ResolveByName(cfg.Theme, themeOptionsFromConfig(cfg), cwd)
			if err != nil {
				p.pendingWarnings = append(p.pendingWarnings, fmt.Sprintf("theme %q: %v", cfg.Theme, err))
			} else {
				p.composeOpts.CSS = css
				p.composeOpts.ShowCover = sections.Cover
				p.composeOpts.ShowTOC = sections.TOC
				p.renderOpts.PageNumbers = sections.PageNumbers
				p.renderOpts.ShowHeader = sections.Header
			}
		}
		if cfg.CSS != "" {
			data, err := os.ReadFile(cfg.CSS)
			if err == nil {
				p.composeOpts.CSS = string(data)
			} else {
				p.pendingWarnings = append(p.pendingWarnings, fmt.Sprintf("reading CSS file %s: %v", cfg.CSS, err))
			}
		}
		if cfg.Template != "" {
			data, err := os.ReadFile(cfg.Template)
			if err == nil {
				p.composeOpts.Template = string(data)
			} else {
				p.pendingWarnings = append(p.pendingWarnings, fmt.Sprintf("reading template file %s: %v", cfg.Template, err))
			}
		}
	}
}

// WithFullConfig applies every field of cfg: source/output/title/subtitle
// /author (via WithConfig), CSS/template/theme (via WithConfigCSSAndTemplate),
// variable substitution (cfg.Vars), and render settings (cfg.Render:
// timeout, paper size, margins, header title). Unlike WithConfig and
// WithConfigCSSAndTemplate, which only cover a subset of Config, this is
// the single option needed to fully apply a loaded go-pretty-pdf.yml.
func WithFullConfig(cfg *config.Config) Option {
	return func(p *PDF) {
		WithConfig(cfg)(p)
		WithConfigCSSAndTemplate(cfg)(p)

		if len(cfg.Vars) > 0 {
			p.parser.SetVars(cfg.Vars)
		}

		if cfg.Render.Timeout != "" {
			if d, err := time.ParseDuration(cfg.Render.Timeout); err == nil {
				p.renderOpts.Timeout = d
			}
		}

		if w, h, ok := config.ParsePaperSize(cfg.Render.Paper); ok {
			p.renderOpts.PaperWidth = w
			p.renderOpts.PaperHeight = h
		}

		defOpts := render.DefaultOptions()
		mt := config.ParseCSSUnit(cfg.Render.MarginTop)
		mb := config.ParseCSSUnit(cfg.Render.MarginBot)
		ml := config.ParseCSSUnit(cfg.Render.MarginLeft)
		mr := config.ParseCSSUnit(cfg.Render.MarginRight)
		if mt != 0 || mb != 0 || ml != 0 || mr != 0 {
			if mt == 0 {
				mt = defOpts.MarginTop
			}
			if mb == 0 {
				mb = defOpts.MarginBottom
			}
			if ml == 0 {
				ml = defOpts.MarginLeft
			}
			if mr == 0 {
				mr = defOpts.MarginRight
			}
			p.renderOpts.MarginTop = mt
			p.renderOpts.MarginBottom = mb
			p.renderOpts.MarginLeft = ml
			p.renderOpts.MarginRight = mr
		}

		if cfg.Render.HeaderTitle != "" {
			headerTitle := cfg.Render.HeaderTitle
			for k, v := range cfg.Vars {
				headerTitle = strings.ReplaceAll(headerTitle, "{{"+k+"}}", v)
			}
			p.renderOpts.HeaderTitle = headerTitle
			p.headerTitleSet = true
		}
	}
}

func New(opts ...Option) (*PDF, error) {
	p := &PDF{
		sourceDir:   "book",
		outputFile:  "out.pdf",
		parser:      mdx.NewParser(),
		composeOpts: compose.DefaultOptions(),
		renderOpts:  render.DefaultOptions(),
	}

	for _, o := range opts {
		o(p)
	}

	if !p.headerTitleSet {
		p.renderOpts.HeaderTitle = p.composeOpts.Title
	}

	if p.verbose {
		for _, w := range p.pendingWarnings {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
		}
	}
	p.pendingWarnings = nil

	return p, nil
}

func (p *PDF) Build(ctx context.Context) error {
	if p.verbose {
		fmt.Printf("Parsing MDX files in %s...\n", p.sourceDir)
	}

	docs, err := p.parser.ParseDir(p.sourceDir)
	if err != nil && len(docs) == 0 {
		return fmt.Errorf("parsing: %w", err)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: some files failed to parse: %v\n", err)
	}

	if p.verbose {
		fmt.Printf("Found %d document(s)\n", len(docs))
	}

	if p.validator != nil {
		p.logVerbose("Running validation...")
		var allErrs []mdx.ValidationError
		for _, doc := range docs {
			errs := p.validator.Validate(doc)
			allErrs = append(allErrs, errs...)
		}
		if len(allErrs) > 0 {
			for _, e := range allErrs {
				fmt.Printf("  - %v\n", e)
			}
			return fmt.Errorf("validation failed: %d error(s)", len(allErrs))
		}
		p.logVerbose("Validation passed")
	}

	p.logVerbose("Composing HTML...")
	html, err := compose.ComposeHTML(docs, p.composeOpts)
	if err != nil {
		return fmt.Errorf("composing HTML: %w", err)
	}

	p.logVerbose(fmt.Sprintf("Rendering PDF to %s...", p.outputFile))
	if err := render.RenderToPDF(html, p.outputFile, p.renderOpts); err != nil {
		return fmt.Errorf("rendering PDF: %w", err)
	}

	return nil
}

func (p *PDF) Validate(ctx context.Context) ([]mdx.ValidationError, error) {
	docs, err := p.parser.ParseDir(p.sourceDir)
	if err != nil {
		return nil, err
	}

	if p.validator == nil {
		return nil, fmt.Errorf("no validator configured")
	}

	var allErrs []mdx.ValidationError
	for _, doc := range docs {
		errs := p.validator.Validate(doc)
		allErrs = append(allErrs, errs...)
	}

	return allErrs, nil
}

func (p *PDF) ParseDir() ([]*mdx.Document, error) {
	return p.parser.ParseDir(p.sourceDir)
}

func (p *PDF) ValidateDoc(doc *mdx.Document) []mdx.ValidationError {
	if p.validator == nil {
		return nil
	}
	return p.validator.Validate(doc)
}

func (p *PDF) ValidateAll(docs []*mdx.Document) []mdx.ValidationError {
	if p.validator == nil {
		return nil
	}
	return p.validator.ValidateAll(docs)
}

func (p *PDF) ComposeHTML(docs []*mdx.Document) (string, error) {
	return compose.ComposeHTML(docs, p.composeOpts)
}

func (p *PDF) Render(html string) error {
	return render.RenderToPDF(html, p.outputFile, p.renderOpts)
}

func (p *PDF) logVerbose(msg string) {
	if p.verbose {
		fmt.Println(msg)
	}
}
