GOCMD      = go
MAIN       = ./cmd/pretty-pdf
BINARY     = pretty-pdf
BUILD_DIR  = bin
VERSION    = $(shell git describe --tags --always --dirty 2>nul || echo dev)
LDFLAGS    = -s -w -X github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf.version=$(VERSION)

.PHONY: lint test test-verbose test-cover build build-release clean

lint:
	golangci-lint run --timeout=5m

test:
	$(GOCMD) test -race ./...

test-verbose:
	$(GOCMD) test -race -v ./...

test-cover:
	$(GOCMD) test -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

build:
	if not exist "$(BUILD_DIR)" mkdir "$(BUILD_DIR)"
	$(GOCMD) build -ldflags="$(LDFLAGS)" -o "$(BUILD_DIR)\$(BINARY)" $(MAIN)

build-release:
	if not exist "$(BUILD_DIR)" mkdir "$(BUILD_DIR)"
	$(GOCMD) build -ldflags="$(LDFLAGS)" -o "$(BUILD_DIR)\$(BINARY)" $(MAIN)

release-dry-run:
	goreleaser release --clean --skip=publish --snapshot

clean:
	if exist "$(BUILD_DIR)" rmdir /s /q "$(BUILD_DIR)"
	if exist coverage.out del /q coverage.out
	if exist coverage.html del /q coverage.html
	if exist out.pdf del /q out.pdf
