# go-pretty-pdf — System Memory

## Purpose
Transform MDX source files into print-ready PDF via headless Chrome.
Both a Go library (`prettypdf`) and CLI tool (`pretty-pdf`).

---

## Pipeline

```
Raw MDX → substituteVars() ({{var}}) → goldmark parse → Transpile custom components → Compose HTML (TOC + template + CSS) → Render PDF (headless Chrome)
```

---

## Package Map

```
cmd/pretty-pdf/          CLI entrypoint (cobra) — build, check, init, version, watch
cmd/pretty-pdf/output/   DX output package: lipgloss styles, spinner, banner, panels, pipeline progress
config/                  YAML config loader — Config struct, Load(), FindConfig(), Default()
pdf.go                   Root package prettypdf — New(), Build(), Validate(), 18 functional options
                         ↑ step-by-step: ParseDir(), ValidateDoc(), ComposeHTML(), Render()
mdx/                     Parser (goldmark), component transpiler, DefaultValidator, Document type
                         ↑ partial parsing: ParseFileError, ParseErrors (collect per-file, continue)
compose/                 HTML composition — template.html + print.css (go:embed), TOC builder
render/                  Chrome headless PDF rendering via chromedp
theme/                   Theme struct with Default and Minimal built-in themes
```

---

## Config File (`go-pretty-pdf.yml`)

Auto-discovered in CWD (or via `--config` flag). Precedence: **CLI flags > config values > code defaults**.

### All Fields

```yaml
title: "My Book"
subtitle: "A Guide"
author: "Jane Doe"
source: book
output: out.pdf
theme: default

css: custom.css
template: custom-template.html

vars:
  api_version: "v2.1"
  company: "Acme Corp"

lint:
  require_frontmatter: [id, title]
  require_id_format: "[X.Y.Z]"
  no_duplicate_ids: true
  max_heading_depth: 3
  require_lowercase_filenames: false
  check_broken_links: false

render:
  timeout: 60s
  paper: a4
  margin_top: 20mm
  margin_bottom: 20mm
  margin_left: 15mm
  margin_right: 15mm
  header_title: "{{title}}"
```

### Defaults (when field is absent)

| Field | Default |
|-------|---------|
| `source` | `"book"` |
| `output` | `"out.pdf"` |
| `title` | `"Document"` |
| `author` | `"go-pretty-pdf"` |
| `theme` | `""` (→ default theme) |
| Lint: require_frontmatter | `[id, title]` |
| Lint: no_duplicate_ids | `true` |
| Lint: max_heading_depth | `3` |
| Render: timeout | `60s` |
| Render: paper | `A4` (8.27×11.69 in) |
| Render: margins | top/bottom=0.8in, left/right=0.6in |

### CSS/Template Path Resolution

**Critical**: CSS and template file paths in the YAML config are resolved **relative to the config file's directory**, not CWD. This prevents breakage when running from a different working directory.

- `loadConfig()` at `main.go:111` resolves paths:
  - Config file paths (`cfg.CSS`, `cfg.Template`): joined with `filepath.Dir(configPath)`
  - CLI flag paths (`--css`, `--template`): resolved via `filepath.Abs()`
  - `FindConfig()` at `config/config.go:72` returns **absolute path** via `filepath.Abs()`
- **Gotcha**: CSS/template files must exist in config or be explicitly set via CLI; there's NO auto-detection of `custom.css` in the source dir

---

## CLI Commands

### `pretty-pdf build`

Animated pipeline: banner → pre-flight → parse (spinner) → validate (spinner) → compose (spinner) → render (spinner) → build summary panel.

Flags:
| Flag | Type | Overrides config | Extras |
|------|------|-----------------|--------|
| `--config` | string | — | |
| `--source`, `-s` | string | `cfg.Source` | |
| `--out`, `-o` | string | `cfg.Output` | |
| `--title` | string | `cfg.Title` | |
| `--subtitle` | string | `cfg.Subtitle` | |
| `--author` | string | `cfg.Author` | |
| `--theme` | string | `cfg.Theme` | |
| `--css` | string | `cfg.CSS` | |
| `--template` | string | `cfg.Template` | |
| `--timeout` | string | `cfg.Render.Timeout` | |
| `--verbose`, `-v` | bool | — | |
| `--json` | bool | — | silent JSON output |
| `--no-color` | bool | — | ASCII-only output |
| `--quiet` | bool | — | suppress config display |

