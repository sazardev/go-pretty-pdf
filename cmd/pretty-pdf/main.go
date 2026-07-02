package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	prettypdf "github.com/sazardev/go-pretty-pdf"
	"github.com/sazardev/go-pretty-pdf/config"
	"github.com/sazardev/go-pretty-pdf/mdx"
	"github.com/sazardev/go-pretty-pdf/theme"
)

//go:embed initassets/*
var initAssets embed.FS

var (
	cfgFile    string
	sourceDir  string
	outPath    string
	title      string
	subtitle   string
	author     string
	themeName  string
	cssPath    string
	tmplPath   string
	timeoutStr string
	verbose    bool
	strict     bool

	version = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pretty-pdf",
		Short: "go-pretty-pdf — Beautiful PDF generation from MDX",
		Long: `Transforms a directory of MDX files into a high-fidelity, print-ready PDF
using headless Chrome. Supports custom MDX components, themes,
variable substitution, built-in validation, and per-book configuration.`,
	}

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build PDF from MDX source directory",
		Long: `Parses all .mdx files in the source directory, validates them,
composes HTML, and renders a PDF via headless Chrome.

Reads go-pretty-pdf.yml in the current directory for configuration.
CLI flags override config values.`,
		RunE: runBuild,
	}
	buildCmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file (default: go-pretty-pdf.yml in cwd)")
	buildCmd.Flags().StringVar(&sourceDir, "source", "", "Path to the source MDX directory")
	buildCmd.Flags().StringVar(&outPath, "out", "", "Output PDF path")
	buildCmd.Flags().StringVar(&title, "title", "", "Document title")
	buildCmd.Flags().StringVar(&subtitle, "subtitle", "", "Document subtitle")
	buildCmd.Flags().StringVar(&author, "author", "", "Document author")
	buildCmd.Flags().StringVar(&themeName, "theme", "", "Theme name (default, minimal)")
	buildCmd.Flags().StringVar(&cssPath, "css", "", "Path to custom CSS file")
	buildCmd.Flags().StringVar(&tmplPath, "template", "", "Path to custom HTML template")
	buildCmd.Flags().StringVar(&timeoutStr, "timeout", "", "Render timeout (e.g. 30s, 1m)")
	buildCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Lint and validate MDX files without rendering",
		Long: `Parses all .mdx files and runs the built-in validator to check for
common issues: missing frontmatter, invalid IDs, duplicate IDs,
excessive heading depth, and more.

Use --strict to treat warnings as errors.`,
		RunE: runCheck,
	}
	checkCmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file")
	checkCmd.Flags().StringVar(&sourceDir, "source", "", "Path to the source MDX directory")
	checkCmd.Flags().BoolVar(&strict, "strict", false, "Treat all findings as errors")

	initCmd := &cobra.Command{
		Use:   "init [dir]",
		Short: "Scaffold a new book directory with example files",
		Long: `Creates a new book directory with example MDX files and a
go-pretty-pdf.yml configuration file.

If no directory is specified, "book" is used.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runInit,
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("pretty-pdf %s\n", version)
		},
	}

	rootCmd.AddCommand(buildCmd, checkCmd, initCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

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

	validator := mdx.NewDefaultValidator()
	if cfg.Lint.RequireFrontmatter != nil && len(cfg.Lint.RequireFrontmatter) > 0 {
		validator.RequireFrontmatter = cfg.Lint.RequireFrontmatter
	}
	validator.NoDuplicateIDs = cfg.Lint.NoDuplicateIDs
	validator.MaxHeadingDepth = cfg.Lint.MaxHeadingDepth
	if validator.MaxHeadingDepth == 0 {
		validator.MaxHeadingDepth = 3
	}
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

	mt := parseCSSUnit(renderCfg.MarginTop)
	mb := parseCSSUnit(renderCfg.MarginBot)
	ml := parseCSSUnit(renderCfg.MarginLeft)
	mr := parseCSSUnit(renderCfg.MarginRight)
	if mt != 0 || mb != 0 || ml != 0 || mr != 0 {
		if mt == 0 {
			mt = 0.8
		}
		if mb == 0 {
			mb = 0.8
		}
		if ml == 0 {
			ml = 0.6
		}
		if mr == 0 {
			mr = 0.6
		}
		opts = append(opts, prettypdf.WithRenderMargins(mt, mb, ml, mr))
	}

	if renderCfg.HeaderTitle != "" {
		opts = append(opts, prettypdf.WithHeaderTitle(renderCfg.HeaderTitle))
	}

	return opts
}

func runBuild(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	if verbose {
		if cfgFileFound() {
			fmt.Println("Using config file")
		}
		fmt.Printf("Source: %s\n", cfg.Source)
		fmt.Printf("Output: %s\n", cfg.Output)
		fmt.Printf("Title: %s\n", cfg.Title)
	}

	opts := buildOpts(cfg)
	pdf, err := prettypdf.New(opts...)
	if err != nil {
		return err
	}

	if err := pdf.Build(cmd.Context()); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("PDF generated: %s\n", cfg.Output)
	return nil
}

func runCheck(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	fmt.Printf("Checking MDX files in %s\n", cfg.Source)

	validator := mdx.NewDefaultValidator()
	if cfg.Lint.RequireFrontmatter != nil && len(cfg.Lint.RequireFrontmatter) > 0 {
		validator.RequireFrontmatter = cfg.Lint.RequireFrontmatter
	}
	validator.NoDuplicateIDs = cfg.Lint.NoDuplicateIDs
	validator.MaxHeadingDepth = cfg.Lint.MaxHeadingDepth
	if validator.MaxHeadingDepth == 0 {
		validator.MaxHeadingDepth = 3
	}

	parserOpts := []mdx.ParserOption{}
	if cfg.Vars != nil && len(cfg.Vars) > 0 {
		parserOpts = append(parserOpts, mdx.WithVars(cfg.Vars))
	}
	parser := mdx.NewParser(parserOpts...)

	docs, err := parser.ParseDir(cfg.Source)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	fmt.Printf("Found %d document(s)\n", len(docs))

	errs := validator.ValidateAll(docs)

	warnings := 0
	errors := 0
	for _, e := range errs {
		if e.Field == "content" && !strict {
			warnings++
			fmt.Printf("  [WARN]  %v\n", e)
		} else {
			errors++
			fmt.Printf("  [ERROR] %v\n", e)
		}
	}

	fmt.Printf("\n%d error(s), %d warning(s)\n", errors, warnings)

	if errors > 0 {
		return fmt.Errorf("validation failed with %d error(s)", errors)
	}
	if warnings > 0 && !strict {
		fmt.Println("Check passed with warnings.")
	} else {
		fmt.Println("All checks passed.")
	}

	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := "book"
	if len(args) > 0 {
		dir = args[0]
	}

	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("directory %q already exists", dir)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	entries, err := initAssets.ReadDir("initassets")
	if err != nil {
		return fmt.Errorf("reading embedded assets: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := initAssets.ReadFile("initassets/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading embedded %s: %w", entry.Name(), err)
		}
		dst := filepath.Join(dir, entry.Name())
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", dst, err)
		}
		fmt.Printf("  Created %s\n", dst)
	}

	fmt.Printf("\nBook scaffolded in %s/\n", dir)
	fmt.Println("Run: pretty-pdf build")
	return nil
}

func cfgFileFound() bool {
	if cfgFile != "" {
		return true
	}
	path, _ := config.FindConfig()
	return path != ""
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
