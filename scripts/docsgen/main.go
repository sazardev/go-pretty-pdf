package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkHtml "github.com/yuin/goldmark/renderer/html"
)

type Section struct {
	ID      string
	Title   string
	Content string
}

func main() {
	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error finding repo root: %v\n", err)
		os.Exit(1)
	}

	mdRenderer := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(goldmarkHtml.WithHardWraps(), goldmarkHtml.WithUnsafe()),
	)

	readme, _ := os.ReadFile(filepath.Join(root, "README.md"))
	cli, _ := os.ReadFile(filepath.Join(root, "docs", "cli.md"))
	changelog, _ := os.ReadFile(filepath.Join(root, "CHANGELOG.md"))

	sections := make([]Section, 1, 28)
	sections[0] = heroSection()
	sections = append(sections, readmeSections(readme, mdRenderer)...)
	sections = append(sections, cliSections(cli, mdRenderer)...)
	sections = append(sections, changelogSection(changelog, mdRenderer)...)

	html := buildHTML(sections)
	outDir := filepath.Join(root, "_site")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating output directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(outDir, "index.html"), []byte(html), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing index.html: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Documentation site generated at _site/index.html")
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func heroSection() Section {
	ascii := `                             .__
  ____   ______             ______  |  |   _____  ____  __ __
 / ___\ /  ___/  ______   /  ___/  |  |   \__  \ \__  \|  |  \
/ /_/  >\___ \  /_____/   \___ \   |  |__  / __ \_/ __ \   Y  \
\___  /____  >           /____  >  |____/ (____  (____  /___|  /
/_____/    \/                 \/               \/     \/     \/
`
	return Section{
		ID:    "hero",
		Title: "go-pretty-pdf",
		Content: `<pre class="hero-ascii">` + ascii + `</pre>
<div class="hero-line"></div>
<p class="hero-tagline">Transform a directory of MDX files into a beautiful, print-ready PDF via headless Chrome.</p>
<div class="hero-meta">
  <span>Library + CLI</span>
  <span>Go 1.26+</span>
  <span>MIT</span>
</div>
<div class="hero-install">
  <div class="install-block">
    <span class="install-label">CLI</span>
    <pre class="install-cmd"><code>$ go install github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf@latest</code></pre>
  </div>
  <div class="install-block">
    <span class="install-label">Library</span>
    <pre class="install-cmd"><code>$ go get github.com/sazardev/go-pretty-pdf</code></pre>
  </div>
</div>
<p class="hero-requirements">
  Requires Chrome or Chromium for PDF rendering.
  <a href="#getting-started">Get started</a> &middot;
  <a href="https://github.com/sazardev/go-pretty-pdf">GitHub</a> &middot;
  <a href="https://pkg.go.dev/github.com/sazardev/go-pretty-pdf">pkg.go.dev</a>
</p>`,
	}
}

func renderMarkdown(src []byte, md goldmark.Markdown) string {
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return fmt.Sprintf("<p>Error rendering markdown: %v</p>", err)
	}
	return buf.String()
}

func readmeSections(src []byte, md goldmark.Markdown) []Section {
	html := renderMarkdown(src, md)
	parts := splitByHeadings(html)

	sectionMap := map[string]string{
		"install":             "Installation",
		"quick-start":         "Quick Start",
		"how-it-works":        "How It Works",
		"mdx-format":          "MDX Format",
		"built-in-components": "Built-in Components",
		"configuration":       "Configuration",
		"library-api":         "Library API",
		"themes":              "Themes",
		"cli-reference":       "CLI Reference",
	}

	sections := make([]Section, 0, len(sectionMap))
	for _, part := range parts {
		id := anchorFromHeading(part.Heading)
		if title, ok := sectionMap[id]; ok {
			sections = append(sections, Section{ID: id, Title: title, Content: part.Body})
		}
	}
	return sections
}

func cliSections(src []byte, md goldmark.Markdown) []Section {
	html := renderMarkdown(src, md)
	parts := splitByHeadings(html)

	sectionMap := map[string]string{
		"overview":           "CLI Overview",
		"requirements":       "Requirements",
		"usage":              "Usage",
		"global-flags":       "Global Flags",
		"commands":           "Commands",
		"config-file":        "Config File",
		"themes":             "Themes",
		"template-variables": "Template Variables",
		"environment":        "Environment",
		"exit-codes":         "Exit Codes",
	}

	sections := make([]Section, 0, 17)
	for _, part := range parts {
		id := anchorFromHeading(part.Heading)
		if title, ok := sectionMap[id]; ok {
			sections = append(sections, Section{ID: "cli-" + id, Title: title, Content: part.Body})
		}
		if id == "commands" {
			cmdSubs := splitByH3(part.Body)
			cmdMap := map[string]string{
				"build":      "build",
				"check":      "check",
				"init":       "init",
				"serve":      "serve",
				"watch":      "watch",
				"version":    "version",
				"completion": "completion",
			}
			for _, sub := range cmdSubs {
				subID := anchorFromHeading(sub.Heading)
				if cmdLabel, ok := cmdMap[subID]; ok {
					sections = append(sections, Section{
						ID:      "cmd-" + subID,
						Title:   "pretty-pdf " + cmdLabel,
						Content: sub.Body,
					})
				}
			}
		}
	}
	return sections
}

