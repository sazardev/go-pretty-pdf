package main

import (
	"fmt"
	"os"
	"path/filepath"
	"os/signal"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	prettypdf "github.com/sazardev/go-pretty-pdf"
	"github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf/output"
)

func runWatch(cmd *cobra.Command, args []string) error {
	if noColor {
		output.NoColor()
	}

	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	if !quiet {
		output.PrintBanner(version)
		fmt.Println("  " + output.KeyValue("Watching", cfg.Source))
		fmt.Println("  " + output.KeyValue("Output", cfg.Output))
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	if err := filepath.WalkDir(cfg.Source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walking source: %w", err)
	}

	if !quiet {
		output.PrintWatchBanner()
	}

	stats := output.WatchStats{Running: true}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	rebuildCh := make(chan string, 1)

	debounceTimer := time.NewTimer(0)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Ext(event.Name) != ".mdx" && filepath.Ext(event.Name) != ".yaml" && filepath.Ext(event.Name) != ".yml" {
					continue
				}
				if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
					debounceTimer.Reset(300 * time.Millisecond)
				}
			case <-debounceTimer.C:
				rebuildCh <- "rebuild"
			}
		}
	}()

	go func() {
		for range sigCh {
			fmt.Println()
			output.PrintWatchSummary(stats)
			os.Exit(0)
		}
	}()

	for range rebuildCh {
		output.PrintWatchRebuild()

		startTime := time.Now()

		opts := buildOpts(cfg)
		pdf, err := prettypdf.New(opts...)

		if err != nil {
			stats.RecordError()
			fmt.Println(output.Error(fmt.Sprintf("Initialization failed: %v", err)))
			continue
		}

		spinner := output.StartSpinner("Rebuilding...")
		err = pdf.Build(cmd.Context())
		if err != nil {
			spinner.Fail(err.Error())
			stats.RecordError()
		} else {
			spinner.Done(fmt.Sprintf("Done in %s", time.Since(startTime).Round(time.Millisecond)))
			stats.RecordBuild()
		}
	}

	return nil
}
