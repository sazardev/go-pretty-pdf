# go-pretty-pdf — System Memory

## Purpose
Transform MDX source files into print-ready PDF via headless Chrome.
Both a Go library (`prettypdf`) and CLI tool (`pretty-pdf`).

---

## Pipeline

```
Parse MDX (goldmark) → Transpile custom components → Compose HTML (TOC + template + CSS) → Render PDF (headless Chrome)
```

---

## Package Map

```
cmd/pretty-pdf/          CLI entrypoint (cobra) — build, check, init, version
config/                  YAML config loader — Config struct, Load(), FindConfig(), Default()
pdf.go                   Root package prettypdf — New(), Build(), Validate(), functional options
mdx/                     Parser (goldmark), component transpiler, DefaultValidator, Document type
compose/                 HTML composition — template.html + print.css (go:embed), TOC builder
render/                  Chrome headless PDF rendering via chromedp
theme/                   Theme struct with Default and Minimal built-in themes
```

---

## Config File (`go-pretty-pdf.yml`)

Loaded from cwd (or `--config` path). CLI flags override config values. Precedence: **CLI > config > defaults**.

### Fields

```yaml
title: "My Book"           # Document title
subtitle: "A Guide"        # Document subtitle
author: "Jane Doe"         # Author name
source: book               # MDX source dir (default: book)
output: out.pdf            # Output PDF path (default: out.pdf)
theme: default             # Theme: default | minimal

css: custom.css            # Path to CSS file (relative to config file dir)
template: template.html    # Path to HTML template (relative to config file dir)

# Variable substitution — use {{var_name}} in MDX content & frontmatter
vars:
  api_version: "v2.1"
  company: "Acme Corp"

lint:
  require_frontmatter: [id, title]     # Required YAML frontmatter fields
  require_id_format: "[X.Y.Z]"         # ID format validation
  no_duplicate_ids: true               # Fail on duplicate [X.Y.Z] IDs
  max_heading_depth: 3                 # Max heading level (h4+=warning)
  require_lowercase_filenames: false   # Enforce lowercase .mdx names
  check_broken_links: false            # Detect broken internal links

render:
  timeout: 60s             # Render timeout (Go duration string)
  paper: a4                # a4 | letter | legal
  margin_top: 20mm         # CSS-like units: mm, cm, in, pt, px
  margin_bottom: 15mm
  margin_left: 15mm
  margin_right: 15mm
  header_title: "{{title}}" # PDF header text
```

### Defaults

```go
Source: "book", Output: "out.pdf", Title: "Document", Author: "go-pretty-pdf"
Lint: require [id, title], noDuplicateIDs=true, maxHeadingDepth=3
Render: timeout=60s, paper=A4, margins 0.8/0.8/0.6/0.6 inches
```

### CSS/Template Path Resolution

**Critical**: In `loadConfig()` at `main.go:134`, CSS and template file paths in the YAML config are resolved **relative to the config file's directory**, not CWD. This prevents breakage when running from a different working directory (e.g., `go run ./cmd/pretty-pdf build` from repo root while config is in a subdirectory).

CLI flag paths (`--css`, `--template`) are resolved via `filepath.Abs()` in `loadConfig()`.

`FindConfig()` returns an **absolute path** via `filepath.Abs()` to ensure reliable path resolution regardless of how the program is invoked.

---

## CLI Commands

### `pretty-pdf init [dir]`
Scaffolds `book/` (or custom dir) with:
- `go-pretty-pdf.yml` — default config
- `[1.0.0]-introduction.mdx` — intro chapter
- `[1.1.0]-getting-started.mdx` — getting started
- `[1.1.1]-installation.mdx` — with variable substitution example

Embedded via `//go:embed initassets/*` in `cmd/pretty-pdf/main.go:19`.

### `pretty-pdf build`
Full pipeline: parse → validate → compose → render.
Flags: `--config`, `--source`, `--out`, `--title`, `--subtitle`, `--author`,
`--theme`, `--css`, `--template`, `--timeout`, `--verbose`.

Config loading in `loadConfig()`:
1. If `--config` set, resolve relative to CWD and load that file
2. Else look for `go-pretty-pdf.yml` in cwd
3. Apply defaults, then config values, then CLI flag overrides

### `pretty-pdf check`
Parse + validate only. Uses `DefaultValidator.ValidateAll()`.
`--strict` promotes heading depth warnings to errors.
`--verbose` prints warnings and extra info.

### `pretty-pdf version`
Prints version string (default `"dev"`, overridden via `-ldflags`).

---

## Core API (`pdf.go`)

### Options

| Option | Sets | Default |
|--------|------|---------|
| `WithSourceDir(dir)` | source | `"book"` |
| `WithOutputFile(path)` | output | `"out.pdf"` |
| `WithTitle(t)` | compose title | `"Document"` |
| `WithSubtitle(s)` | compose subtitle | `""` |
| `WithAuthor(a)` | compose author | `"go-pretty-pdf"` |
| `WithCSS(css)` | compose CSS content | embedded |
| `WithTemplate(html)` | compose template | embedded |
| `WithTheme(t)` | CSS + Template from theme | default |
| `WithComponent(name, handler)` | registers component **(appends)** | DeepDive, Warning, Axiom |
| `WithValidator(v)` | validator | nil (none) |
| `WithTimeout(d)` | render timeout | 60s |
| `WithHeaderTitle(t)` | PDF header title | compose Title |
| `WithVerbose(bool)` | verbose stdout logging | false |
| `WithVars(map)` | parser variable substitution | none |
| `WithRenderMargins(t,b,l,r)` | PDF margins in inches | 0.8/0.8/0.6/0.6 |
| `WithPaperSize(w,h)` | PDF paper in inches | A4 (8.27×11.69) |
| `WithConfig(*config.Config)` | source, output, title, subtitle, author | — |
| `WithConfigCSSAndTemplate(*config.Config, configDir, verbose)` | CSS & template file paths, theme | — |

