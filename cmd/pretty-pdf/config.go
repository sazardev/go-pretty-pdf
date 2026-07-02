package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	prettypdf "github.com/sazardev/go-pretty-pdf"
	"github.com/sazardev/go-pretty-pdf/config"
	"github.com/sazardev/go-pretty-pdf/mdx"
	"github.com/sazardev/go-pretty-pdf/render"
	"github.com/sazardev/go-pretty-pdf/theme"
)

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	cfg := config.Default()

	configPath := cfgFile
	if configPath == "" {
		found, err := config.FindConfig()
		if err != nil {
			return nil, err
		}
		configPath = found
	}

	if configPath != "" {
		loaded, err := config.Load(configPath)
		if err != nil {
			return nil, fmt.Errorf("loading config: %w", err)
		}
		cfg = loaded

		configDir := filepath.Dir(configPath)
		if cfg.CSS != "" && !filepath.IsAbs(cfg.CSS) {
			cfg.CSS = filepath.Join(configDir, cfg.CSS)
		}
		if cfg.Template != "" && !filepath.IsAbs(cfg.Template) {
			cfg.Template = filepath.Join(configDir, cfg.Template)
		}
	}

	if cmd.Flags().Changed("source") {
		cfg.Source = sourceDir
	}
	if cmd.Flags().Changed("out") {
		cfg.Output = outPath
	}
	if cmd.Flags().Changed("title") {
		cfg.Title = title
	}
	if cmd.Flags().Changed("subtitle") {
		cfg.Subtitle = subtitle
	}
	if cmd.Flags().Changed("author") {
		cfg.Author = author
	}
	if cmd.Flags().Changed("theme") {
		cfg.Theme = themeName
	}
	if cmd.Flags().Changed("css") {
		cfg.CSS = cssPath
		if cfg.CSS != "" && !filepath.IsAbs(cfg.CSS) {
			abs, err := filepath.Abs(cfg.CSS)
			if err == nil {
				cfg.CSS = abs
			}
		}
	}
	if cmd.Flags().Changed("template") {
		cfg.Template = tmplPath
		if cfg.Template != "" && !filepath.IsAbs(cfg.Template) {
			abs, err := filepath.Abs(cfg.Template)
			if err == nil {
				cfg.Template = abs
			}
		}
	}
	if cmd.Flags().Changed("timeout") {
		cfg.Render.Timeout = timeoutStr
	}

	if verbose {
		if cfg.CSS != "" {
			if _, err := os.Stat(cfg.CSS); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: CSS file not found: %s\n", cfg.CSS)
			}
		}
		if cfg.Template != "" {
			if _, err := os.Stat(cfg.Template); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: template file not found: %s\n", cfg.Template)
			}
		}
	}

	return cfg, nil
}

func buildOpts(cfg *config.Config) []prettypdf.Option {
	opts := []prettypdf.Option{
		prettypdf.WithVerbose(verbose),
		prettypdf.WithConfig(cfg),
		prettypdf.WithConfigCSSAndTemplate(cfg),
	}

	if cfg.Vars != nil && len(cfg.Vars) > 0 {
		opts = append(opts, prettypdf.WithVars(cfg.Vars))
	}

	validator := validatorFromConfig(cfg)
	opts = append(opts, prettypdf.WithValidator(validator))

	if cfg.Theme != "" && cfg.Theme != "default" {
		switch cfg.Theme {
		case "minimal":
			opts = append(opts, prettypdf.WithTheme(theme.Minimal))
		}
	}

	if cfg.Render.Timeout != "" {
		d, err := time.ParseDuration(cfg.Render.Timeout)
		if err == nil {
			opts = append(opts, prettypdf.WithTimeout(d))
		}
	}

	renderCfg := cfg.Render
	if renderCfg.Paper != "" {
		switch strings.ToLower(renderCfg.Paper) {
		case "letter":
			opts = append(opts, prettypdf.WithPaperSize(8.5, 11))
		case "legal":
			opts = append(opts, prettypdf.WithPaperSize(8.5, 14))
		case "a4":
			opts = append(opts, prettypdf.WithPaperSize(8.27, 11.69))
		}
	}

	defOpts := render.DefaultOptions()
	mt := parseCSSUnit(renderCfg.MarginTop)
	mb := parseCSSUnit(renderCfg.MarginBot)
	ml := parseCSSUnit(renderCfg.MarginLeft)
	mr := parseCSSUnit(renderCfg.MarginRight)
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
		opts = append(opts, prettypdf.WithRenderMargins(mt, mb, ml, mr))
	}

	if renderCfg.HeaderTitle != "" {
		opts = append(opts, prettypdf.WithHeaderTitle(renderCfg.HeaderTitle))
	}

	return opts
}

func validatorFromConfig(cfg *config.Config) *mdx.DefaultValidator {
	v := mdx.NewDefaultValidator()
	if cfg.Lint.RequireFrontmatter != nil && len(cfg.Lint.RequireFrontmatter) > 0 {
		v.RequireFrontmatter = cfg.Lint.RequireFrontmatter
	}
	v.NoDuplicateIDs = cfg.Lint.NoDuplicateIDs
	v.MaxHeadingDepth = cfg.Lint.MaxHeadingDepth
	if v.MaxHeadingDepth == 0 {
		v.MaxHeadingDepth = 3
	}
	return v
}

func parserFromConfig(cfg *config.Config) *mdx.Parser {
	parserOpts := []mdx.ParserOption{}
	if cfg.Vars != nil && len(cfg.Vars) > 0 {
		parserOpts = append(parserOpts, mdx.WithVars(cfg.Vars))
	}
	return mdx.NewParser(parserOpts...)
}

func parseCSSUnit(s string) float64 {
	if s == "" {
		return 0
	}
	s = strings.TrimSpace(s)

	var value float64
	var unit string
	fmt.Sscanf(s, "%f%s", &value, &unit)

	switch strings.ToLower(unit) {
	case "in":
		return value
	case "mm":
		return value / 25.4
	case "cm":
		return value / 2.54
	case "pt":
		return value / 72.0
	case "px":
		return value / 96.0
	default:
		return 0
	}
}

func cfgFileFound() bool {
	if cfgFile != "" {
		return true
	}
	path, _ := config.FindConfig()
	return path != ""
}
