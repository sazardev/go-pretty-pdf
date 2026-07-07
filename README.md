# go-pretty-pdf

[![Go Reference](https://pkg.go.dev/badge/github.com/sazardev/go-pretty-pdf.svg)](https://pkg.go.dev/github.com/sazardev/go-pretty-pdf)
[![CI](https://github.com/sazardev/go-pretty-pdf/actions/workflows/ci.yml/badge.svg)](https://github.com/sazardev/go-pretty-pdf/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sazardev/go-pretty-pdf)](https://goreportcard.com/report/github.com/sazardev/go-pretty-pdf)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Transform a directory of MDX files into a beautiful, print-ready PDF via headless Chrome.

**Library + CLI.** Use it as a composable Go library or as a standalone command-line tool.

## Install

### CLI (binary)

```bash
go install github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf@latest
```

### Library

```bash
go get github.com/sazardev/go-pretty-pdf
```

### Requirements

- **Go 1.26+**
- **Chrome or Chromium** — optional. If none is found on your system, `pretty-pdf` automatically downloads and caches a small headless-only Chrome build the first time you run it (like Playwright/Puppeteer do). Already have Chrome installed? It's used as-is, nothing is downloaded. Prefer to control this yourself? Pass `--chrome-path /path/to/chrome` or set `PRETTY_PDF_CHROME_PATH`. Auto-download currently covers linux/amd64, darwin/amd64, darwin/arm64, and windows/amd64 — on linux/arm64 (no official build exists yet) install Chromium via your package manager and point `--chrome-path` at it.

## Quick start

### CLI

```bash
# Scaffold a new book project (interactive wizard)
pretty-pdf init my-book

# Build a PDF
pretty-pdf build --source my-book --out my-book.pdf

# Watch for changes and rebuild
pretty-pdf watch --source my-book --out my-book.pdf

# Validate MDX files
pretty-pdf check --source my-book
```

### Library

```go
package main

import (
	"context"
	"log"

	prettypdf "github.com/sazardev/go-pretty-pdf"
)

func main() {
	pdf, err := prettypdf.New(
		prettypdf.WithSourceDir("./docs"),
		prettypdf.WithOutputFile("output.pdf"),
		prettypdf.WithTitle("My Documentation"),
		prettypdf.WithAuthor("Jane Doe"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := pdf.Build(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

## How it works

```
MDX files → Parse frontmatter & markdown → Transpile components → Compose HTML → Render PDF
```

1. **Parse** — goldmark parses MDX files with YAML frontmatter
2. **Transpile** — custom components (`<DeepDive>`, `<Warning>`, `<Axiom>`) become styled HTML
3. **Compose** — HTML assembled with embedded template + CSS + auto-generated Table of Contents
4. **Render** — headless Chrome prints to PDF with headers, footers, and PDF bookmarks

Documents are sorted by their `[X.Y.Z]` frontmatter ID, not filename.

## Trust model

MDX is parsed with raw HTML passthrough enabled, and custom components
don't escape their inner content — this lets authors embed arbitrary
HTML/CSS for rich documents, but it also means a `.mdx` file can contain
a `<script>` tag that will execute during rendering. By default, headless
Chrome's network access is blocked while rendering (see `WithNetworkAccess`),
so scripts can't exfiltrate data or fetch remote content — but they still
run. **Only build PDFs from MDX you trust.** See [SECURITY.md](SECURITY.md)
for details.

## MDX format

```mdx
---
id: "[1.0.0]"
title: "Getting Started"
subtitle: "A simple introduction"
tags: [example, intro]
difficulty: "beginner"
status: complete
completeness: 100
depends_on: []
---

# Welcome to Your Book

This is the first chapter.

## Variables

You can use {{key}} syntax for variable substitution: running {{product}} v{{version}}.
```

Required frontmatter fields: `id` (format `[X.Y.Z]`), `title`.

## Built-in components

| Component | Usage | Appearance |
|-----------|-------|------------|
| `<DeepDive>` | `<DeepDive title="Details">...</DeepDive>` | Blue info panel |
| `<Warning>` | `<Warning title="Note">...</Warning>` | Orange warning panel |
| `<Axiom>` | `<Axiom>...</Axiom>` | Green italic quote |

Register custom components via `WithComponent()`:

```go
prettypdf.WithComponent("Callout", func(attrs map[string]string, inner string) string {
	level := attrs["level"]
	return fmt.Sprintf(`<div class="callout callout-%s">%s</div>`, level, inner)
})
```

## Configuration

Create a `go-pretty-pdf.yml` in your project:

```yaml
title: "My Book"
subtitle: "A Complete Guide"
author: "Jane Doe"
source: book
output: out.pdf
theme: default

css: custom.css
template: custom-template.html

vars:
  product: "go-pretty-pdf"
  version: "1.0"

lint:
  require_frontmatter: [id, title]
  require_id_format: "[X.Y.Z]"
  no_duplicate_ids: true
  max_heading_depth: 3

render:
  timeout: 60s
  paper: a4
  margin_top: 20mm
  margin_bottom: 20mm
  margin_left: 15mm
  margin_right: 15mm
  header_title: "{{title}}"
```

## Library API

```go
// Constructor with functional options
pdf, err := prettypdf.New(opts...)

// All-in-one build pipeline
pdf.Build(ctx)

// Step-by-step pipeline
docs, _ := pdf.ParseDir()
errs := pdf.ValidateDoc(doc)
html, _ := pdf.ComposeHTML(docs)
pdf.Render(html)

// Validation-only
errs, _ := pdf.Validate(ctx)
```

### Available options

| Option | Description |
|--------|-------------|
| `WithSourceDir(dir)` | MDX source directory (default: `book`) |
| `WithOutputFile(path)` | Output PDF path (default: `out.pdf`) |
| `WithTitle(title)` | Document title |
| `WithSubtitle(sub)` | Document subtitle |
| `WithAuthor(author)` | Document author |
| `WithCSS(css)` | Custom CSS content string |
| `WithTemplate(html)` | Custom HTML template string |
| `WithTheme(t)` | Apply a raw `theme.Theme` (no customization/section toggles) |
| `WithThemeName(name, opts)` | Resolve a theme by name (builtin, custom, or file path) with color/font/section customization |
| `WithComponent(name, handler)` | Register custom MDX component |
| `WithValidator(v)` | Custom validation logic |
| `WithTimeout(d)` | Chrome render timeout (default: 60s) |
| `WithHeaderTitle(t)` | PDF header title |
| `WithVerbose(bool)` | Enable verbose logging |
| `WithVars(map)` | Variable substitution map |
| `WithRenderMargins(t,b,l,r)` | PDF margins in inches |
| `WithPaperSize(w,h)` | Paper size in inches |
| `WithConfig(cfg)` | Apply source/output/title/subtitle/author from config |
| `WithConfigCSSAndTemplate(cfg)` | Load CSS/template from config file paths |
| `WithFullConfig(cfg)` | Apply the entire config struct (source, CSS/template, theme, vars, render settings) in one call |
| `WithNetworkAccess(bool)` | Allow headless Chrome to make network requests while rendering (default: `false`, blocked) |

## Themes

Eight built-in themes, each a palette/typography layer over one shared
structural stylesheet — clean and professional by default, easy to
customize without writing CSS, and extendable with your own custom themes:

`default` &middot; `minimal` &middot; `modern` &middot; `classic` &middot; `corporate` &middot; `dark` &middot; `academic` &middot; `editorial`

```bash
# Pick a theme, tweak colors/fonts/density, drop sections you don't want
pretty-pdf build --theme corporate \
  --color-primary "#0ea5e9" --font-heading "Georgia, serif" \
  --no-cover --no-page-numbers --density compact

# Scaffold your own reusable theme
pretty-pdf theme new my-report --from corporate
pretty-pdf theme list
```

```go
prettypdf.WithThemeName("corporate", theme.Options{
	Colors:   theme.Colors{Primary: "#0ea5e9"},
	Sections: theme.Sections{Cover: theme.BoolPtr(false)},
})
```

Custom themes live in `<name>.theme.yml` files (project-local `./themes/`
or a global themes directory) and `extends` a builtin theme. Full reference,
all customization fields, and the `pretty-pdf theme` command family:
see [docs/cli.md#themes](docs/cli.md#themes).

## CLI reference

```
pretty-pdf build     Build a PDF from MDX source files
pretty-pdf check     Validate MDX files without building
pretty-pdf theme     List, inspect, and manage themes
pretty-pdf init      Scaffold a new book project (interactive wizard)
pretty-pdf watch     Watch for changes and rebuild automatically
pretty-pdf serve     Preview MDX as HTML with live reload (no Chrome required)
pretty-pdf version   Print the version number
```

Run `pretty-pdf <command> --help` for the full flag list of any command.

Global flags: `--config`, `--source`, `--verbose`, `--no-color`, `--quiet`

## License

MIT — see [LICENSE](LICENSE).
