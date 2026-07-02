package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	prettypdf "github.com/sazardev/go-pretty-pdf"
	"github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf/output"
	"github.com/sazardev/go-pretty-pdf/config"
	"github.com/sazardev/go-pretty-pdf/render"
	"github.com/sazardev/go-pretty-pdf/version"
)

func runBuild(cmd *cobra.Command, args []string) error {
	if noColor {
		output.NoColor()
	}

	if jsonOutput {
		return runBuildJSON(cmd)
	}

	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	output.PrintBanner(version.Version)

	preflightResults := runPreFlight(cfg)
	output.PrintPreFlight(preflightResults)

	failed := false
	for _, r := range preflightResults {
		if !r.Passed && !r.Warning {
			failed = true
			break
		}
	}
	if failed {
		return fmt.Errorf("pre-flight checks failed")
	}

	if !quiet {
		if cfgFileFound() {
			fmt.Println("  " + output.MutedStyle.Render("Config: "+cfgFile))
		}
		fmt.Println("  " + output.KeyValue("Source", cfg.Source))
		fmt.Println("  " + output.KeyValue("Output", cfg.Output))
		fmt.Println()
	}

	pipeline := output.NewPipelineProgress(
		"Parsing MDX files...",
		"Running validation...",
		"Composing HTML...",
		"Rendering PDF...",
	)

	startTime := time.Now()

	pipeline.Start("Parsing MDX files...")
	opts := buildOpts(cfg)
	pdf, err := prettypdf.New(opts...)
	if err != nil {
		pipeline.Fail("Parsing MDX files...", err.Error())
		return fmt.Errorf("initializing: %w", err)
	}

	docs, err := pdf.ParseDir()
	if err != nil && len(docs) == 0 {
		pipeline.Fail("Parsing MDX files...", err.Error())
		return fmt.Errorf("parsing: %w", err)
	}
	if err != nil {
		fmt.Printf("    %s\n", output.Warn(fmt.Sprintf("Some files failed to parse: %v", err)))
	}
	pipeline.Done("Parsing MDX files...")

	if verbose {
		fmt.Println("    " + output.MutedStyle.Render(fmt.Sprintf("%d document(s) found", len(docs))))
	}

	pipeline.Start("Running validation...")
	allErrs := pdf.ValidateAll(docs)
	if len(allErrs) > 0 {
		for _, e := range allErrs {
			fmt.Printf("    %s\n", e)
		}
		pipeline.Fail("Running validation...", fmt.Sprintf("%d error(s)", len(allErrs)))
		return fmt.Errorf("validation failed: %d error(s)", len(allErrs))
	}
	pipeline.Done("Running validation...")

	pipeline.Start("Composing HTML...")
	html, err := pdf.ComposeHTML(docs)
	if err != nil {
		pipeline.Fail("Composing HTML...", err.Error())
		return fmt.Errorf("composing HTML: %w", err)
	}
	pipeline.Done("Composing HTML...")

	pipeline.Start("Rendering PDF...")
	if err := pdf.Render(html); err != nil {
		pipeline.Fail("Rendering PDF...", err.Error())
		return fmt.Errorf("rendering PDF: %w", err)
	}
	pipeline.Done("Rendering PDF...")

	elapsed := time.Since(startTime)

	fileSize := "unknown"
	if info, err := os.Stat(cfg.Output); err == nil {
		fileSize = formatBytes(info.Size())
	}

	themeLabel := cfg.Theme
	if themeLabel == "" {
		themeLabel = "default"
	}

	output.PrintBuildSummary(output.BuildStats{
		Documents: len(docs),
		Output:    cfg.Output,
		FileSize:  fileSize,
		Duration:  elapsed,
		Theme:     themeLabel,
		Warnings:  0,
	})

	return nil
}

