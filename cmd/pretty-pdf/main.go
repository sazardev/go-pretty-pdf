package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

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

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(versionCmd)

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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func scaffoldBook(targetDir string) error {
	bookDir := filepath.Join(targetDir, "book")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		return fmt.Errorf("creating book directory: %w", err)
	}

	indexContent := `---
id: "[1.0.0]"
title: "Getting Started"
subtitle: "A simple introduction"
tags:
  - example
  - intro
difficulty: "beginner"
status: complete
completeness: 100
depends_on: []
---

# Welcome to Your Book

This is the first chapter of your book, created by **go-pretty-pdf**.

## Overview

Write your content using **MDX** — an extended Markdown format with support
for custom components and variables.

### What is go-pretty-pdf?

go-pretty-pdf transforms your MDX files into a beautiful, print-ready PDF using headless Chrome.

## Next Steps

- Edit this file to start writing your book
- Run ` + "`pretty-pdf build --source ./book --out my-book.pdf`" + ` to generate a PDF
- Run ` + "`pretty-pdf check --source ./book`" + ` to validate your content
`

	indexPath := filepath.Join(bookDir, "index.mdx")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("writing index.mdx: %w", err)
	}

	configContent := `# go-pretty-pdf configuration
source: "book"
output: "out.pdf"
title: "My Book"
subtitle: ""
author: ""

theme: "default"

lint:
  requireFrontmatter:
    - "id"
    - "title"
  noDuplicateIDs: true
  maxHeadingDepth: 3

vars:
  BOOK_TITLE: "My Book"
  AUTHOR_NAME: "Author"

render:
  paper: "a4"
  timeout: "30s"
`

	configPath := filepath.Join(targetDir, ".pretty-pdf.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
