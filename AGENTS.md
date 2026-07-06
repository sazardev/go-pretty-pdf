# AGENTS.md

## Overview

`go-pretty-pdf` transforms a directory of MDX files into a print-ready PDF via headless Chrome.
It is both a Go library (`github.com/sazardev/go-pretty-pdf`) and a CLI tool.

## Commands

```bash
go run ./cmd/pretty-pdf build --source ./docs --out out.pdf
go run ./cmd/pretty-pdf validate --source ./docs
go test ./...            # only mdx/ has tests
go test ./mdx/... -run TestParserParseFile -v
go build ./cmd/pretty-pdf
```

## Architecture

```
cmd/pretty-pdf/    CLI entrypoint (cobra)
pdf.go             Root package — public API: New(), Build(), Validate(), functional options
mdx/               MDX parser (goldmark-based), custom component transpiler, validator interface
compose/           HTML composition: TOC, go:embed'd template.html + print.css
render/            Chrome headless PDF rendering via chromedp
theme/             Theme engine: 8 builtin themes over a shared base.css, custom .theme.yml themes, section toggles
```

### Pipeline

`Parse MDX` → `Transpile custom components` → `Compose HTML` (embed assets, TOC) → `Render PDF` (Chrome headless)

## Requirements

- **Chrome or Chromium must be installed** on the system for rendering to work.
- Go 1.26+

## Key conventions

- MDX frontmatter is **required** — at minimum an `id` field formatted as `[X.Y.Z]` (e.g. `[1.0.0]`).
- Documents are sorted by their `[X.Y.Z]` ID key, **not** by filename or filesystem order.
- Built-in custom components: `<DeepDive>`, `<Warning>`, `<Axiom>`. Additional components registrable via `WithComponent()`.
- Embedded assets live in `compose/assets/` and are loaded at compile time via `//go:embed` — no runtime file reads.
- Common frontmatter fields: `id`, `title`, `subtitle`, `tags`, `difficulty`, `status`, `completeness`, `depends_on`.