**Flow** (`build.go:19`):
```
runBuild()
  → output.NoColor() if --no-color
  → output.PrintBanner(version)          # ASCII art "GO → PDF"
  → runPreFlight(cfg)                    # check Chrome, source, output, CSS, template
  → output.PrintPreFlight(results)       # ✓/✗ per check
  → abort if any non-warning failure
  → NewPipelineProgress(4 steps)
  → for each step: Start (spinner) → work → Done/Fail
  → output.PrintBuildSummary(BuildStats)  # styled panel
```

Pre-flight checks (runPreFlight, `build.go:198`):
- Chrome/Chromium available (calls `render.CheckChromeAvailable()`)
- Source directory exists + has .mdx files
- Output directory writable (auto-creates if needed)
- CSS file exists (if configured, warning on fail)
- Template file exists (if configured, warning on fail)

**Step-by-step API** (`pdf.go`):
```
pdf.ParseDir()       → []*mdx.Document, error (partial: docs + err list)
pdf.ValidateDoc(doc) → []mdx.ValidationError (runs DefaultValidator per doc)
pdf.ComposeHTML(docs)→ string (HTML), error
pdf.Render(html)     → error (headless Chrome)
```

### `pretty-pdf check`

Styled validation: spinner for parsing → results panel.

Flags: `--config`, `--source`, `--strict`, `--verbose`, `--json`, `--no-color`

- `--strict`: promotes heading depth warnings to errors
- `--verbose`: prints warnings during config loading

Output: colored summary `N error(s), M warning(s)` via `PrintValidationSummary`.
Exit code 1 if errors > 0.

### `pretty-pdf init [dir]`

**Interactive mode** (default): huh form wizard asking for title, author, theme (default/minimal), source dir, confirms before scaffolding.

**JSON/bare mode** (`--json`): scaffold with all defaults, no prompts.

Scaffolds:
- `go-pretty-pdf.yml` (default config)
- `[1.0.0]-introduction.mdx`
- `[1.1.0]-getting-started.mdx`
- `[1.1.1]-installation.mdx` (has `{{product}}` / `{{company}}` var example)

Fails if target directory already exists.

### `pretty-pdf watch`

fsnotify recursive watcher. 300ms debounce. Animated banners per rebuild.

Ctrl+C handler prints summary panel (builds, errors, last build time).

Flags: same as `build` + `--config`.

### `pretty-pdf version`

Prints `pretty-pdf <version>`. Version defaults to `"dev"`, override with `-ldflags "-X main.version=X.Y.Z"`.

### Global Flags

| Flag | Affects |
|------|---------|
| `--json` | build, check, init — silent structured output |
| `--no-color` | all — ASCII-only (tearmenv.Ascii) |
| `--quiet` | build — suppress config info display |
| `--verbose`, `-v` | all |

---

## Core API (`pdf.go`)

### Options Table

| Option | Signature | Sets | Default |
|--------|-----------|------|---------|
| `WithSourceDir` | `(dir string)` | source | `"book"` |
| `WithOutputFile` | `(path string)` | output | `"out.pdf"` |
| `WithTitle` | `(t string)` | compose title | `"Document"` |
| `WithSubtitle` | `(s string)` | compose subtitle | `""` |
| `WithAuthor` | `(a string)` | compose author | `"go-pretty-pdf"` |
| `WithCSS` | `(css string)` | compose CSS content | embedded `print.css` |
| `WithTemplate` | `(html string)` | compose template | embedded `template.html` |
| `WithTheme` | `(t theme.Theme)` | CSS + Template | default |
| `WithComponent` | `(name string, handler mdx.ComponentHandler)` | registers **(appends)** | DeepDive, Warning, Axiom |
| `WithValidator` | `(v mdx.Validator)` | validator | nil (none) |
| `WithTimeout` | `(d time.Duration)` | render timeout | 60s |
| `WithHeaderTitle` | `(t string)` | PDF header title | compose Title |
| `WithVerbose` | `(v bool)` | verbose logging | false |
| `WithVars` | `(vars map[string]string)` | parser var substitution | nil |
| `WithRenderMargins` | `(t,b,l,r float64)` | PDF margins (inches) | 0.8/0.8/0.6/0.6 |
| `WithPaperSize` | `(w,h float64)` | PDF paper (inches) | A4 (8.27×11.69) |
| `WithConfig` | `(cfg *config.Config)` | source, output, title, subtitle, author | — |
| `WithConfigCSSAndTemplate` | `(cfg *config.Config)` | CSS file, template file, theme | — |

### `WithConfigCSSAndTemplate` Details

