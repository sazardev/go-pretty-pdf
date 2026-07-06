package mdx

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const (
	testFileA    = "a.mdx"
	testSecondID = "[2.0.0]"
)

func TestParserParseDirNoMDXFiles(t *testing.T) {
	dir := t.TempDir()

	p := NewParser()
	_, err := p.ParseDir(dir)
	if err == nil {
		t.Fatal("expected error for directory with no .mdx files")
	}
	if !contains(err.Error(), "no .mdx files found") {
		t.Errorf("expected 'no .mdx files found' error, got: %v", err)
	}
}

func TestParserParseFileMissingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-frontmatter.mdx")
	if err := os.WriteFile(path, []byte("# Just a heading\n\nNo frontmatter here.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewParser()
	_, err := p.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
	if !contains(err.Error(), "missing frontmatter") {
		t.Errorf("expected 'missing frontmatter' error, got: %v", err)
	}
}

func TestParserParseDirPartialFailure(t *testing.T) {
	dir := t.TempDir()

	valid := `---
id: "[1.0.0]"
title: Valid
---

# Valid
`
	invalid := "# No frontmatter\n"

	if err := os.WriteFile(filepath.Join(dir, "valid.mdx"), []byte(valid), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "invalid.mdx"), []byte(invalid), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewParser()
	docs, err := p.ParseDir(dir)

	if len(docs) != 1 {
		t.Fatalf("expected 1 successfully parsed doc, got %d", len(docs))
	}
	if err == nil {
		t.Fatal("expected a non-nil error describing the partial failure")
	}
	var parseErrs ParseErrors
	if !errors.As(err, &parseErrs) {
		t.Fatalf("expected error to be a ParseErrors, got %T", err)
	}
	if len(parseErrs) != 1 {
		t.Fatalf("expected 1 parse error, got %d", len(parseErrs))
	}
}

func TestParseErrorsError(t *testing.T) {
	var empty ParseErrors
	if got := empty.Error(); got != "" {
		t.Errorf("expected empty string for 0 errors, got %q", got)
	}

	single := ParseErrors{{File: testFileA, Err: errors.New("boom")}}
	if got := single.Error(); !contains(got, testFileA) || !contains(got, "boom") {
		t.Errorf("expected single error message to mention file and cause, got %q", got)
	}

	multi := ParseErrors{
		{File: testFileA, Err: errors.New("boom")},
		{File: "b.mdx", Err: errors.New("bang")},
	}
	got := multi.Error()
	if !contains(got, "2 file(s) failed to parse") {
		t.Errorf("expected aggregate message for multiple errors, got %q", got)
	}
}

func TestParseFileErrorUnwrap(t *testing.T) {
	cause := errors.New("underlying failure")
	pfe := ParseFileError{File: "x.mdx", Err: cause}

	if !errors.Is(pfe, cause) {
		t.Error("expected errors.Is to find the wrapped cause via Unwrap")
	}
	if !contains(pfe.Error(), "x.mdx") {
		t.Errorf("expected error message to include file name, got %q", pfe.Error())
	}
}

func TestParserParseAll(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		testFileA: `---
id: "[2.0.0]"
title: Second
---

# Second
`,
		"b.mdx": `---
id: "[1.0.0]"
title: First
---

# First
`,
	}

	paths := make([]string, 0, len(files))
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}

	p := NewParser()
	docs, err := p.ParseAll(paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(docs))
	}
	if docs[0].ID() != defaultIDValue || docs[1].ID() != testSecondID {
		t.Errorf("expected docs sorted by ID, got %s then %s", docs[0].ID(), docs[1].ID())
	}
}

func TestParserParseAllPartialFailure(t *testing.T) {
	dir := t.TempDir()
	goodPath := filepath.Join(dir, "good.mdx")
	if err := os.WriteFile(goodPath, []byte(`---
id: "[1.0.0]"
title: Good
---

# Good
`), 0644); err != nil {
		t.Fatal(err)
	}
	missingPath := filepath.Join(dir, "missing.mdx")

	p := NewParser()
	docs, err := p.ParseAll([]string{goodPath, missingPath})
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc from the successful path, got %d", len(docs))
	}
	if err == nil {
		t.Fatal("expected error for the missing file")
	}
}