**`WithConfigCSSAndTemplate`** reads CSS/template files from disk using paths relative to `configDir`. Prints `os.ReadFile` errors when verbose is true. Maps theme name ("default"|"minimal") to theme constants.

### WithComponent — appends (not replaces)
Register via `p.parser.RegisterComponent(name, handler)` which appends without losing previously registered components or the built-in DeepDive/Warning/Axiom.

### Default source changed
From `"."` (cwd) to `"book"` — aligns with "just create a book folder" DX.

---

## MDX Parser (`mdx/parser.go`)

### Variable Substitution
`Parser.substituteVars(raw)` runs **before** goldmark parsing, so variables can appear anywhere (frontmatter YAML values, body text, code blocks, component attributes).

Substitution syntax: `{{var_name}}` in the raw MDX text. The value in the vars map replaces the placeholder.

### Components
Registered in `ComponentRegistry`. Built-in: `DeepDive`, `Warning`, `Axiom`.
Transpiled via regex after goldmark output.
`RegisterComponent()` on Parser allows non-destructive registration from outside the package.

### ParserOption
- `WithComponent(name, handler)` — register a component handler
- `WithVars(vars map[string]string)` — set variables for substitution

---

## Validator (`mdx/validator.go`)

### `DefaultValidator`
Rules:

| Rule | Condition | Severity |
|------|-----------|----------|
| Required frontmatter field | Field missing or empty | Error |
| ID format | Doesn't match `^\[\d+\.\d+\.\d+\]$` | Error |
| Duplicate IDs | Same `[X.Y.Z]` in >1 file | Error |
| Heading depth | HTML has `<hN>` with N > MaxHeadingDepth | Warning (error with `--strict`) |

`ValidateAll(docs)` runs `Validate` per doc, then checks cross-doc rules (duplicates).

`ValidationError` has `.File`, `.Field`, `.Message` fields.

---

## Render (`render/render.go`)

Headless Chrome via `chromedp`. Renders HTML encoded as `data:text/html;...` URI.
PDF features: `PrintBackground`, `DisplayHeaderFooter`, `GenerateDocumentOutline`,
`GenerateTaggedPDF`, custom margins, A4 default.

Paper sizes from CLI flags: A4 (8.27×11.69), Letter (8.5×11), Legal (8.5×14).

---

## Testing

```bash
go test ./...                    # 26 tests: mdx, config, validator, vars
go test ./mdx/... -v -run TestDefaultValidatorValidate  # Specific validator
go test ./config/... -v          # Config loading tests
```

---

## Project Structure (Key Files)

| File | Lines | Role |
|------|-------|------|
| `pdf.go` | 176 | Root API, options, Build/Validate pipeline |
| `config/config.go` | 82 | Config struct, YAML load, defaults, file find |
| `config/config_test.go` | 112 | Config tests |
| `cmd/pretty-pdf/main.go` | 291 | CLI build/check/init/version, config loading, CSS unit parser |
| `cmd/pretty-pdf/initassets/*` | 4 files | Scaffold templates for `init` command |
| `mdx/parser.go` | 155 | Goldmark parser, var substitution, registration |
| `mdx/validator.go` | 102 | DefaultValidator with lint rules |
| `mdx/validator_test.go` | 252 | Validator + variable tests |
| `mdx/mdx.go` | 133 | Document type, frontmatter accessors, ID utilities |
| `mdx/component.go` | 104 | Component transpilation |
| `compose/compose.go` | 107 | HTML composition, template execution, keywords |
| `compose/toc.go` | 45 | TOC builder from [X.Y.Z] hierarchy |
| `compose/assets/template.html` | 29 | Default HTML template (go:embed) |
| `compose/assets/print.css` | 277 | Default print CSS (go:embed) |
| `render/render.go` | 126 | Chrome headless PDF render |
| `theme/theme.go` | 87 | Theme struct, Default & Minimal themes |
| `examples/full-demo/config.yml` | — | Full-featured demo config with custom CSS+template |
| `examples/full-demo/custom.css` | — | Demo custom CSS overriding default |
| `examples/full-demo/custom-template.html` | — | Demo custom template overriding default |

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

- **CSS/template paths in config** are resolved relative to the config file's directory, not CWD (fix applied in `loadConfig()`)
- **`FindConfig()` returns absolute path** to avoid issues with `go run` on Windows
- **`os.ReadFile` failures** in `WithConfigCSSAndTemplate` are silently swallowed unless verbose is true — enables graceful fallback to embedded assets
- **`WithComponent` must be called after `WithVars`** if both are used, since `WithVars` creates a fresh `mdx.NewParser` internally (though both now use `RegisterComponent` for non-destructive addition)
- **Options order matters** in `New()`: `WithVerbose` should come before `WithConfigCSSAndTemplate` for verbose logging of file read errors