Reads `cfg.CSS` and `cfg.Template` from disk using `os.ReadFile`. Errors are silently swallowed unless verbose is true (prints to stderr). Maps `cfg.Theme` to theme constants ("default" → `theme.Default`, "minimal" → `theme.Minimal`).

### `WithComponent` Fix

Originally replaced the entire parser (`p.parser = mdx.NewParser(...)`). Now calls `p.parser.RegisterComponent(name, handler)` which appends without losing previously registered components or built-in defaults.

### Options Order Matters

Options execute in the order they're passed to `New()`. `WithVerbose` should come **before** `WithConfigCSSAndTemplate` so verbose logging of file read errors works.

### `New()` Constructor

```
New(opts...)
  → Defaults: source="book", output="out.pdf", parser=NewParser(), composeOpts=DefaultOptions(), renderOpts=DefaultOptions()
  → Auto-sets: renderOpts.HeaderTitle = composeOpts.Title
  → Applies all opts in order
  → Returns *PDF, nil
```

### `Build(ctx)` Pipeline

```
parser.ParseDir(sourceDir)                  # walk .mdx files, parse, transpile, sort by [X.Y.Z]
  → if errors: log partial failures, continue with valid docs
  → if validator: validator.ValidateAll(docs)
  → compose.ComposeHTML(docs, composeOpts)   # TOC + template + CSS
  → render.RenderToPDF(html, output, renderOpts)  # headless Chrome
```

### `Validate(ctx)` Pipeline

```
parser.ParseDir(sourceDir)
  → if validator: validator.ValidateAll(docs)
  → return errors (no compose/render)
```

### Step-by-step API (added in DX overhaul)

```
pdf.ParseDir()       →  ([]*mdx.Document, error)
  Delegates to parser.ParseDir, returns partial docs + ParseErrors.
  Suppresses frontmatter-not-found per-file (not an error).

pdf.ValidateDoc(doc) →  []mdx.ValidationError
  Runs pdf.validator.Validate(doc) if set.

pdf.ComposeHTML(docs)→  (string, error)
  Delegates to compose.ComposeHTML(docs, pdf.composeOpts).

pdf.Render(html)     →  error
  Delegates to render.RenderToPDF with configured opts.
```

---

## MDX Parser (`mdx/parser.go`)

### Parser Struct

```go
type Parser struct {
    md         goldmark.Markdown
    components *ComponentRegistry
    vars       map[string]string
}
```

### ParseFile Flow
1. `os.ReadFile(path)`
2. `substituteVars(raw)` — replaces `{{key}}` with `vars[key]` before goldmark
3. Goldmark conversion (GFM, meta, unsafe HTML, auto heading IDs)
4. Extract frontmatter from goldmark context via `meta.Get(ctx)`
5. `components.Transpile(html)` — regex replacement of custom component tags

### ParseDir Flow
1. Walk directory for `.mdx` (case-insensitive suffix)
2. `parseFile` each
3. Sort by `SortKey()` = `[X.Y.Z]` → `[3]int` (major.minor.patch)

### Variable Substitution

`substituteVars(raw string) string` runs **before** goldmark parsing, so vars work in:
- Frontmatter YAML values (e.g. `header_title: "{{product}} v{{version}}"`)
- Body text
- Code blocks
- Component attributes

Syntax: `{{key}}` — simple string replacement using `strings.ReplaceAll(raw, "{{"+k+"}}", v)`.

Set via `WithVars(map)` parser option or `prettypdf.WithVars(map)` root option.

### Components

Registered in `ComponentRegistry`. Handlers transpiled via regex after goldmark HTML.

Built-in:
- `<DeepDive title="...">` → `<aside class="component-deep-dive">` (blue)
- `<Warning title="...">` → `<div class="component-warning">` (orange)
- `<Axiom>` → `<blockquote class="component-axiom">` (green italic)

`RegisterComponent(name, handler)` on Parser allows non-destructive addition.

---

## Validator (`mdx/validator.go`)

### DefaultValidator

```go
type DefaultValidator struct {
    RequireFrontmatter []string
    RequireIDFormat    string
    NoDuplicateIDs     bool
    MaxHeadingDepth    int
}

func NewDefaultValidator() *DefaultValidator
func (v *DefaultValidator) Validate(doc *Document) []ValidationError
func (v *DefaultValidator) ValidateAll(docs []*Document) []ValidationError
```

### Rules

