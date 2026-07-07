# Contributing to go-pretty-pdf

Thanks for your interest in contributing!

## Development setup

Requires Go 1.26+. Chrome/Chromium is optional — `pretty-pdf` auto-downloads
a headless build if none is found (see `chromemgr`).

```bash
git clone https://github.com/sazardev/go-pretty-pdf.git
cd go-pretty-pdf
go mod download
```

## Running tests

```bash
make test          # all tests with race detector
make test-cover    # with HTML coverage report
```

## Linting

```bash
make lint
```

Requires [golangci-lint](https://golangci-lint.run/usage/install/).

## Building

```bash
make build         # dev build to bin/
make build-release # stripped build
```

## Adding a builtin theme

Adding a theme to the CLI/library requires exactly one file and one
registry entry — nothing else to keep in sync:

1. Create `theme/assets/<name>.css`. Set the `--pdf-*` custom properties
   (see the contract documented at the top of `theme/assets/base.css`) plus
   any structural CSS deltas (e.g. a bordered `.cover`).
2. In `theme/builtin.go`: add a `//go:embed` var, a `Name<Foo>` constant, an
   entry in the `registry` map (set `Accented: true` if the theme uses its
   accent color as a bold structural element — a cover border, an
   accent-colored blockquote — rather than just for links), and append the
   constant to `order`.
3. Run `go test ./theme/... ./scripts/docsgen/...`.

That's it. The docs website (`scripts/docsgen`) reads colors, fonts, and
the accent treatment straight out of `theme.List()` and each theme's own
CSS at build time — the theme switcher, its swatch colors, and a
downloadable "docs as a PDF" rendered in that theme all appear
automatically on the next `go run ./scripts/docsgen`. There is no
site-side file to hand-edit or duplicate a palette into.

## Code conventions

- MDX frontmatter `id` field must use `[X.Y.Z]` format
- Documents sorted by ID, not filename
- Components registered via `WithComponent()` — never overwrite the parser
- Config file paths resolved relative to config file directory
- Pre-flight checks: Chrome availability checked first (hard failure), then source/output
- Partial parsing: per-file errors collected, never abort the whole parse

## Commit style

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` — new feature
- `fix:` — bug fix
- `docs:` — documentation
- `refactor:` — code restructuring
- `test:` — adding or updating tests
- `ci:` — CI/CD changes
- `chore:` — maintenance

## Pull requests

1. Fork and branch from `main`
2. Add tests for new functionality
3. Run `make lint` and `make test` before submitting
4. Keep PRs focused — one feature or fix per PR
