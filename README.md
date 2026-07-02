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
- **Chrome or Chromium** installed on the system (for PDF rendering)

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
| `WithTheme(t)` | Pre-built theme (`theme.Default`, `theme.Minimal`) |
| `WithComponent(name, handler)` | Register custom MDX component |
| `WithValidator(v)` | Custom validation logic |
| `WithTimeout(d)` | Chrome render timeout (default: 60s) |
| `WithHeaderTitle(t)` | PDF header title |
| `WithVerbose(bool)` | Enable verbose logging |
| `WithVars(map)` | Variable substitution map |
| `WithRenderMargins(t,b,l,r)` | PDF margins in inches |
| `WithPaperSize(w,h)` | Paper size in inches |
| `WithConfig(cfg)` | Apply config struct |
| `WithConfigCSSAndTemplate(cfg)` | Load CSS/template from config file paths |

## Themes

Two themes are built in:

- **Default** — embedded `print.css` + `template.html` (professional, feature-rich)
- **Minimal** — stripped-down CSS, system font stack

```go
prettypdf.WithTheme(theme.Minimal)
```

## CLI reference

```
pretty-pdf build     Build a PDF from MDX source files
pretty-pdf check     Validate MDX files without building
pretty-pdf init      Scaffold a new book project (interactive wizard)
pretty-pdf watch     Watch for changes and rebuild automatically
pretty-pdf version   Print the version number
```

Global flags: `--config`, `--source`, `--verbose`, `--no-color`, `--quiet`

## License

MIT — see [LICENSE](LICENSE).
