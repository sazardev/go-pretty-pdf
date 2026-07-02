package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/sazardev/go-pretty-pdf/mdx"
)

type BuildStats struct {
	Documents int
	Output    string
	FileSize  string
	Duration  time.Duration
	Theme     string
	Warnings  int
}

func PrintBuildSummary(stats BuildStats) {
	var lines []string
	lines = append(lines, KeyValue("Documents", NumberStyle.Render(fmt.Sprintf("%d", stats.Documents))))
	lines = append(lines, KeyValue("Output", fmt.Sprintf("%s (%s)", stats.Output, stats.FileSize)))
	lines = append(lines, KeyValue("Duration", stats.Duration.Round(time.Millisecond).String()))
	if stats.Theme != "" {
		lines = append(lines, KeyValue("Theme", stats.Theme))
	}
	if stats.Warnings > 0 {
		lines = append(lines, KeyValue("Warnings", WarningStyle.Render(fmt.Sprintf("%d", stats.Warnings))))
	} else {
		lines = append(lines, KeyValue("Warnings", "0"))
	}

	fmt.Println()
	fmt.Println(Panel("Build Complete!", strings.Join(lines, "\n")))
}

func PrintValidationSummary(errs []mdx.ValidationError, warnings int) {
	if len(errs) == 0 && warnings == 0 {
		fmt.Println(Success("All checks passed!"))
		return
	}

	errors := len(errs) - warnings

	fmt.Println()

	for _, e := range errs {
		prefix := ErrorSymbol
		style := ErrorStyle

		if e.Field == "content" {
			prefix = WarningSymbol
			style = WarningStyle
		}

		fmt.Printf("  %s %s: %s\n", prefix, FilePathStyle.Render(e.File), style.Render(e.Message))
	}

	fmt.Println()

	summary := KeyValue("Errors", NumberStyle.Render(fmt.Sprintf("%d", errors))) + "  " +
		KeyValue("Warnings", NumberStyle.Render(fmt.Sprintf("%d", warnings)))

	if errors > 0 {
		fmt.Println(Panel("Validation Failed", summary))
	} else {
		fmt.Println(Panel("Check Passed with Warnings", summary))
	}
}

type PreFlightResult struct {
	Name    string
	Passed  bool
	Message string
	Warning bool
}

func PrintPreFlight(results []PreFlightResult) {
	fmt.Println()
	fmt.Println("  " + HeadingStyle.Render("Pre-flight checks"))
	fmt.Println()

	for _, r := range results {
		if r.Passed {
			fmt.Printf("  %s %s\n", SuccessSymbol, r.Name)
		} else if r.Warning {
			fmt.Printf("  %s %s — %s\n", WarningSymbol, r.Name, WarningStyle.Render(r.Message))
		} else {
			fmt.Printf("  %s %s — %s\n", ErrorSymbol, r.Name, ErrorStyle.Render(r.Message))
		}
	}

	failed := 0
	for _, r := range results {
		if !r.Passed && !r.Warning {
			failed++
		}
	}

	if failed > 0 {
		fmt.Println()
		fmt.Printf("  %s %s\n", ErrorSymbol, ErrorStyle.Render(fmt.Sprintf("%d pre-flight check(s) failed", failed)))
	}

	fmt.Println()
}
