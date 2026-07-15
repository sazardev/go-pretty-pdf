package mdx

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkHtml "github.com/yuin/goldmark/renderer/html"
)

type ParseFileError struct {
	File string
	Err  error
}

func (e ParseFileError) Error() string {
	return fmt.Sprintf("%s: %v", e.File, e.Err)
}

func (e ParseFileError) Unwrap() error {
	return e.Err
}

type ParseErrors []ParseFileError

func (pe ParseErrors) Error() string {
	if len(pe) == 0 {
		return ""
	}
	if len(pe) == 1 {
		return pe[0].Error()
	}
	return fmt.Sprintf("%d file(s) failed to parse (first: %v)", len(pe), pe[0])
}

type Parser struct {
	md         goldmark.Markdown
	components *ComponentRegistry
	varsMu     sync.RWMutex
	vars       map[string]string
}

type ParserOption func(*Parser)

func WithComponent(name string, handler ComponentHandler) ParserOption {
	return func(p *Parser) {
		p.components.Register(name, handler)
	}
}

func WithVars(vars map[string]string) ParserOption {
	return func(p *Parser) {
		p.vars = vars
	}
}

func NewParser(opts ...ParserOption) *Parser {
	p := &Parser{
		md: goldmark.New(
			goldmark.WithExtensions(
				meta.New(meta.WithStoresInDocument()),
				extension.GFM,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				goldmarkHtml.WithUnsafe(),
				// XHTML-style self-closing void elements (<img/>, <hr/>,
				// <br/>) are still perfectly valid HTML5 — Chrome's print
				// pipeline doesn't care — but they're required for the
				// epub package's chapter files to be well-formed XHTML, so
				// this is set globally rather than per-consumer.
				goldmarkHtml.WithXHTML(),
			),
		),
		components: NewComponentRegistry(),
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func (p *Parser) RegisterComponent(name string, handler ComponentHandler) {
	p.components.Register(name, handler)
}

func (p *Parser) SetVars(vars map[string]string) {
	p.varsMu.Lock()
	defer p.varsMu.Unlock()
	p.vars = vars
}

func (p *Parser) ParseDir(dir string) ([]*Document, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".mdx") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking source dir: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no .mdx files found in %s", dir)
	}

	var docs []*Document
	var parseErrs ParseErrors

	for _, file := range files {
		doc, err := p.ParseFile(file)
		if err != nil {
			parseErrs = append(parseErrs, ParseFileError{File: file, Err: err})
			continue
		}
		docs = append(docs, doc)
	}

	sortDocuments(docs)

	if len(parseErrs) > 0 {
		if len(docs) == 0 {
			return nil, parseErrs
		}
		return docs, parseErrs
	}

	return docs, nil
}

func (p *Parser) ParseFile(path string) (*Document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	raw = p.substituteVars(raw)

	ctx := parser.NewContext()
	var buf bytes.Buffer

	if err := p.md.Convert(raw, &buf, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	frontmatter := meta.Get(ctx)
	if frontmatter == nil {
		return nil, fmt.Errorf("%s: missing frontmatter", path)
	}

	html := buf.String()
	html = p.components.Transpile(html)

	return &Document{
		Path:        path,
		Frontmatter: frontmatter,
		HTML:        html,
	}, nil
}

func (p *Parser) ParseAll(paths []string) ([]*Document, error) {
	var docs []*Document
	var parseErrs ParseErrors

	for _, path := range paths {
		doc, err := p.ParseFile(path)
		if err != nil {
			parseErrs = append(parseErrs, ParseFileError{File: path, Err: err})
			continue
		}
		docs = append(docs, doc)
	}

	sortDocuments(docs)

	if len(parseErrs) > 0 {
		if len(docs) == 0 {
			return nil, parseErrs
		}
		return docs, parseErrs
	}

	return docs, nil
}

// substituteVars replaces every {{key}} placeholder with its configured
// value in a single pass over raw via strings.Replacer, rather than looping
// over the vars map and doing a sequential ReplaceAll per key. Two things
// would otherwise go wrong: map iteration order is randomized in Go, and a
// sequential ReplaceAll rescans the *already-substituted* text on each
// iteration, so one var's value can itself contain another var's
// placeholder and get expanded again — making the result depend on
// iteration order and change from run to run for identical input. A single
// Replacer pass finds all matches against the original text and expands
// each exactly once, so the result is deterministic and placeholders never
// chain into each other.
func (p *Parser) substituteVars(raw []byte) []byte {
	p.varsMu.RLock()
	vars := p.vars
	p.varsMu.RUnlock()

	if len(vars) == 0 {
		return raw
	}
	pairs := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		pairs = append(pairs, "{{"+k+"}}", v)
	}
	return []byte(strings.NewReplacer(pairs...).Replace(string(raw)))
}

func sortDocuments(docs []*Document) {
	sort.Slice(docs, func(i, j int) bool {
		ai := docs[i].SortKey()
		aj := docs[j].SortKey()
		if ai[0] != aj[0] {
			return ai[0] < aj[0]
		}
		if ai[1] != aj[1] {
			return ai[1] < aj[1]
		}
		return ai[2] < aj[2]
	})
}
