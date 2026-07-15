package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	prettypdf "github.com/sazardev/go-pretty-pdf"
	"github.com/sazardev/go-pretty-pdf/chromemgr"
	"github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf/output"
	"github.com/sazardev/go-pretty-pdf/config"
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

	chromeExecPath, chromeErr := resolveChromePath()

	preflightResults := runPreFlight(cfg, chromeExecPath, chromeErr)
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
	opts := buildOpts(cfg, chromeExecPath)
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
		themeLabel = defaultTheme
	}

	audit := pdf.LastAudit()

	output.PrintBuildSummary(output.BuildStats{
		Documents: len(docs),
		Output:    cfg.Output,
		FileSize:  fileSize,
		Duration:  elapsed,
		Theme:     themeLabel,
		Warnings:  len(audit.Issues),
	})

	if audit.HasIssues() {
		fmt.Println()
		fmt.Println("  " + output.MutedStyle.Render("PDF quality checks flagged:"))
		for _, issue := range audit.Issues {
			fmt.Printf("    %s %s\n", output.Warn("["+issue.Check+"]"), issue.Message)
		}
	}

	return nil
}

type buildJSONWarning struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

type buildJSONResult struct {
	Documents  int                `json:"documents"`
	Output     string             `json:"output"`
	SizeBytes  int64              `json:"size_bytes"`
	DurationMs int64              `json:"duration_ms"`
	Theme      string             `json:"theme"`
	Warnings   []buildJSONWarning `json:"warnings"`
}

func runBuildJSON(cmd *cobra.Command) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	chromeExecPath, err := resolveChromePath()
	if err != nil {
		return fmt.Errorf("resolving Chrome: %w", err)
	}

	opts := buildOpts(cfg, chromeExecPath)
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

	if err = pdf.Render(html); err != nil {
		return fmt.Errorf("rendering PDF: %w", err)
	}

	elapsed := time.Since(startTime)
	var fileSize int64
	if info, statErr := os.Stat(cfg.Output); statErr == nil {
		fileSize = info.Size()
	}

	audit := pdf.LastAudit()
	warnings := make([]buildJSONWarning, 0, len(audit.Issues))
	for _, issue := range audit.Issues {
		warnings = append(warnings, buildJSONWarning{Check: issue.Check, Message: issue.Message})
	}

	out, err := json.Marshal(buildJSONResult{
		Documents:  len(docs),
		Output:     filepath.ToSlash(cfg.Output),
		SizeBytes:  fileSize,
		DurationMs: elapsed.Milliseconds(),
		Theme:      cfg.Theme,
		Warnings:   warnings,
	})
	if err != nil {
		return fmt.Errorf("encoding JSON result: %w", err)
	}
	fmt.Println(string(out))

	return nil
}

// resolveChromePath finds (or, as a last resort, downloads) a usable
// Chrome/Chromium binary so users are never required to install one by
// hand. See the chromemgr package for the full resolution order.
//
// A returned empty path with a nil error means a system install was found
// and chromedp should use its own default discovery; render.Options
// treats "" the same way.
func resolveChromePath() (string, error) {
	ctx := context.Background()

	if chromePath != "" {
		return chromemgr.EnsureChrome(ctx, chromePath, nil)
	}
	if chromemgr.SystemChromeAvailable(ctx) {
		return "", nil
	}

	spinner := output.StartSpinner("Chrome/Chromium not found — downloading a headless build (one-time)...")
	path, err := chromemgr.EnsureChrome(ctx, "", nil)
	if err != nil {
		spinner.Fail(err.Error())
		return "", err
	}
	spinner.Done("Chrome downloaded and cached at " + path)
	return path, nil
}

func runPreFlight(cfg *config.Config, chromeExecPath string, chromeErr error) []output.PreFlightResult {
	var results []output.PreFlightResult

	if chromeErr != nil {
		results = append(results, output.PreFlightResult{
			Name:    "Chrome/Chromium available",
			Passed:  false,
			Message: chromeErr.Error(),
		})
	} else {
		name := "Chrome/Chromium available"
		switch {
		case chromePath != "":
			name = "Chrome/Chromium available (--chrome-path)"
		case chromeExecPath != "":
			name = "Chrome/Chromium available (auto-downloaded)"
		}
		results = append(results, output.PreFlightResult{
			Name:   name,
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
					Name:    outputDirWritable,
					Passed:  false,
					Message: fmt.Sprintf("cannot create %s: %v", outDir, err),
				})
			} else {
				results = append(results, output.PreFlightResult{
					Name:   outputDirWritable,
					Passed: true,
				})
			}
		} else {
			results = append(results, output.PreFlightResult{
				Name:   outputDirWritable,
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

	if cfg.Render.CoverImage != "" {
		results = append(results, coverImagePreFlight(cfg.Render.CoverImage))
	}

	return results
}

func coverImagePreFlight(path string) output.PreFlightResult {
	info, err := os.Stat(path)
	if err != nil {
		return output.PreFlightResult{
			Name:    coverImageExists,
			Passed:  false,
			Message: fmt.Sprintf("%s: %v", path, err),
		}
	}
	if info.IsDir() {
		return output.PreFlightResult{
			Name:    coverImageExists,
			Passed:  false,
			Message: fmt.Sprintf("%s is a directory, not an image file", path),
		}
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg":
	default:
		return output.PreFlightResult{
			Name:    coverImageExists,
			Passed:  false,
			Message: fmt.Sprintf("%s: unsupported format (expected .png, .jpg, or .jpeg)", path),
		}
	}
	return output.PreFlightResult{
		Name:   coverImageExists,
		Passed: true,
	}
}

func countMDXFiles(dir string) int {
	count := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
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
