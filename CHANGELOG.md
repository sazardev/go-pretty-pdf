# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-07-06

### Added

- **Theme engine overhaul**: `theme` package now has a proper engine with `Resolve()`, `ResolveByName()` — merges base CSS + theme CSS + `:root` custom property overrides (colors, fonts, density) + section toggles
- **8 builtin themes**: default, minimal, modern, classic, corporate, dark, academic, editorial — each with dedicated CSS files embedded via `//go:embed`, with categories (professional, editorial, dark, academic, minimal)
- **Custom theme system**: `<name>.theme.yml` files extending a builtin theme, discovered in `./themes/` (project-local) and `~/.config/pretty-pdf/themes` (global) — full YAML schema with `extends`, `colors`, `fonts`, `sections`, `density`, and raw `css` escape hatch
- **`theme` CLI subcommand** with 4 subcommands:
  - `theme list` — shows builtin + custom themes with descriptions
  - `theme show <name>` — prints fully-resolved CSS to stdout
  - `theme new <name>` — scaffolds a starter `.theme.yml` (with `--from` and `--global` flags)
  - `theme add <path>` — imports existing `.theme.yml` or `.css` files as managed custom themes (with `--as` and `--global` flags)
- **Section toggles** (cover, TOC, page numbers, header) controlled via:
  - CLI flags: `--no-cover`, `--no-toc`, `--no-page-numbers`, `--no-header`
  - Config: `theme_options.sections.cover`, `.toc`, `.page_numbers`, `.header` (nullable booleans — unset = theme default)
  - Template gating: `{{if .ShowCover}}` / `{{if .ShowTOC}}` wrapping the cover block and TOC in `template.html`
  - CSS gating: `.cover{display:none !important;}` / `.toc{display:none !important;}` appended by `Resolve()` for disabled sections
  - `render.Options.PageNumbers` and `render.Options.ShowHeader` — when disabled, Chrome header/footer templates render `<div></div>` (empty)
- **Color/font customization**: `--color-*` and `--font-*` CLI flags + `theme_options.colors`/`fonts` in config — drives `--pdf-*` CSS custom properties in a `:root` block
- **Density control**: `--density compact|normal|relaxed` CLI flag + `theme_options.density` — adjusts `--pdf-line-height` and `--pdf-space-scale`
- **Google Fonts support**: `fonts.google_fonts` in theme YAML/config, fetched only when `allow_network_fonts: true` (network disabled by default for security)
- **`WithThemeName(name, opts)` option**: resolves a theme by name (builtin, custom, or file path) with full opts customization, wiring section toggles into `composeOpts`/`renderOpts`
- **`WithNetworkAccess(bool)` wired into CLI**: `--allow-network-fonts` flag enables outbound Chrome requests
- **Config struct**: `ThemeOptionsConfig`, `ColorsConfig`, `FontsConfig`, `SectionsConfig` with full YAML serialization
- **Test suite**: 20 new tests across `theme/`, `pdf_test.go`, `compose/compose_test.go`, `render/render_test.go`, `config/config_test.go`

### Changed

- `theme.Theme` struct: now includes `Description`, `Category`, `Sections` (resolved defaults), and `CSS` comes from dedicated asset files instead of raw Go strings
- `WithTheme(t)` — now applies CSS only (no template); section toggles must be set separately
- `WithConfigCSSAndTemplate` — resolves `cfg.Theme` via `ResolveByName` with full `ThemeOptionsConfig` customization before applying explicit CSS/template file overrides
- Old hardcoded `theme.Minimal.CSS` inline string replaced by `//go:embed assets/minimal.css`
- `cmd/pretty-pdf/main.go` — `--theme` flag usage now lists all builtin theme names dynamically from `theme.List()`
- `cmd/pretty-pdf/config.go` — maps CLI flags to `cfg.ThemeOptions` (colors, fonts, sections, density, network)
- `docs/cli.md` — comprehensive docs for all new flags, config fields, theme subcommands, custom theme workflow

### Removed

- Inline `minimalCSS` string in `theme/theme.go` — CSS now lives in `theme/assets/*.css`

### Security

- Google Fonts (`fonts.google_fonts`) require explicit `allow_network_fonts: true` — network access remains blocked by default during headless Chrome rendering

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

[0.4.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.4.0
[0.3.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.3.0
[0.2.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.2.0
[0.1.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.1.0
