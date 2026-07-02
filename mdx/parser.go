package mdx

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkHtml "github.com/yuin/goldmark/renderer/html"
	meta "github.com/yuin/goldmark-meta"
)

type Parser struct {
	md         goldmark.Markdown
	components *ComponentRegistry
}

type ParserOption func(*Parser)

func WithComponent(name string, handler ComponentHandler) ParserOption {
	return func(p *Parser) {
		p.components.Register(name, handler)
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
			),
		),
		components: NewComponentRegistry(),
	}
	for _, o := range opts {
		o(p)
	}
	return p
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
	for _, file := range files {
		doc, err := p.ParseFile(file)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	sortDocuments(docs)

	return docs, nil
}

func (p *Parser) ParseFile(path string) (*Document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

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
	for _, path := range paths {
		doc, err := p.ParseFile(path)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	sortDocuments(docs)
	return docs, nil
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