func runBuildJSON(cmd *cobra.Command) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	opts := buildOpts(cfg)
	pdf, err := prettypdf.New(opts...)
	if err != nil {
		return err
	}

	startTime := time.Now()

	docs, err := pdf.ParseDir()
	if err != nil && len(docs) == 0 {
		return fmt.Errorf("parsing: %w", err)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: some files failed to parse: %v\n", err)
	}

	allErrs := pdf.ValidateAll(docs)
	if len(allErrs) > 0 {
		for _, e := range allErrs {
			fmt.Fprintf(os.Stderr, "%s\n", e)
		}
		return fmt.Errorf("validation failed: %d error(s)", len(allErrs))
	}

	html, err := pdf.ComposeHTML(docs)
	if err != nil {
		return fmt.Errorf("composing HTML: %w", err)
	}

	if err := pdf.Render(html); err != nil {
		return fmt.Errorf("rendering PDF: %w", err)
	}

	elapsed := time.Since(startTime)
	fileSize := "0"
	if info, err := os.Stat(cfg.Output); err == nil {
		fileSize = fmt.Sprintf("%d", info.Size())
	}

	fmt.Printf(`{"documents":%d,"output":"%s","size_bytes":%s,"duration_ms":%d,"theme":"%s"}`+"\n",
		len(docs), filepath.ToSlash(cfg.Output), fileSize, elapsed.Milliseconds(), cfg.Theme)

	return nil
}

func runPreFlight(cfg *config.Config) []output.PreFlightResult {
	var results []output.PreFlightResult

	if err := render.CheckChromeAvailable(); err != nil {
		results = append(results, output.PreFlightResult{
			Name:    "Chrome/Chromium available",
			Passed:  false,
			Message: fmt.Sprintf("%v — Install Chrome or Chromium to render PDFs", err),
		})
	} else {
		results = append(results, output.PreFlightResult{
			Name:   "Chrome/Chromium available",
			Passed: true,
		})
	}

	srcInfo, err := os.Stat(cfg.Source)
	if err != nil {
		results = append(results, output.PreFlightResult{
			Name:    "Source directory exists",
			Passed:  false,
			Message: fmt.Sprintf("%s: %v", cfg.Source, err),
		})
	} else if !srcInfo.IsDir() {
		results = append(results, output.PreFlightResult{
			Name:    "Source is a directory",
			Passed:  false,
			Message: fmt.Sprintf("%s is not a directory", cfg.Source),
		})
	} else {
		mdxCount := countMDXFiles(cfg.Source)
		if mdxCount == 0 {
			results = append(results, output.PreFlightResult{
				Name:    "MDX files found",
				Passed:  false,
				Warning: true,
				Message: fmt.Sprintf("no .mdx files in %s", cfg.Source),
			})
		} else {
			results = append(results, output.PreFlightResult{
				Name:   fmt.Sprintf("Source directory (%d .mdx files)", mdxCount),
				Passed: true,
			})
		}
	}

	outDir := filepath.Dir(cfg.Output)
	if outDir != "." {
		if _, err := os.Stat(outDir); os.IsNotExist(err) {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				results = append(results, output.PreFlightResult{
					Name:    "Output directory writable",
					Passed:  false,
					Message: fmt.Sprintf("cannot create %s: %v", outDir, err),
				})
			} else {
				results = append(results, output.PreFlightResult{
					Name:   "Output directory writable",
					Passed: true,
				})
			}
		} else {
			results = append(results, output.PreFlightResult{
				Name:   "Output directory writable",
				Passed: true,
			})
		}
	} else {
		results = append(results, output.PreFlightResult{
			Name:   "Output path writable",
			Passed: true,
		})
	}

	if cfg.CSS != "" {
		if _, err := os.Stat(cfg.CSS); err != nil {
			results = append(results, output.PreFlightResult{
				Name:    "CSS file exists",
				Passed:  false,
				Warning: true,
				Message: fmt.Sprintf("%s: %v", cfg.CSS, err),
			})
		} else {
			results = append(results, output.PreFlightResult{
				Name:   "CSS file exists",
				Passed: true,
			})
		}
	}

	if cfg.Template != "" {
		if _, err := os.Stat(cfg.Template); err != nil {
			results = append(results, output.PreFlightResult{
				Name:    "Template file exists",
				Passed:  false,
				Warning: true,
				Message: fmt.Sprintf("%s: %v", cfg.Template, err),
			})
		} else {
			results = append(results, output.PreFlightResult{
				Name:   "Template file exists",
				Passed: true,
			})
		}
	}

	return results
}

func countMDXFiles(dir string) int {
	count := 0
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".mdx") {
			count++
		}
		return nil
	})
	return count
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
