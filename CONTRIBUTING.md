# Contributing to go-pretty-pdf

Thanks for your interest in contributing!

## Development setup

Requires Go 1.26+ and Chrome/Chromium installed.

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
