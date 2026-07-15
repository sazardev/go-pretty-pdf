package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf/output"
	"github.com/sazardev/go-pretty-pdf/epub"
	"github.com/sazardev/go-pretty-pdf/mdx"
)

func runEpub(cmd *cobra.Command, args []string) error {
	if noColor {
		output.NoColor()
	}

	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	if !quiet {
		fmt.Println("  " + output.KeyValue("Source", cfg.Source))
		fmt.Println("  " + output.KeyValue("Output", epubOutPath))
		fmt.Println()
	}

	parser := parserFromConfig(cfg)
	validator := validatorFromConfig(cfg)

	spinner := output.StartSpinner("Parsing MDX files...")
	docs, err := parser.ParseDir(cfg.Source)
	if err != nil && len(docs) == 0 {
		spinner.Fail(err.Error())
		return fmt.Errorf("parsing: %w", err)
	}
	spinner.Done(fmt.Sprintf("Found %d document(s)", len(docs)))
	if err != nil {
		fmt.Printf("    %s\n", output.Warn(fmt.Sprintf("Some files failed to parse: %v", err)))
	}

	errs := validator.ValidateAll(docs)
	errorCount := 0
	for _, e := range errs {
		if e.Field != mdx.ContentField {
			errorCount++
			fmt.Printf("  %v\n", e)
		}
	}
	if errorCount > 0 {
		return fmt.Errorf("validation failed with %d error(s)", errorCount)
	}

	opts := epub.DefaultOptions()
	if cfg.Title != "" {
		opts.Title = cfg.Title
	}
	opts.Subtitle = cfg.Subtitle
	if cfg.Author != "" {
		opts.Author = cfg.Author
	}
	if epubLanguage != "" {
		opts.Language = epubLanguage
	}
	opts.CoverImage = cfg.Render.CoverImage

	writeSpinner := output.StartSpinner("Writing EPUB...")
	if err := epub.Write(docs, opts, epubOutPath); err != nil {
		writeSpinner.Fail(err.Error())
		return fmt.Errorf("writing EPUB: %w", err)
	}

	size := "unknown"
	if info, statErr := os.Stat(epubOutPath); statErr == nil {
		size = formatBytes(info.Size())
	}
	writeSpinner.Done(fmt.Sprintf("Wrote %s (%s)", epubOutPath, size))

	return nil
}
