package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf/output"
)

func runInit(cmd *cobra.Command, args []string) error {
	if noColor {
		output.NoColor()
	}

	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	if initBare {
		return runInitBare(targetDir, title, author, themeName, sourceDir, jsonOutput)
	}

	if jsonOutput {
		return runInitBare(targetDir, "My Book", "go-pretty-pdf", "default", "book", true)
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

func runInitBare(targetDir, bookTitle, authorName, themeChoice, srcDir string, json bool) error {
	if err := scaffoldWithConfig(targetDir, bookTitle, authorName, themeChoice, srcDir); err != nil {
		return err
	}

	if json {
		fmt.Printf(`{"directory":"%s","book_title":"%s","author":"%s","theme":"%s","source":"%s"}`+"\n",
			targetDir, bookTitle, authorName, themeChoice, srcDir)
	} else {
		absTarget, _ := filepath.Abs(targetDir)
		fmt.Println(output.Success(fmt.Sprintf("Project created at %s", absTarget)))
		fmt.Println("  " + output.MutedStyle.Render("Run:") +
			" " + output.CodeStyle.Render(fmt.Sprintf("cd %s && pretty-pdf build", targetDir)))
	}
	return nil
}

func scaffoldWithConfig(targetDir, bookTitle, authorName, themeChoice, sourceDir string) error {
	bookDir := filepath.Join(targetDir, sourceDir)
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		return fmt.Errorf("creating book directory: %w", err)
	}

	replacer := strings.NewReplacer(
		"{{BOOK_TITLE}}", bookTitle,
		"{{AUTHOR_NAME}}", authorName,
		"{{THEME}}", themeChoice,
		"{{SOURCE_DIR}}", sourceDir,
	)

	assets := []string{
		"initassets/[1.0.0]-introduction.mdx",
		"initassets/[1.1.0]-getting-started.mdx",
		"initassets/[1.1.1]-installation.mdx",
	}
	for _, asset := range assets {
		data, err := initAssets.ReadFile(asset)
		if err != nil {
			return fmt.Errorf("reading embedded %s: %w", asset, err)
		}
		content := replacer.Replace(string(data))
		outPath := filepath.Join(bookDir, filepath.Base(asset))
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
	}

	configData, err := initAssets.ReadFile("initassets/go-pretty-pdf.yml")
	if err != nil {
		return fmt.Errorf("reading embedded config: %w", err)
	}
	configContent := replacer.Replace(string(configData))
	configPath := filepath.Join(targetDir, "go-pretty-pdf.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
