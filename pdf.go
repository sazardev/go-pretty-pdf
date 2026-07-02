package prettypdf

import (
	"context"
	"fmt"
	"time"

	"github.com/sazardev/go-pretty-pdf/compose"
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
		p.parser = mdx.NewParser(mdx.WithComponent(name, handler))
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

func New(opts ...Option) (*PDF, error) {
	p := &PDF{
		sourceDir:   ".",
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
	docs, err := p.parser.ParseDir(p.sourceDir)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	if p.validator != nil {
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
	}

	html, err := compose.ComposeHTML(docs, p.composeOpts)
	if err != nil {
		return fmt.Errorf("composing HTML: %w", err)
	}

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
