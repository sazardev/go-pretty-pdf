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
| `--theme` | `"default"` | Theme name (builtin, custom, or a `.theme.yml`/`.css` path) — see [Themes](#themes) |
| `--css` | `""` | Custom CSS file path (overrides the theme entirely) |
| `--template` | `""` | Custom HTML template file path (overrides the theme's HTML) |
| `--timeout` | `""` | Render timeout (e.g. `30s`, `1m`) |
| `--json` | `false` | Output as JSON |
| `--no-cover` | `false` | Omit the cover page |
| `--no-toc` | `false` | Omit the table of contents |
| `--no-page-numbers` | `false` | Omit page numbers |
| `--no-header` | `false` | Omit the running page header |
| `--color-primary` | `""` | Theme override: primary color (e.g. `#1a56db`) |
| `--color-accent` | `""` | Theme override: accent color |
| `--color-text` | `""` | Theme override: body text color |
| `--color-muted` | `""` | Theme override: muted/caption text color |
| `--color-bg` | `""` | Theme override: page background color |
| `--font-heading` | `""` | Theme override: heading font family |
| `--font-body` | `""` | Theme override: body font family |
| `--font-code` | `""` | Theme override: code font family |
| `--density` | `""` | Spacing density: `compact`, `normal`, or `relaxed` |
| `--allow-network-fonts` | `false` | Allow fetching Google Fonts declared by the theme (enables network access) |

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

### `theme`

List, inspect, and manage themes.

```
pretty-pdf theme list
pretty-pdf theme show <name>
pretty-pdf theme new <name> [flags]
pretty-pdf theme add <path> [flags]
```

#### `theme list`

Prints every builtin theme (name + description) followed by any custom
themes discovered in `./themes/` (project) and the global themes directory
(`~/.config/pretty-pdf/themes` on Linux, via `os.UserConfigDir()`).

#### `theme show <name>`

Resolves a theme (builtin, custom, or a `.theme.yml`/`.css` path) with no
customization and prints its final, fully-assembled CSS to stdout — useful
to inspect a theme or pipe it somewhere (`pretty-pdf theme show dark > dark.css`).

#### `theme new <name>`

Scaffolds a starter `<name>.theme.yml` you can hand-edit.

| Flag | Default | Description |
|---|---|---|
| `--from` | `"default"` | Builtin theme to base the scaffold on |
| `--global` | `false` | Write to the global themes directory instead of `./themes` |

Refuses to overwrite an existing file.

#### `theme add <path>`

Imports an existing `.theme.yml` or raw `.css` file as a managed custom
theme (a loose `.css` file is wrapped into a minimal `.theme.yml` with
`extends: default` and the file's content as its `css:` block).

| Flag | Default | Description |
|---|---|---|
| `--as` | `""` | Name to register the imported theme under (default: derived from the file name) |
| `--global` | `false` | Copy to the global themes directory instead of `./themes` |

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
theme: corporate
css: custom.css
template: custom.html
vars:
  version: "1.0"
  year: "2026"

theme_options:
  colors:
    primary: "#1a56db"
    accent: "#0ea5e9"
  fonts:
    heading: "Georgia, serif"
    google_fonts: ["Inter:400,600"]   # only fetched with allow_network_fonts: true
  sections:
    cover: true
    toc: true
    page_numbers: true
    header: true
  density: normal        # compact | normal | relaxed
  allow_network_fonts: false

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
| `theme` | `""` | Theme name (builtin, custom, or a `.theme.yml`/`.css` path) — see [Themes](#themes) |
| `css` | `""` | Path to custom CSS file (overrides the theme entirely) |
| `template` | `""` | Path to custom HTML template file (overrides the theme's HTML) |
| `vars` | `{}` | Template variables for `{{key}}` substitution |
| `theme_options` | `{}` | Theme customization — see [Themes](#themes) |

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

Eight built-in themes are available, each a palette/typography layer over a
shared structural stylesheet (`theme/assets/base.css`):

| Theme | Category | Description |
|---|---|---|
| `default` | professional | Clean, professional look that fits any technical document. |
| `minimal` | minimal | Stripped down: smaller type, no borders, maximum simplicity. |
| `modern` | professional | Sans-serif with generous whitespace and bold accent underlines. |
| `classic` | editorial | Serif, traditional book layout — ink on paper. |
| `corporate` | professional | Structured blue/gray palette for client-facing reports. |
| `dark` | dark | Dark background with light text. Best for on-screen PDFs. |
| `academic` | academic | Formal serif layout for theses, papers, and reports. |
| `editorial` | editorial | Magazine-style display headings and pull-quote blockquotes. |

Run `pretty-pdf theme list` to see this list plus any custom themes, and
`pretty-pdf theme show <name>` to print a theme's final resolved CSS.

### Customizing a theme without writing CSS

`theme_options` (config) or the matching `--color-*`/`--font-*`/`--density`/
`--no-*` flags (CLI) customize any theme — builtin or custom — without
touching CSS:

```bash
pretty-pdf build --theme corporate \
  --color-primary "#0ea5e9" --font-heading "Georgia, serif" \
  --no-cover --no-page-numbers --density compact
```

| `theme_options` field | Description |
|---|---|
| `colors.primary/accent/text/muted/background` | CSS custom properties for the theme's palette |
| `fonts.heading/body/code` | Font-family overrides (system-safe stacks recommended) |
| `fonts.google_fonts` | Google Fonts family names (e.g. `["Inter:400,600"]`) — only fetched when `allow_network_fonts: true`, since network access is otherwise blocked during rendering |
| `sections.cover/toc/page_numbers/header` | `true`/`false`/unset (unset = theme's own default) |
| `density` | `compact`, `normal`, or `relaxed` — adjusts line-height and a handful of spacing rules |
| `allow_network_fonts` | Enables outbound network access during rendering so `fonts.google_fonts` can be fetched |

Section toggles set via `--no-cover`/`--no-toc`/`--no-page-numbers`/
`--no-header` only apply to the default HTML template; a custom `--template`
owns its own HTML and must implement any toggles itself (the default
template gates its cover block on `{{if .ShowCover}}`).

### Custom themes

A custom theme is a `<name>.theme.yml` file that extends a builtin theme:

```yaml
name: my-report
description: "Client report with a teal accent"
extends: corporate

colors:
  accent: "#0d9488"
fonts:
  heading: "Georgia, serif"
sections:
  page_numbers: false
density: normal

css: |
  /* raw CSS appended last — wins over everything above */
  .cover h1 { text-transform: uppercase; }
```

Custom themes are discovered by name in `./themes/` (project-local, checked
first) and then in the global themes directory
(`~/.config/pretty-pdf/themes` on Linux). Use them the same way as a
builtin: `--theme my-report` or `theme: my-report` in config.

Manage them with:

```bash
pretty-pdf theme new my-report --from corporate   # scaffold ./themes/my-report.theme.yml
pretty-pdf theme add ./some-theme.theme.yml        # import an existing theme file
pretty-pdf theme add ./some.css --as my-report     # or wrap a plain CSS file
pretty-pdf theme list                              # see builtins + everything discovered
pretty-pdf theme show my-report                    # print the fully resolved CSS
```

A `--theme` value ending in `.theme.yml`/`.css` is treated as a direct file
path instead of a name, so you can also point straight at a file without
installing it into a themes directory.

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
