# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **`build` silently failed on large books with `chromedp render: page load error net::ERR_ABORTED`**: rendering navigated Chrome to a `data:text/html;charset=utf-8;base64,...` URI holding the entire composed document, which Chrome (via chromedp/CDP's `Page.navigate`) aborts once the encoded payload crosses roughly 2MB — with no message indicating size was the cause. A book with a few hundred thousand words of prose and code crosses that threshold easily. Rendering now writes the composed HTML to a temporary file and navigates to it via a `file://` URL instead (removed once the page is captured), which has no such ceiling; the existing default-network-blocking behavior (`NetworkAccess: false` still blocks external `http(s)://` resources) is unaffected. No public API changed. New regression test `TestRenderToPDFLargeDocumentPastOldDataURILimit` generates a document past the old limit and confirms it renders.

## [0.7.0] - 2026-07-07

### Added

- **`gruvbox` builtin theme** (17 total): retro warm dark palette inspired by the popular Gruvbox editor theme (`#282828` background, orange `#fe8019` accent, monospace throughout), requested by name.
- `theme.ExtractCSSVars(css)`: shared parser for `--pdf-*` custom property declarations, extracted out of `scripts/docsgen` so `render` can reuse it too.
- **Automatic PDF quality audit**: `build` now runs a best-effort visual/structural audit right after rendering and reports what it finds — advisory only, it never fails the build. Checks: `overflow-x` (content wider than its box that print will clip instead of wrap), `broken-image`, `empty-content` (near-zero visible text, usually a sign composition silently produced nothing), `low-contrast` (visible text too close in color to its effective background), `heading-clip-risk` (a heading that forces a page break without enough top margin to clear the print engine's header strip — the general form of the bug fixed below, now caught automatically for custom themes/CSS too), and `page-count` (the generated PDF has no detectable pages). Surfaced as a real `Warnings` count and itemized list in `build`'s terminal output and `--json`'s `warnings` array. New library API: `render.RenderToPDFWithAudit` (returns an `*render.AuditReport` alongside the existing error; `render.RenderToPDF`'s signature is unchanged) and `PDF.LastAudit()`.

### Fixed