func changelogSection(src []byte, md goldmark.Markdown) []Section {
	return []Section{{
		ID:      "changelog",
		Title:   "Changelog",
		Content: renderMarkdown(src, md),
	}}
}

type headingPart struct {
	Heading string
	Level   int
	Body    string
}

func splitByH3(html string) []headingPart {
	h3Re := regexp.MustCompile(`<h3[^>]*>`)
	h3CloseRe := regexp.MustCompile(`</h3>`)
	headingTextRe := regexp.MustCompile(`<h3[^>]*>(.*?)</h3>`)

	openMatches := h3Re.FindAllStringSubmatchIndex(html, -1)
	if len(openMatches) == 0 {
		return []headingPart{{Heading: "", Level: 1, Body: html}}
	}

	parts := make([]headingPart, len(openMatches))
	for i, om := range openMatches {
		bodyStart := h3CloseRe.FindStringIndex(html[om[1]:])
		if bodyStart == nil {
			continue
		}
		contentStart := om[1] + bodyStart[1]
		contentEnd := len(html)
		if i+1 < len(openMatches) {
			contentEnd = openMatches[i+1][0]
		}
		headingMatch := headingTextRe.FindStringSubmatch(html[om[0]:contentStart])
		heading := ""
		if len(headingMatch) >= 2 {
			heading = headingMatch[1]
		}
		parts[i] = headingPart{Heading: heading, Level: 3, Body: strings.TrimSpace(html[contentStart:contentEnd])}
	}
	return parts
}

func splitByHeadings(html string) []headingPart {
	h2Re := regexp.MustCompile(`<h2[^>]*>`)
	h2CloseRe := regexp.MustCompile(`</h2>`)
	headingTextRe := regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`)

	openMatches := h2Re.FindAllStringSubmatchIndex(html, -1)
	if len(openMatches) == 0 {
		return []headingPart{{Heading: "", Level: 1, Body: html}}
	}

	parts := make([]headingPart, len(openMatches))
	for i, om := range openMatches {
		bodyStart := h2CloseRe.FindStringIndex(html[om[1]:])
		if bodyStart == nil {
			continue
		}
		contentStart := om[1] + bodyStart[1]
		contentEnd := len(html)
		if i+1 < len(openMatches) {
			contentEnd = openMatches[i+1][0]
		}
		headingMatch := headingTextRe.FindStringSubmatch(html[om[0]:contentStart])
		heading := ""
		if len(headingMatch) >= 2 {
			heading = headingMatch[1]
		}
		parts[i] = headingPart{Heading: heading, Level: 2, Body: strings.TrimSpace(html[contentStart:contentEnd])}
	}
	return parts
}

func anchorFromHeading(html string) string {
	lower := strings.ToLower(stripHTMLTags(html))
	slug := regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.TrimSpace(lower), "-")
	return strings.Trim(slug, "-")
}

func stripHTMLTags(s string) string {
	return regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, "")
}

func buildHTML(sections []Section) string {
	n := len(sections)
	navItems := make([]string, n)
	bodyParts := make([]string, n)

	for i, s := range sections {
		navItems[i] = fmt.Sprintf(`<a href="#%s">%s</a>`, s.ID, s.Title)
		cls := "section"
		if s.ID == "hero" {
			cls = "section hero-section"
		}
		bodyParts[i] = fmt.Sprintf(
			`<section id="%s" class="%s"><h2>%s</h2><div class="section-content">%s</div></section>`,
			s.ID, cls, s.Title, s.Content)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>go-pretty-pdf &mdash; Documentation</title>
<meta name="description" content="Transform MDX files into beautiful, print-ready PDFs via headless Chrome.">
<style>
%s
</style>
</head>
<body>
<nav class="sidebar">
  <div class="sidebar-brand">
    <a href="#hero">go-pretty-pdf</a>
  </div>
  <div class="sidebar-nav">
    %s
  </div>
</nav>
<main class="main">
  %s
  <footer class="footer">
    <p>Generated from source &mdash; <a href="https://github.com/sazardev/go-pretty-pdf">GitHub</a> &middot; <a href="https://pkg.go.dev/github.com/sazardev/go-pretty-pdf">pkg.go.dev</a></p>
  </footer>
</main>
</body>
</html>`, css(), strings.Join(navItems, "\n    "), strings.Join(bodyParts, "\n  "))
}

