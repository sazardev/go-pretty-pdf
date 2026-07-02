package prettypdf

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sazardev/go-pretty-pdf/compose"
	"github.com/sazardev/go-pretty-pdf/config"
	"github.com/sazardev/go-pretty-pdf/mdx"
	"github.com/sazardev/go-pretty-pdf/render"
	"github.com/sazardev/go-pretty-pdf/theme"
)

type PDF struct {
	sourceDir   string
	outputFile  string
	parser      *mdx.Parser
	composeOpts compose.Options
	renderOpts  render.Options
	validator   mdx.Validator
	verbose     bool
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

func WithTheme(t theme.Theme) Option {
	return func(p *PDF) {
		if t.CSS != "" {
			p.composeOpts.CSS = t.CSS
		}
		if t.Template != "" {
			p.composeOpts.Template = t.Template
		}
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

func WithHeaderTitle(title string) Option {
	return func(p *PDF) {
		p.renderOpts.HeaderTitle = title
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

func WithConfigCSSAndTemplate(cfg *config.Config) Option {
	return func(p *PDF) {
		if cfg.CSS != "" {
			data, err := os.ReadFile(cfg.CSS)
			if err == nil {
				p.composeOpts.CSS = string(data)
			} else if p.verbose {
				fmt.Fprintf(os.Stderr, "Warning: reading CSS file %s: %v\n", cfg.CSS, err)
			}
		}
		if cfg.Template != "" {
			data, err := os.ReadFile(cfg.Template)
			if err == nil {
				p.composeOpts.Template = string(data)
			} else if p.verbose {
				fmt.Fprintf(os.Stderr, "Warning: reading template file %s: %v\n", cfg.Template, err)
			}
		}
		if cfg.Theme != "" {
			switch cfg.Theme {
			case "minimal":
				p.composeOpts.CSS = theme.Minimal.CSS
			case "default":
			}
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

	p.renderOpts.HeaderTitle = p.composeOpts.Title

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