| Rule | Severity | Condition |
|------|----------|-----------|
| Required frontmatter field | Error | Field missing or empty string |
| ID format | Error | Doesn't match `^\[\d+\.\d+\.\d+\]$` |
| Duplicate IDs | Error | Same `[X.Y.Z]` in >1 file |
| Heading depth | Warning (Error with `--strict`) | HTML has `<hN>` with N > `MaxHeadingDepth` |

`ValidateAll()` runs `Validate()` per doc, then checks cross-doc rules (duplicates).

### ValidationError

```go
type ValidationError struct {
    File    string
    Field   string
    Message string
}
```

---

## Render (`render/render.go`)

Headless Chrome via `chromedp`. HTML encoded as `data:text/html;charset=utf-8,...` URI.

PDF features enabled: `PrintBackground`, `DisplayHeaderFooter`, `GenerateDocumentOutline` (bookmarks), `GenerateTaggedPDF` (accessibility).

### Paper Size Mapping (CLI)

| Name | Width×Height (inches) |
|------|----------------------|
| A4 | 8.27 × 11.69 |
| Letter | 8.5 × 11 |
| Legal | 8.5 × 14 |

### CSS Unit Parser (`main.go:405`)

Converts CSS unit strings to inches:
- `mm` → / 25.4
- `cm` → / 2.54
- `in` → value (identity)
- `pt` → / 72.0
- `px` → / 96.0

---

## Theme System (`theme/theme.go`)

```go
type Theme struct {
    Name     string
    CSS      string   // Override CSS (empty = use embedded)
    Template string   // Override template (empty = use embedded)
}
```

Built-in: `Default` (empty → embedded assets), `Minimal` (stripped-down CSS, default template).

---

## Testing

```bash
go test ./...                    # 26 tests, all pass
go test ./mdx/... -v             # Component (6), parser/file/dir (3), sort key (1), validator (6), vars (2), custom comp (1)
go test ./config/... -v          # Config defaults (1), Load (1), partial defaults (1), FindConfig (1)
```

### Test Breakdown

| Package | Test Count | What |
|---------|-----------|------|
| `mdx/` | 9 | Component transpile (6), ParseFile (1), ParseDir (1), SortKey (1) |
| `mdx/` (validator) | 6 | DefaultValidator (4), ValidateAll/duplicates (1), heading depth/strict (1) |
| `mdx/` (vars + partial parse) | 3 | Vars substitution (1), vars with components (1), partial parse errors (1) |
| `config/` | 4 | Default(), Load(), LoadDefaultsOnMissingKeys(), FindConfig() |

---

## Key File Reference

| File | Lines | Role |
|------|-------|------|
| `cmd/pretty-pdf/main.go` | ~110 | Thin cobra entry — global vars, root/builtin/version/cmd/watcher commands |
| `cmd/pretty-pdf/config.go` | ~120 | `loadConfig()`, `buildOpts()`, `parseCSSUnit()`, `validatorFromConfig()`, `parserFromConfig()` |
| `cmd/pretty-pdf/build.go` | ~332 | `runBuild()` animated pipeline, `runBuildJSON()`, `runPreFlight()`, `countMDXFiles()`, `formatBytes()` |
| `cmd/pretty-pdf/check.go` | ~100 | `runCheck()` styled validation with spinner + summary panel |
| `cmd/pretty-pdf/init.go` | ~160 | `runInit()` huh interactive wizard, `runInitBare()` JSON mode, `scaffoldWithConfig()` |
| `cmd/pretty-pdf/watch.go` | ~140 | `runWatch()` fsnotify recursive watcher with debounce + stats |
| `cmd/pretty-pdf/output/styles.go` | ~120 | Lipgloss styles: colors, symbols, Panel/Success/Error/Warn/Info helpers, `NoColor()` |
| `cmd/pretty-pdf/output/spinner.go` | ~60 | Animated goroutine spinner with `ack` channel sync |
| `cmd/pretty-pdf/output/banner.go` | ~30 | ASCII art "GO → PDF" banner |
| `cmd/pretty-pdf/output/panels.go` | ~110 | `BuildStats`, `PrintBuildSummary`, `PrintValidationSummary`, `PreFlightResult`, `PrintPreFlight` |
| `cmd/pretty-pdf/output/progress.go` | ~165 | `PipelineProgress` (Start/Done/Fail/Skip/PrintSummary), `WatchStats`, `PrintWatchBanner/Rebuild/Summary` |
| `cmd/pretty-pdf/initassets/*` | 4 files | Scaffold templates for `init` command |
| `pdf.go` | ~220 | Root API, 18 options, Build/Validate pipeline, step-by-step methods, WithComponent fix |
| `config/config.go` | 82 | Config struct, YAML Load(), FindConfig(), Default() |
| `config/config_test.go` | 145 | 4 config tests |
| `mdx/parser.go` | ~210 | Goldmark parser, variable substitution, RegisterComponent, `ParseFileError`, `ParseErrors` |
| `mdx/validator.go` | ~105 | DefaultValidator with 4 lint rules |
| `mdx/validator_test.go` | ~252 | Validator + variable tests |
| `mdx/mdx.go` | 133 | Document type, frontmatter accessors, ID utilities |
| `mdx/component.go` | 104 | Component transpilation (DeepDive, Warning, Axiom) |
| `mdx/mdx_test.go` | 314 | Original parser/component tests |
| `compose/compose.go` | 107 | HTML composition, template execution, keywords |
| `compose/toc.go` | 45 | TOC builder from `[X.Y.Z]` hierarchy |
| `compose/assets/template.html` | 29 | Default HTML template (go:embed) |
| `compose/assets/print.css` | 277 | Default print CSS (go:embed) |
| `render/render.go` | 126 | Chrome headless PDF render |
| `theme/theme.go` | 87 | Theme struct, Default & Minimal |
| `examples/full-demo/go-pretty-pdf.yml` | 35 | Full demo config with all sections including CSS+template refs |
| `examples/full-demo/custom.css` | 208 | Demo custom CSS (dark cover, blue accents, code dark mode) |
| `examples/full-demo/custom-template.html` | 30 | Demo custom template (minimal, dark cover) |
| `examples/full-demo/*.mdx` | 4 files | Demo content exercising vars, components, frontmatter |

