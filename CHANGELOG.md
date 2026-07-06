# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-07-06

### Added

- `WithFullConfig(cfg)` option: applies the entire `config.Config` struct (source, output, title, subtitle, author, CSS/template, theme, vars, render settings) in a single call
- `WithNetworkAccess(bool)` option: control whether headless Chrome can make outbound network requests during rendering (default: `false`)
- `config.ParsePaperSize(name)` and `config.ParseCSSUnit(s)` exported functions (moved from `cmd/pretty-pdf/config.go`)
- `config.PaperLetter` constant for YAML config comparisons
- `render.Options.NetworkAccess` field, with default network blocking via `chromedp/cdproto/network`
- Concurrent-safe `ComponentRegistry` (`sync.RWMutex` protecting the handler map)
- `headerTitleSet` tracking in `PDF`: prevents `New()` from overwriting an explicit header title with the document title
- Deferred warning buffer in `PDF`: `WithConfigCSSAndTemplate` file-read failures are collected and flushed by `New()` after all options run, making warning output order-independent from `WithVerbose`
- Comprehensive test suite: `pdf_test.go` (16 tests), `compose/compose_test.go` (6 tests), `compose/toc_test.go` (3 tests), `config/units_test.go` (2 tests), `config/units_test.go` (2 tests), `mdx/component_test.go` (1 race test), `mdx/parser_test.go` (7 tests), `render/render_test.go` (4 tests)
- Showcase book: 8-document MDX example under `examples/showcase/` with 8 custom components (`Callout`, `Badge`, `Steps`, `Card`, `Stat`, `Timeline`, `Quote`, `Progress`)
- Showcase integration test: `examples/showcase_test.go` verifies compose output and full PDF rendering
- Trust model documentation in `README.md`, `SECURITY.md`, and package-level `doc.go`

### Changed

- `cmd/pretty-pdf/buildOpts` simplified to `WithFullConfig` + `WithValidator` (removed duplicated config-to-option mapping)
- `WithConfigCSSAndTemplate` file-read warnings now deferred to `New()` instead of printed inline (order-independent from `WithVerbose`)

### Security

- **Network access blocked by default** during headless Chrome rendering: prevents SSRF/exfiltration from untrusted MDX content via `<script>`, `<img>`, `<link>`, etc.
- Detailed trust model documented across `README.md`, `SECURITY.md`, and `doc.go`

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
- Support for h4 and h5 heading levels in MDX documents, with matching TOC styling and CSS

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
