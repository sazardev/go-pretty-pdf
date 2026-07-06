package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	prettypdf "github.com/sazardev/go-pretty-pdf"
	"github.com/sazardev/go-pretty-pdf/config"
	"github.com/sazardev/go-pretty-pdf/mdx"
)

const (
	defaultTheme      = "default"
	outputDirWritable = "Output directory writable"
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
	return []prettypdf.Option{
		prettypdf.WithVerbose(verbose),
		prettypdf.WithFullConfig(cfg),
		prettypdf.WithValidator(validatorFromConfig(cfg)),
	}
}

func validatorFromConfig(cfg *config.Config) *mdx.DefaultValidator {
	v := mdx.NewDefaultValidator()
	if len(cfg.Lint.RequireFrontmatter) > 0 {
		v.RequireFrontmatter = cfg.Lint.RequireFrontmatter
	}
	v.NoDuplicateIDs = cfg.Lint.NoDuplicateIDs
	v.MaxHeadingDepth = cfg.Lint.MaxHeadingDepth
	if v.MaxHeadingDepth == 0 {
		v.MaxHeadingDepth = 5
	}
	return v
}

func parserFromConfig(cfg *config.Config) *mdx.Parser {
	parserOpts := []mdx.ParserOption{}
	if len(cfg.Vars) > 0 {
		parserOpts = append(parserOpts, mdx.WithVars(cfg.Vars))
	}
	return mdx.NewParser(parserOpts...)
}

func cfgFileFound() bool {
	if cfgFile != "" {
		return true
	}
	path, _ := config.FindConfig()
	return path != ""
}
