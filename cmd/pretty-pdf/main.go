package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf/output"
	"github.com/sazardev/go-pretty-pdf/version"
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
	noColor    bool
	quiet      bool
	jsonOutput bool
	initBare   bool
	servePort  int
)

var rootCmd = &cobra.Command{
	Use:   "pretty-pdf",
	Short: "Transform MDX files into beautiful, print-ready PDFs",
	Long: output.PrimaryStyle.Render(`
  go-pretty-pdf transforms a directory of MDX files into a print-ready PDF
  via headless Chrome. Documents are sorted by their [X.Y.Z] frontmatter ID.
  Supports custom components, themes, CSS overrides, and YAML configuration.
`) + "\n  " + output.MutedStyle.Render("https://github.com/sazardev/go-pretty-pdf"),
	SilenceUsage: true,
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a PDF from MDX source files",
	Long:  "Parse MDX files, validate them, compose HTML, and render to PDF via headless Chrome.",
	RunE:  runBuild,
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate MDX files",
	Long:  "Parse and validate all MDX files in the source directory without building a PDF.",
	RunE:  runCheck,
}

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Scaffold a new book project",
	Long: `Scaffold a new book project with a sample MDX file, configuration, and directory structure.
Run 'pretty-pdf init my-book' to create a new project in the 'my-book' directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for changes and rebuild automatically",
	Long:  "Watch the source directory for changes and rebuild the PDF on every file change.",
	RunE:  runWatch,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(output.PrimaryStyle.Render("go-pretty-pdf") + " " + output.MutedStyle.Render("v"+version.Version))
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Preview MDX as HTML in the browser",
	Long:  "Parse MDX files, compose HTML, and serve with live reload on file changes. No Chrome required.",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(serveCmd)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file")
	rootCmd.PersistentFlags().StringVar(&sourceDir, "source", "book", "source MDX directory")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress non-error output")

	buildCmd.Flags().StringVar(&outPath, "out", "out.pdf", "output PDF path")
	buildCmd.Flags().StringVar(&title, "title", "", "book title")
	buildCmd.Flags().StringVar(&subtitle, "subtitle", "", "book subtitle")
	buildCmd.Flags().StringVar(&author, "author", "", "book author")
	buildCmd.Flags().StringVar(&themeName, "theme", "default", "book theme (default, minimal)")
	buildCmd.Flags().StringVar(&cssPath, "css", "", "custom CSS file path")
	buildCmd.Flags().StringVar(&tmplPath, "template", "", "custom HTML template file path")
	buildCmd.Flags().StringVar(&timeoutStr, "timeout", "", "render timeout (e.g. 30s, 1m)")
	buildCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	checkCmd.Flags().BoolVar(&strict, "strict", false, "treat content warnings as errors")

	initCmd.Flags().BoolVar(&initBare, "bare", false, "non-interactive init with flags")
	initCmd.Flags().StringVar(&title, "title", "My Book", "book title (for --bare)")
	initCmd.Flags().StringVar(&author, "author", "go-pretty-pdf", "book author (for --bare)")
	initCmd.Flags().StringVar(&themeName, "theme", "default", "book theme (for --bare)")
	initCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	serveCmd.Flags().IntVar(&servePort, "port", 8080, "HTTP server port")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}