---

## Dependencies

```
github.com/chromedp/cdproto              Chrome DevTools Protocol types
github.com/chromedp/chromedp              Chrome headless automation
github.com/charmbracelet/lipgloss         CLI styling (colors, panels, formatting)
github.com/charmbracelet/huh              Interactive form UI (init wizard)
github.com/charmbracelet/bubbles          Bubble tea components (spinner frames)
github.com/charmbracelet/bubbletea        TUI framework (huh dependency)
github.com/fsnotify/fsnotify              File system watcher (watch mode)
github.com/mattn/go-isatty                Terminal detection (lipgloss dep)
github.com/muesli/termenv                 Terminal profiles, ASCII fallback (NoColor)
github.com/spf13/cobra                    CLI framework
github.com/spf13/pflag                    Flag parsing (cobra dep)
github.com/yuin/goldmark                 Markdown parser
github.com/yuin/goldmark-meta            YAML frontmatter for goldmark
gopkg.in/yaml.v3                          YAML config file parsing
```

---

## Known Fixes & Gotchas

1. **CSS/template paths in config** resolved relative to config file directory, NOT CWD
2. **CLI flag paths** (`--css`, `--template`) resolved via `filepath.Abs()`
3. **`FindConfig()`** returns absolute path (avoids `go run` on Windows issues)
4. **`os.ReadFile` failures** in `WithConfigCSSAndTemplate` silently swallowed unless verbose
5. **Options order**: `WithVerbose` before `WithConfigCSSAndTemplate` for file-read error logging
6. **No auto-detect of CSS/template files** — must be explicitly configured in YAML or CLI
7. **`WithComponent` appends** (no longer replaces parser)
8. **Config without CSS/Template fields** → uses embedded assets or theme
9. **Source default** is `"book"` (not CWD) — aligns with `init` scaffold
10. **Spinner deadlock fixed**: `Done()`/`Fail()` used `<-s.result` but goroutine never sent. Replaced with `ack` channel — goroutine closes `ack` on exit (via defer), `Done`/`Fail` wait on `<-s.ack` before printing
11. **`.gitignore` `pretty-pdf` pattern** was too broad (matched `cmd/pretty-pdf/` dir). Fixed to `/pretty-pdf` (root-anchored only)
12. **Pre-flight order**: Chrome check first (hard failure), then source/output/CSS/template (warnings can pass)
13. **Partial parsing**: `ParseDir`/`ParseAll` never abort on per-file errors. Returns partial docs + `ParseErrors`. Caller decides whether to continue. Frontmatter-not-found logged as debug-level, not error
14. **Watch mode debounce**: 300ms debounce via `time.AfterFunc`. Subsequent events within window reset the timer. Prevents double-builds on editor save (fsnotify fires multiple events per save)
