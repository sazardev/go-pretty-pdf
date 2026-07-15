package compose

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"github.com/sazardev/go-pretty-pdf/mdx"
)

//go:embed assets/template.html
var defaultTemplate string

//go:embed assets/print.css
var defaultCSS string

type Options struct {
	Title     string
	Subtitle  string
	Author    string
	Template  string
	CSS       string
	ShowCover bool
	ShowTOC   bool
}

func DefaultOptions() Options {
	return Options{
		Title:     "Document",
		Subtitle:  "",
		Author:    "go-pretty-pdf",
		ShowCover: true,
		ShowTOC:   true,
	}
}

func ComposeHTML(docs []*mdx.Document, opts Options) (string, error) {
	tmplHTML := opts.Template
	if tmplHTML == "" {
		tmplHTML = defaultTemplate
	}
	css := opts.CSS
	if css == "" {
		css = defaultCSS
	}

	tmpl, err := template.New("book").Parse(tmplHTML)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	keywords := collectKeywords(docs)

	var bodyBuf bytes.Buffer
	if opts.ShowTOC {
		bodyBuf.WriteString(buildTOC(docs))
		bodyBuf.WriteString(`<div style="page-break-before:always; break-before:page;"></div>` + "\n")
	}

	for _, doc := range docs {
		fmt.Fprintf(&bodyBuf, `<section id="%s">`+"\n", template.HTMLEscapeString(mdx.AnchorID(doc.ID())))
		bodyBuf.WriteString(doc.HTML)
		bodyBuf.WriteString("</section>\n")
	}

	data := templateData{
		Title:     opts.Title,
		Subtitle:  opts.Subtitle,
		Author:    opts.Author,
		Keywords:  keywords,
		CSS:       template.CSS(css),
		Body:      template.HTML(bodyBuf.String()),
		BuiltAt:   time.Now().Format("2006-01-02 15:04:05 UTC"),
		TotalDocs: len(docs),
		ShowCover: opts.ShowCover,
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return out.String(), nil
}

type templateData struct {
	Title     string
	Subtitle  string
	Author    string
	Keywords  string
	CSS       template.CSS
	Body      template.HTML
	BuiltAt   string
	TotalDocs int
	ShowCover bool
}

func collectKeywords(docs []*mdx.Document) string {
	seen := make(map[string]bool)
	var tags []string
	for _, d := range docs {
		for _, t := range d.Tags() {
			if !seen[t] {
				seen[t] = true
				tags = append(tags, t)
			}
		}
	}
	sort.Strings(tags)
	return strings.Join(tags, ", ")
}