- **Dark themes no longer print with a white border**: `page.PrintToPDF`'s margin area sits outside the page's paintable content box, so Chrome never fills it with the document's own background — every dark theme (`dark`, `blueprint`, `gruvbox`, ...) rendered as a dark rectangle floating inside a plain white page. Left/right margins are now 0 (the reading margin moved to CSS `padding` on `<body>` instead, which the theme's background paints straight through for true edge-to-edge bleed), and the native top/bottom margin — kept only because that's the one place the running header/page-number footer can render — is now painted with the theme's own `--pdf-bg`/`--pdf-muted` instead of hardcoded white/gray, so it blends into the page instead of showing up as a separate band. `displayHeaderFooter` is now only enabled when a header or page numbers are actually wanted, so documents with both disabled get a fully clean, gap-free page.
- **`@page { margin: ... }` in `base.css` was silently overriding every configured margin**: this Chromium version honors an `@page` margin (even `margin: 0`) over `Page.printToPDF`'s imperative `marginTop`/`marginBottom`/`marginLeft`/`marginRight`, so `render.Options` and `go-pretty-pdf.yml`'s `render.margin_*` had no visible effect whensoever they disagreed with `@page` — silently, with no error. `@page` no longer declares a margin at all; the imperative API parameters are now the single real source of truth.
- **`margin_top`/`margin_bottom`/`margin_left`/`margin_right: "0mm"` in `go-pretty-pdf.yml` was silently ignored**: `WithFullConfig` detected "was a margin configured?" from the *parsed* value being non-zero, making an explicit `0mm` indistinguishable from the field being absent — a config asking for a true full-bleed page got the default margins instead. Now gated on the config *string* being non-empty.
- **Chapter titles and the "Table of Contents" heading rendered clipped, overlapping the running header**: every `h1` that starts a fresh page (`page-break-before: always`, or the TOC's own heading right after the forced break following the cover) had `margin-top: 0`, putting its text flush against the top of the page's content box. With a header/page-number footer enabled, chrome-headless-shell clips roughly the first 0.3in of whatever sits there — confirmed by disabling the header (clean render) and by testing plain body text landing on a fresh page through natural pagination instead of a forced break (also clean), so the defect is specific to content flush against a *forced* page break while a header/footer is displayed. `h1` now keeps a `0.35in` top margin, comfortably clearing the dead zone on every theme (none of which override it), guarded by `TestBaseCSSH1HasTopMarginBuffer`.

Known limitation: Chrome reserves a small, fixed ~0.2in strip at the very top/bottom of the page whenever a header/footer is displayed at all, regardless of the configured margin or the header/footer template's own CSS — confirmed by direct testing, not something this project's CSS can override. It's colored to match the theme so it no longer looks like a stray white band, but on a dark theme with a header or page numbers enabled you may still see a hairline. Disable both (`--no-header --no-page-numbers`, or `theme_options.sections.header`/`page_numbers: false`) and set all four margins to `0mm`/`0in` for a page with zero gap on every side.

## [0.6.0] - 2026-07-07

### Added

- **8 new builtin themes**, bringing the total to 16 — each required only a CSS file and a registry entry in `theme/builtin.go` to appear correctly in the CLI (`pretty-pdf theme list`) and the docs website (switcher + its own dogfooded PDF), with zero other files touched:
  - `sepia` — warm, soft palette for long reading sessions
  - `terminal` — all-monospace, terminal-inspired, with a `$ ` prompt-style cover
  - `blueprint` — dark technical palette with cyan highlights and a dashed cover border
  - `ivy` — classic Ivy League university letterhead (forest green and gold)
  - `government` — formal official-document palette (navy and bronze, centered headings, double rules)
  - `resume` — clean ATS-friendly sans-serif for CVs; disables cover/TOC/page numbers/header by default
  - `legal` — stark, formal brief style with no color used as decoration
  - `latex` — mathematical/scientific paper look with automatic, chapter-scoped section numbering (1., 1.1, 1.2, 2., ...) via CSS counters
- Five new theme categories: `warm`, `technical`, `institutional`, `resume`, `formal`.
- `theme.Theme.Accented`: marks builtin themes (classic, modern, corporate, editorial, terminal, blueprint, ivy, government) that use their accent color as a bold structural element (cover border, accent blockquote) rather than just for links.
- `resumeSections`: a `ResolvedSections` preset (all sections off) for themes meant for short, single-flow documents.

### Changed

- **docs site theme automation**: `scripts/docsgen` now derives every theme's colors, fonts, and accent treatment straight from `theme.List()` and each theme's own CSS at build time (see `themevars.go`), instead of a hand-maintained, hardcoded copy of the palette in `site.css`. Adding a builtin theme to `theme/builtin.go` is now the only step needed for it to appear in the docs site's theme switcher, get correct swatch colors, and get its own dogfooded "docs as a PDF" download — nothing to keep in sync by hand. This is also what caused the "incoherent blue" bug fixed earlier: the site's copy had drifted from the real theme CSS.

## [0.5.0] - 2026-07-07

### Added

- **Automatic Chrome management** (`chromemgr` package): `pretty-pdf build`/`watch` no longer require Chrome/Chromium to be installed manually. If none is found, a small official "chrome-headless-shell" build (Google's Chrome for Testing distribution) is downloaded and cached under the OS user cache dir on first use, then reused on every later run — no different from what Playwright/Puppeteer do. A system-installed Chrome/Chromium is always preferred and used as-is when present. New `--chrome-path` flag / `PRETTY_PDF_CHROME_PATH` env var let users pin a specific binary and skip detection entirely. Covers linux/amd64, darwin/amd64, darwin/arm64, windows/amd64 (linux/arm64 has no upstream prebuilt binary yet and falls back to a clear error asking for a manual install).
- **`render.Options.ChromeExecPath`** and **`prettypdf.WithChromeExecPath`**: point rendering at a specific Chrome/Chromium binary instead of chromedp's default discovery.
- **Named theme constants**: `NameDefault`, `NameMinimal`, `NameModern`, `NameClassic`, `NameCorporate`, `NameDark`, `NameAcademic`, `NameEditorial` exported from `theme` package — custom theme code and tests can now reference themes by constant instead of raw strings, eliminating goconst lint warnings across the codebase

### Changed

- **Chrome startup timeout**: raised `chromedp.WSURLReadTimeout` to 45s in `RenderToPDF` and 15s in `CheckChromeAvailable`, plus boosted `CheckChromeAvailable` context timeout from 10s to 20s — prevents spurious "websocket url timeout reached" failures on cold/loaded CI runners

### Fixed

- **goconst lint**: all magic string literals for builtin theme names and categories replaced with named constants in `theme/builtin.go`; tests updated to use constants and share a `testCustomThemeName` const

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

- `build` and `theme` CLI commands now carry expanded help text with full `Long` descriptions, `Example` blocks, and dynamic theme name listing
- `README.md` theme section updated: lists all 8 builtin themes, documents `WithThemeName(name, opts)`, adds CLI usage examples for theme customization
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

## [0.2.1] - 2026-07-03

### Fixed

- **Data URI corruption**: switched from `url.QueryEscape` to base64 encoding for the HTML data URI passed to Chrome — `QueryEscape` converts spaces to `+`, which Chrome does not decode in data URIs, so every space in rendered text showed up as a literal `+`

## [0.2.0] - 2026-07-03

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

[0.6.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.6.0
[0.5.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.5.0
[0.4.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.4.0
[0.3.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.3.0
[0.2.1]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.2.1
[0.2.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.2.0
[0.1.0]: https://github.com/sazardev/go-pretty-pdf/releases/tag/v0.1.0
