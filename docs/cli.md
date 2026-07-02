# CLI Reference

## Overview

`pretty-pdf` transforms a directory of MDX files into a print-ready PDF via headless Chrome.
Documents are sorted by their `[X.Y.Z]` frontmatter ID, not by filename.

GitHub: <https://github.com/sazardev/go-pretty-pdf>

## Requirements

- **Chrome or Chromium** must be installed on the system for PDF rendering.
- Go 1.26+ (if building from source).

## Usage

```
pretty-pdf [command] [flags]
```

## Global Flags

| Flag | Default | Description |
|---|---|---|
| `--config` | `""` | Path to config file |
| `--source` | `"book"` | Source MDX directory |
| `--verbose` | `false` | Verbose output |
| `--no-color` | `false` | Disable colored output |
| `--quiet` | `false` | Suppress non-error output |
| `-h, --help` | | Help for any command |

## Commands

### `build`

Parse MDX files, validate them, compose HTML, and render to PDF.

```
pretty-pdf build [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--out` | `"out.pdf"` | Output PDF path |
| `--title` | `""` | Book title |
| `--subtitle` | `""` | Book subtitle |
| `--author` | `""` | Book author |
| `--theme` | `"default"` | Book theme (`default`, `minimal`) |
| `--css` | `""` | Custom CSS file path |
| `--template` | `""` | Custom HTML template file path |
| `--timeout` | `""` | Render timeout (e.g. `30s`, `1m`) |
| `--json` | `false` | Output as JSON |

#### Build Pipeline

The `build` command runs through these stages:

1. **Parse** — Read and parse all MDX files in the source directory
2. **Validate** — Check frontmatter, duplicate IDs, heading depth, content warnings
3. **Compose** — Assemble HTML with TOC, cover page, and embedded CSS/template
4. **Render** — Generate PDF via headless Chrome

#### Pre-flight Checks

Before the pipeline starts, `build` verifies:

- Chrome/Chromium is available
- Source directory exists
- At least one MDX file is present
- Output directory is writable
- Custom CSS file exists (if specified)
- Custom template file exists (if specified)

---

### `check`

Parse and validate all MDX files without building a PDF. Previously named `validate`.

```
pretty-pdf check [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--strict` | `false` | Treat content warnings as errors |

---

### `init`

Scaffold a new book project with sample MDX files and configuration.

```
pretty-pdf init [directory] [flags]
```

Interactive mode (default): runs a terminal form asking for title, author, theme, source directory.

| Flag | Default | Description |
|---|---|---|
| `--bare` | `false` | Non-interactive init with flags |
| `--title` | `"My Book"` | Book title (for `--bare`) |
| `--author` | `"go-pretty-pdf"` | Book author (for `--bare`) |
| `--theme` | `"default"` | Book theme (for `--bare`) |
| `--json` | `false` | Output as JSON |

---

### `serve`

Parse MDX files, compose HTML, and serve with live reload on file changes. No Chrome required.

```
pretty-pdf serve [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--port` | `8080` | HTTP server port |

Uses Server-Sent Events for live reload. Watches `.mdx`, `.yaml`, and `.yml` files for changes.

---

### `watch`

Watch the source directory for changes and rebuild the PDF on every file change.

```
pretty-pdf watch [flags]
```

Debounces changes by 300ms. Watches `.mdx`, `.yaml`, and `.yml` files. Prints a build/error summary on `Ctrl+C`.

---

### `version`

Print the version number.

```
pretty-pdf version
```

### `completion`

Generate shell completion scripts.

```
pretty-pdf completion [bash|zsh|fish|powershell]
```

| Shell | Install command |
|---|---|
| bash | `pretty-pdf completion bash > /etc/bash_completion.d/pretty-pdf` |
| zsh | `pretty-pdf completion zsh > "${fpath[1]}/_pretty-pdf"` |
| fish | `pretty-pdf completion fish > ~/.config/fish/completions/pretty-pdf.fish` |
| powershell | `pretty-pdf completion powershell > _pretty-pdf.ps1` then `. .\_pretty-pdf.ps1` |

---

## Config File

`go-pretty-pdf.yml` is auto-discovered by walking up from the working directory.
Can also be specified explicitly with `--config`.

### Example

```yaml
title: "My Book"
subtitle: "A journey into MDX-powered PDFs"
author: "Jane Doe"
source: book
output: out.pdf
theme: default
css: custom.css
template: custom.html
vars:
  version: "1.0"
  year: "2026"

lint:
  require_frontmatter:
    - id
    - title
  no_duplicate_ids: true
  max_heading_depth: 3

render:
  timeout: 30s
  paper: A4
  margin_top: 20mm
  margin_bottom: 15mm
  margin_left: 15mm
  margin_right: 15mm
  header_title: "My Book"
```

### Top-level fields

| Field | Default | Description |
|---|---|---|
| `title` | `"Document"` | Book title |
| `subtitle` | `""` | Book subtitle |
| `author` | `"go-pretty-pdf"` | Book author |
| `source` | `"book"` | Source MDX directory |
| `output` | `"out.pdf"` | Output PDF path |
| `theme` | `"default"` | Visual theme (`default`, `minimal`) |
| `css` | `""` | Path to custom CSS file |
| `template` | `""` | Path to custom HTML template file |
| `vars` | `{}` | Template variables for `{{key}}` substitution |

### `lint` fields

| Field | Default | Description |
|---|---|---|
| `require_frontmatter` | `["id", "title"]` | Required frontmatter fields |
| `no_duplicate_ids` | `true` | Reject duplicate document IDs |
| `max_heading_depth` | `3` | Maximum allowed heading depth |

### `render` fields

| Field | Default | Description |
|---|---|---|
| `timeout` | `""` | Chrome render timeout (e.g. `30s`, `1m`) |
| `paper` | `""` | Paper size: `letter`, `legal`, `A4`, or empty for CSS default |
| `margin_top` | `""` | Top margin as CSS unit (`20mm`, `1in`, `10mm`, `2cm`, `12pt`, `96px`) |
| `margin_bottom` | `""` | Bottom margin as CSS unit |
| `margin_left` | `""` | Left margin as CSS unit |
| `margin_right` | `""` | Right margin as CSS unit |
| `header_title` | `""` | Header title in rendered PDF |

---

## Themes

Two built-in themes are available:

| Theme | Description |
|---|---|
| `default` | Clean, professional look. Uses the embedded `print.css` (291 lines) and `template.html` from `compose/assets/`. Full-featured with page numbers, TOC styling, cover page, card components. |
| `minimal` | Stripped down, no extras. Simpler CSS embedded in Go source. Smaller fonts, minimal borders, no page numbers. |

Custom themes can be created by providing `css` and/or `template` paths in config or via `--css`/`--template` CLI flags.

## Template Variables

Available in HTML templates:

| Variable | Description |
|---|---|
| `{{.Title}}` | Book title |
| `{{.Subtitle}}` | Book subtitle |
| `{{.Author}}` | Book author |
| `{{.CSS}}` | Inline CSS string |
| `{{.Body}}` | Composed document body |
| `{{.BuiltAt}}` | Build timestamp |
| `{{.TotalDocs}}` | Number of documents |
| `{{.Keywords}}` | Tags from documents |

## Environment

- `NO_COLOR` environment variable is respected (disables colored output).

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | General error (parsing, validation, rendering, config) |
