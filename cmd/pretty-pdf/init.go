package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf/output"
	"github.com/sazardev/go-pretty-pdf/version"
)

func runInit(cmd *cobra.Command, args []string) error {
	if noColor {
		output.NoColor()
	}

	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	if !quiet {
		output.PrintBanner(version.Version)
	}

	if jsonOutput {
		return runInitBare(targetDir)
	}

	var (
		bookTitle   string
		authorName  string
		themeChoice string
		sourceDir   string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(output.HeadingStyle.Render("Book Title")).
				Description(output.MutedStyle.Render("The main title of your book")).
				Placeholder("My Book").
				Value(&bookTitle),

			huh.NewInput().
				Title(output.HeadingStyle.Render("Author")).
				Description(output.MutedStyle.Render("The author's name")).
				Placeholder("go-pretty-pdf").
				Value(&authorName),

			huh.NewSelect[string]().
				Title(output.HeadingStyle.Render("Theme")).
				Description(output.MutedStyle.Render("Visual theme for your PDF")).
				Options(
					huh.NewOption("Default — clean, professional look", "default"),
					huh.NewOption("Minimal — stripped down, no extras", "minimal"),
				).
				Value(&themeChoice),

			huh.NewInput().
				Title(output.HeadingStyle.Render("Source Directory")).
				Description(output.MutedStyle.Render("Where your MDX files will live")).
				Placeholder("book").
				Value(&sourceDir),

			huh.NewConfirm().
				Title(output.HeadingStyle.Render("Create Project?")).
				Description(output.MutedStyle.Render(fmt.Sprintf("Will create %s with .pretty-pdf.yaml", targetDir))).
				Affirmative("Create!").
				Negative("Cancel"),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return fmt.Errorf("form cancelled: %w", err)
	}

	if bookTitle == "" {
		bookTitle = "My Book"
	}
	if authorName == "" {
		authorName = "go-pretty-pdf"
	}
	if themeChoice == "" {
		themeChoice = "default"
	}
	if sourceDir == "" {
		sourceDir = "book"
	}

	if !quiet {
		fmt.Println()
		spinner := output.StartSpinner("Scaffolding project...")
		if err := scaffoldWithConfig(targetDir, bookTitle, authorName, themeChoice, sourceDir); err != nil {
			spinner.Fail(err.Error())
			return err
		}
		spinner.Done("Project scaffolded!")
		fmt.Println()
	} else {
		if err := scaffoldWithConfig(targetDir, bookTitle, authorName, themeChoice, sourceDir); err != nil {
			return err
		}
	}

	absTarget, _ := filepath.Abs(targetDir)
	fmt.Println(output.Success(fmt.Sprintf("Project created at %s", absTarget)))
	fmt.Println("  " + output.MutedStyle.Render("Run:") +
		" " + output.CodeStyle.Render(fmt.Sprintf("cd %s && pretty-pdf build", targetDir)))

	return nil
}

func runInitBare(targetDir string) error {
	sourceDir := "book"
	bookTitle := "My Book"
	authorName := "go-pretty-pdf"
	themeChoice := "default"

	if err := scaffoldWithConfig(targetDir, bookTitle, authorName, themeChoice, sourceDir); err != nil {
		return err
	}

	fmt.Printf(`{"directory":"%s","book_title":"%s","author":"%s","theme":"%s","source":"%s"}`+"\n",
		targetDir, bookTitle, authorName, themeChoice, sourceDir)
	return nil
}

func scaffoldWithConfig(targetDir, bookTitle, authorName, themeChoice, sourceDir string) error {
	bookDir := filepath.Join(targetDir, sourceDir)
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		return fmt.Errorf("creating book directory: %w", err)
	}

	indexContent := fmt.Sprintf(`---
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

# Welcome to %s

This is the first chapter, created by **go-pretty-pdf**.

## Overview

Write your content using **MDX** — an extended Markdown format with support
for custom components and variables.

## Next Steps

- Edit this file to start writing your book
- Run ` + "`pretty-pdf build --source ./%s --out %s.pdf`" + ` to generate a PDF
- Run ` + "`pretty-pdf check --source ./%s`" + ` to validate your content
`, bookTitle, sourceDir, bookTitle, sourceDir)

	indexPath := filepath.Join(bookDir, "index.mdx")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("writing index.mdx: %w", err)
	}

	configContent := fmt.Sprintf(`# go-pretty-pdf configuration
source: "%s"
output: "out.pdf"
title: "%s"
subtitle: ""
author: "%s"

theme: "%s"

lint:
  requireFrontmatter:
    - "id"
    - "title"
  noDuplicateIDs: true
  maxHeadingDepth: 3

vars:
  BOOK_TITLE: "%s"
  AUTHOR_NAME: "%s"

render:
  paper: "a4"
  timeout: "30s"
`, sourceDir, bookTitle, authorName, themeChoice, bookTitle, authorName)

	configPath := filepath.Join(targetDir, ".pretty-pdf.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
