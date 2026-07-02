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
cmd/pretty-pdf/          CLI entrypoint (cobra) — build, check, init, version
config/                  YAML config loader — Config struct, Load(), FindConfig(), Default()
pdf.go                   Root package prettypdf — New(), Build(), Validate(), 18 functional options
mdx/                     Parser (goldmark), component transpiler, DefaultValidator, Document type
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

Full pipeline: parse → validate → compose → render.

Flags:
| Flag | Type | Overrides config |
|------|------|-----------------|
| `--config` | string | — |
| `--source`, `-s` | string | `cfg.Source` |
| `--out`, `-o` | string | `cfg.Output` |
| `--title` | string | `cfg.Title` |
| `--subtitle` | string | `cfg.Subtitle` |
| `--author` | string | `cfg.Author` |
| `--theme` | string | `cfg.Theme` |
| `--css` | string | `cfg.CSS` |
| `--template` | string | `cfg.Template` |
| `--timeout` | string | `cfg.Render.Timeout` |
| `--verbose`, `-v` | bool | — (no config equivalent) |

**Flow** (`main.go:270`):
```
runBuild()
  → loadConfig(cmd)                    # main.go:111
      → config.Default()               # start with code defaults
      → cfgFile ? Load(cfgFile) : FindConfig() + Load()
      → resolve CSS/template paths (config dir or filepath.Abs)
      → CLI flag overrides (if cmd.Flags().Changed("flag"))
      → verbose warnings for missing CSS/template files
      → return cfg
  → buildOpts(cfg)                     # main.go:195
      → WithVerbose(verbose)
      → WithConfig(cfg)
      → WithConfigCSSAndTemplate(cfg)
      → WithVars(cfg.Vars)             # if vars exist
      → WithValidator(DefaultValidator configured from cfg.Lint)
      → WithTheme(theme.X)             # if cfg.Theme != "default"
      → WithTimeout(d)                 # if cfg.Render.Timeout set
      → WithPaperSize(w, h)            # a4/letter/legal from cfg.Render.Paper
      → WithRenderMargins(t,b,l,r)     # parsed from CSS-unit strings (mm,cm,in,pt,px)
      → WithHeaderTitle(cfg.Render.HeaderTitle)
  → prettypdf.New(opts...)
  → pdf.Build(ctx)
  → "PDF generated: out.pdf"
```

### `pretty-pdf check`

Parse + validate only (no compose/render). Uses `DefaultValidator`.

Flags: `--config`, `--source`, `--strict`, `--verbose`

- `--strict`: promotes heading depth warnings to errors
- `--verbose`: prints warnings during config loading

Output format: `[WARN]` / `[ERROR]` per finding, summary `N error(s), M warning(s)`.
Exit code 1 if errors > 0.

### `pretty-pdf init [dir]`

Scaffolds a new book directory with embedded assets (`cmd/pretty-pdf/initassets/*`).
Default dir: `book`.

Creates:
- `go-pretty-pdf.yml` (default config)
- `[1.0.0]-introduction.mdx`
- `[1.1.0]-getting-started.mdx`
- `[1.1.1]-installation.mdx` (has `{{product}}` / `{{company}}` variable example)

Fails if target directory already exists.

### `pretty-pdf version`

Prints `pretty-pdf <version>`. Version defaults to `"dev"`, override with `-ldflags "-X main.version=X.Y.Z"`.

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
| `mdx/` (vars) | 2 | Vars substitution in text, Vars with components |
| `config/` | 4 | Default(), Load(), LoadDefaultsOnMissingKeys(), FindConfig() |

---

## Key File Reference

| File | Lines | Role |
|------|-------|------|
| `cmd/pretty-pdf/main.go` | 429 | CLI: 4 commands, `loadConfig()`, `buildOpts()`, `parseCSSUnit()`, embedded init assets |
| `cmd/pretty-pdf/initassets/*` | 4 files | Scaffold templates for `init` command |
| `pdf.go` | ~190 | Root API, 18 options, Build/Validate pipeline, WithComponent fix |
| `config/config.go` | 82 | Config struct, YAML Load(), FindConfig(), Default() |
| `config/config_test.go` | 145 | 4 config tests |
| `mdx/parser.go` | ~155 | Goldmark parser, variable substitution, RegisterComponent |
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