func css() string {
	return `
:root {
  --bg:            #fafafa;
  --bg-card:       #ffffff;
  --bg-hover:      #f0f0f0;
  --border:        #e5e5e5;
  --border-light:  #f0f0f0;
  --text:          #1a1a1a;
  --text-secondary:#555555;
  --text-muted:    #888888;
  --accent:        #4f46e5;
  --accent-light:  #818cf8;
  --accent-bg:     #eef2ff;
  --code-bg:       #1e1e2e;
  --code-text:     #cdd6f4;
  --inline-bg:     #f0f0f5;
  --inline-text:   #4f46e5;
  --green:         #059669;
  --sidebar-w:     250px;
  --radius:        6px;
  --font:          -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  --font-mono:     'SF Mono', 'Fira Code', 'Cascadia Code', Menlo, Consolas, monospace;
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

html { scroll-behavior: smooth; scroll-padding-top: 3rem; }

body {
  font-family: var(--font);
  background: var(--bg);
  color: var(--text);
  line-height: 1.75;
  display: flex;
  min-height: 100vh;
  -webkit-font-smoothing: antialiased;
}

a { color: var(--accent); text-decoration: none; transition: color .15s; }
a:hover { color: var(--accent-light); }

/*** Sidebar ***/

.sidebar {
  position: fixed; top: 0; left: 0;
  width: var(--sidebar-w); height: 100vh;
  background: var(--bg-card);
  border-right: 1px solid var(--border);
  overflow-y: auto;
  padding: 2rem 0 1.5rem;
  z-index: 100;
}
.sidebar-brand {
  padding: 0 1.5rem 1.25rem;
  border-bottom: 1px solid var(--border-light);
  margin-bottom: .75rem;
}
.sidebar-brand a {
  font-size: 1rem;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text) !important;
  letter-spacing: -.01em;
}
.sidebar-nav { padding: .25rem 0; }
.sidebar-nav a {
  display: block;
  padding: .35rem 1.5rem;
  font-size: .8rem;
  color: var(--text-secondary);
  border-left: 2px solid transparent;
  transition: all .15s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sidebar-nav a:hover {
  color: var(--text);
  background: var(--bg-hover);
  border-left-color: var(--accent);
}
.sidebar::-webkit-scrollbar { width: 4px; }
.sidebar::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }

/*** Main ***/

.main {
  margin-left: var(--sidebar-w);
  flex: 1;
  max-width: 800px;
  padding: 3rem 4rem 3rem;
}

/*** Hero ***/

.hero-section {
  padding: 2rem 0 3rem;
  margin-bottom: 3rem;
  border-bottom: 1px solid var(--border);
}
.hero-ascii {
  font-family: var(--font-mono);
  font-size: .55rem;
  line-height: 1.3;
  color: var(--accent);
  white-space: pre;
  margin-bottom: 1.5rem;
  overflow-x: auto;
  opacity: .85;
}
.hero-section h2 {
  font-size: 2.25rem;
  font-weight: 800;
  letter-spacing: -.025em;
  margin-bottom: .25rem;
  color: var(--text);
}
.hero-line {
  width: 48px; height: 3px;
  background: var(--accent);
  border-radius: 3px;
  margin: 1rem 0 1.5rem;
  animation: heroLineGrow .8s ease-out;
}
@keyframes heroLineGrow {
  from { width: 0; opacity: 0; }
  to   { width: 48px; opacity: 1; }
}
.hero-tagline {
  font-size: 1.05rem;
  color: var(--text-secondary);
  line-height: 1.7;
  margin-bottom: 1.5rem;
  max-width: 560px;
}
.hero-meta {
  display: flex; gap: 1.5rem;
  margin-bottom: 2rem;
}
.hero-meta span {
  font-size: .8rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: .06em;
  padding: .2rem .75rem;
  background: var(--accent-bg);
  border-radius: 99px;
}
.hero-install {
  margin-bottom: 1.5rem;
}
.install-block {
  margin-bottom: .75rem;
}
.install-label {
  font-size: .7rem;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: .08em;
  margin-bottom: .3rem;
  display: inline-block;
}
.install-cmd {
  background: var(--code-bg);
  border-radius: var(--radius);
  padding: .75rem 1rem;
  font-family: var(--font-mono);
  font-size: .78rem;
  color: var(--code-text);
  overflow-x: auto;
}
.install-cmd code { background: none; padding: 0; border: none; font-size: inherit; color: inherit; }
.hero-requirements {
  font-size: .82rem;
  color: var(--text-muted);
  line-height: 1.6;
}

/*** Sections ***/

.section {
  margin-bottom: 3.5rem;
  padding-bottom: 2.5rem;
  border-bottom: 1px solid var(--border-light);
}
.section:last-of-type { border-bottom: none; }
.section h2 {
  font-size: 1.35rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 1.25rem;
  letter-spacing: -.015em;
}
.section h3 {
  font-size: 1.05rem;
  color: var(--text);
  margin: 1.75rem 0 .75rem;
  font-weight: 600;
}
.section h4 {
  font-size: .92rem;
  color: var(--text);
  margin: 1.25rem 0 .5rem;
  font-weight: 600;
}
.section p { margin: .75rem 0; color: var(--text); }

/*** Content elements ***/

.section-content table {
  width: 100%; border-collapse: collapse;
  margin: 1rem 0 1.5rem;
  font-size: .85rem;
  border-radius: var(--radius);
  overflow: hidden;
}
.section-content table th,
.section-content table td {
  padding: .55rem .85rem;
  text-align: left;
  border-bottom: 1px solid var(--border-light);
}
.section-content table th {
  background: var(--accent-bg);
  color: var(--accent);
  font-weight: 600;
  font-size: .75rem;
  text-transform: uppercase;
  letter-spacing: .04em;
}
.section-content table td { background: var(--bg-card); }
.section-content table tr:last-child td { border-bottom: none; }

.section-content pre {
  background: var(--code-bg);
  border-radius: var(--radius);
  padding: 1rem 1.15rem;
  overflow-x: auto;
  margin: 1rem 0;
}
.section-content pre code {
  font-family: var(--font-mono);
  font-size: .78rem;
  line-height: 1.6;
  color: var(--code-text);
  background: none; padding: 0; border: none;
}
.section-content code {
  font-family: var(--font-mono);
  font-size: .82rem;
  background: var(--inline-bg);
  color: var(--inline-text);
  padding: .12em .45em;
  border-radius: 3px;
}
.section-content pre code {
  background: none;
  color: var(--code-text);
  padding: 0;
  border-radius: 0;
}
.section-content ul, .section-content ol { padding-left: 1.4rem; margin: .75rem 0; }
.section-content li { margin: .3rem 0; }
.section-content li::marker { color: var(--text-muted); }
.section-content blockquote {
  border-left: 3px solid var(--accent-light);
  padding: .5rem 1rem;
  margin: 1rem 0;
  color: var(--text-secondary);
  background: var(--accent-bg);
  border-radius: 0 var(--radius) var(--radius) 0;
  font-style: italic;
}
.section-content hr { border: none; border-top: 1px solid var(--border-light); margin: 2rem 0; }
.section-content img { max-width: 100%; }
.section-content input[type="checkbox"] { margin-right: .35rem; accent-color: var(--accent); }
.section-content a[href*="github.com"]::after {
  content: " \2197";
  font-size: .65rem; opacity: .4; font-style: normal;
}

/*** Footer ***/

.footer {
  margin-top: 3rem; padding-top: 1.5rem;
  border-top: 1px solid var(--border-light);
  text-align: center;
  color: var(--text-muted);
  font-size: .8rem;
}
.footer a { font-weight: 500; }

/*** Animations ***/

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(18px); }
  to   { opacity: 1; transform: translateY(0); }
}
.section:not(.hero-section) {
  animation: fadeSlideUp .5s ease-out both;
}
.section:nth-child(2)  { animation-delay: .05s; }
.section:nth-child(3)  { animation-delay: .1s;  }
.section:nth-child(4)  { animation-delay: .15s; }
.section:nth-child(5)  { animation-delay: .2s;  }
.section:nth-child(6)  { animation-delay: .25s; }
.section:nth-child(7)  { animation-delay: .3s;  }
.section:nth-child(8)  { animation-delay: .35s; }
.section:nth-child(9)  { animation-delay: .4s;  }
.section:nth-child(10) { animation-delay: .45s; }
.section:nth-child(11) { animation-delay: .5s;  }
.section:nth-child(12) { animation-delay: .55s; }
.section:nth-child(13) { animation-delay: .6s;  }
.section:nth-child(14) { animation-delay: .65s; }
.section:nth-child(15) { animation-delay: .7s;  }

/*** Responsive ***/

@media (max-width: 900px) {
  body { flex-direction: column; }
  .sidebar {
    position: relative; width: 100%; height: auto;
    border-right: none; border-bottom: 1px solid var(--border);
    padding: .75rem 1rem;
  }
  .sidebar-brand { padding: 0; border-bottom: none; margin-bottom: 0; }
  .sidebar-nav { display: none; }
  .main { margin-left: 0; padding: 1.5rem; }
  .hero-section h2 { font-size: 1.75rem; }
}
`
}
