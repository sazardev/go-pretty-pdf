package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	prettypdf "github.com/sazardev/go-pretty-pdf"
)

var (
	sourceDir string
	outPath   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pretty-pdf",
		Short: "go-pretty-pdf — Beautiful PDF generation from MDX",
		Long: `Transforms MDX source files into high-fidelity, print-ready PDF
documents using headless Chrome. Supports custom MDX components,
themes, templates, and validation plugins.`,
	}

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build PDF from MDX source directory",
		RunE:  runBuild,
	}
	buildCmd.Flags().StringVar(&sourceDir, "source", "./docs", "Path to the source MDX directory")
	buildCmd.Flags().StringVar(&outPath, "out", "out.pdf", "Output PDF path")

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate MDX files without rendering",
		RunE:  runValidate,
	}
	validateCmd.Flags().StringVar(&sourceDir, "source", "./docs", "Path to the source MDX directory")

	rootCmd.AddCommand(buildCmd, validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runBuild(cmd *cobra.Command, args []string) error {
	fmt.Printf("Building PDF from %s -> %s\n", sourceDir, outPath)

	pdf, err := prettypdf.New(
		prettypdf.WithSourceDir(sourceDir),
		prettypdf.WithOutputFile(outPath),
		prettypdf.WithTitle("Document"),
	)
	if err != nil {
		return err
	}

	fmt.Println("Rendering PDF...")
	if err := pdf.Build(cmd.Context()); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("Done: %s\n", outPath)
	return nil
}

func runValidate(cmd *cobra.Command, args []string) error {
	fmt.Printf("Validating MDX files in %s\n", sourceDir)

	pdf, err := prettypdf.New(
		prettypdf.WithSourceDir(sourceDir),
	)
	if err != nil {
		return err
	}

	docs, err := pdf.Validate(cmd.Context())
	if err != nil {
		if docs != nil {
			for _, e := range docs {
				fmt.Printf("  - %v\n", e)
			}
		}
		return err
	}

	fmt.Println("All files valid.")
	return nil
}
