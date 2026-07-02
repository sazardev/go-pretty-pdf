# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-07-02

### Added

- `serve` subcommand: preview MDX as HTML in the browser with live reload
- `completion` subcommand: generate shell completion scripts (bash, zsh, fish, powershell)
- `--bare` flag on `init`: minimal non-interactive project scaffolding
- `--port` flag on `serve`: configure HTTP server port
- `ValidateAll(docs)` method on `PDF` and `Validator` interface for batch validation
- `TestAnchorID` test covering bracket-stripping behavior
- `.component-warning-title` CSS class for `<Warning>` title styling
- Template placeholders (`{{BOOK_TITLE}}`, `{{AUTHOR_NAME}}`, `{{SOURCE_DIR}}`, `{{THEME}}`) in init scaffold files

### Changed

- `AnchorID` uses `strings.Trim` instead of `strings.ReplaceAll` for bracket removal
- `PrintValidationSummary` now shows per-file breakdown (passed/errored/warned)
- `render.DefaultOptions()` drives margin defaults in config instead of hardcoded 0.8
- `FindConfig` resolves from `os.Getwd()` instead of relative `.`
- Warning component uses `html.EscapeString` from stdlib instead of custom `escapeHTML`

### Removed

- `AnchorIDRaw` function (unused)
- `PrintSummary` method from `PipelineProgress` (unused)
- Deprecated `LintConfig` fields: `RequireIDFormat`, `RequireLowercaseFilenames`, `CheckBrokenLinks`
- Version banner from `init`, `check`, and `watch` commands (noise reduction)
- Custom `dirname` helper in render — replaced by `filepath.Dir`
- Per-doc validation loop in `build.go` — replaced by single `ValidateAll` call

### Fixed

- `url.PathEscape` → `url.QueryEscape` for correct data URI encoding in Chrome rendering

## [0.1.0] - 2026-07-02

### Added

- MDX parser based on goldmark with YAML frontmatter support
- Custom component transpiler: `<DeepDive>`, `<Warning>`, `<Axiom>` built-in
- Variable substitution (`{{key}}`) before parsing
- HTML composition with embedded template and print CSS
- Table of Contents auto-generated from `[X.Y.Z]` IDs
- Headless Chrome PDF rendering via chromedp
- Theme system with Default and Minimal built-in themes
- CLI with cobra: `build`, `check`, `init`, `watch`, `version` commands
- Animated pipeline UI with lipgloss styles, spinner, and progress panels
- Interactive init wizard using huh form library
- File watcher with 300ms debounce for live rebuilds
- YAML config file (`go-pretty-pdf.yml`) with auto-discovery
- Configurable lint validator: required frontmatter, ID format, duplicates, heading depth
- Paper size presets: A4, Letter, Legal
- CSS unit parser: mm, cm, in, pt, px
- Comprehensive example suite with custom components, themes, and CSS
- 18 functional options on the root PDF type
- Step-by-step API: ParseDir, ValidateDoc, ComposeHTML, Render
- Partial parsing: per-file errors collected, valid docs proceed
- GitHub Actions CI (lint, test, vet, build on 3 OS) and release pipeline (goreleaser)
- Local Makefile with lint, test, build, and release-dry-run targets

[0.1.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.1.0
