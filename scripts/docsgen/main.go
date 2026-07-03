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
	ID       string
	Title    string
	Content  string
	Icon     string
}

func main() {
	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error finding repo root: %v\n", err)
		os.Exit(1)
	}

	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmarkHtml.WithHardWraps(),
			goldmarkHtml.WithUnsafe(),
		),
	)

	sections := []Section{
		heroSection(),
	}

	readme, _ := os.ReadFile(filepath.Join(root, "README.md"))
	sections = append(sections, readmeSections(readme, mdRenderer)...)

	cli, _ := os.ReadFile(filepath.Join(root, "docs", "cli.md"))
	sections = append(sections, cliSections(cli, mdRenderer)...)

	changelog, _ := os.ReadFile(filepath.Join(root, "CHANGELOG.md"))
	sections = append(sections, changelogSection(changelog, mdRenderer)...)

	html := buildHTML(sections)

	outDir := filepath.Join(root, "_site")
	os.MkdirAll(outDir, 0755)
	os.WriteFile(filepath.Join(outDir, "index.html"), []byte(html), 0644)

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
	return Section{
		ID:    "hero",
		Title: "go-pretty-pdf",
		Content: `<p class="hero-tagline">Transform a directory of MDX files into a beautiful, print-ready PDF via headless Chrome.</p>
<div class="hero-badges">
<a href="https://pkg.go.dev/github.com/sazardev/go-pretty-pdf"><img src="https://pkg.go.dev/badge/github.com/sazardev/go-pretty-pdf.svg" alt="Go Reference"></a>
<a href="https://github.com/sazardev/go-pretty-pdf/actions/workflows/ci.yml"><img src="https://github.com/sazardev/go-pretty-pdf/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
<a href="https://goreportcard.com/report/github.com/sazardev/go-pretty-pdf"><img src="https://goreportcard.com/badge/github.com/sazardev/go-pretty-pdf" alt="Go Report Card"></a>
<a href="https://github.com/sazardev/go-pretty-pdf/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
</div>
<div class="hero-install">
<p>Install the CLI:</p>
<pre class="install-cmd"><code>go install github.com/sazardev/go-pretty-pdf/cmd/pretty-pdf@latest</code></pre>
<p>Or use as a library:</p>
<pre class="install-cmd"><code>go get github.com/sazardev/go-pretty-pdf</code></pre>
</div>`,
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
	var sections []Section

	sectionMap := map[string]struct {
		Title string
		Icon  string
	}{
		"install":                      {"Installation", "📥"},
		"quick-start":                  {"Quick Start", "🚀"},
		"how-it-works":                 {"How It Works", "⚙️"},
		"mdx-format":                   {"MDX Format", "📝"},
		"built-in-components":          {"Built-in Components", "🧩"},
		"configuration":                {"Configuration", "🔧"},
		"library-api":                  {"Library API", "💻"},
		"themes":                       {"Themes", "🎨"},
		"cli-reference":                {"CLI Reference", "🖥️"},
	}

	for _, part := range parts {
		id := anchorFromHeading(part.Heading)
		if info, ok := sectionMap[id]; ok {
			sections = append(sections, Section{
				ID:      id,
				Title:   info.Title,
				Icon:    info.Icon,
				Content: part.Body,
			})
		}
	}
	return sections
}

func cliSections(src []byte, md goldmark.Markdown) []Section {
	html := renderMarkdown(src, md)
	parts := splitByHeadings(html)
	var sections []Section

	sectionMap := map[string]struct {
		Title string
		Icon  string
	}{
		"overview":           {"CLI Overview", "📋"},
		"requirements":       {"Requirements", "✅"},
		"usage":              {"Usage", "⌨️"},
		"global-flags":       {"Global Flags", "🏷️"},
		"commands":           {"Commands", "📟"},
		"config-file":        {"Config File", "⚙️"},
		"themes":             {"Themes", "🎨"},
		"template-variables": {"Template Variables", "📊"},
		"environment":        {"Environment", "🌍"},
		"exit-codes":         {"Exit Codes", "🚪"},
	}

	for _, part := range parts {
		id := anchorFromHeading(part.Heading)
		if info, ok := sectionMap[id]; ok {
			sections = append(sections, Section{
				ID:      "cli-" + id,
				Title:   info.Title,
				Icon:    info.Icon,
				Content: part.Body,
			})
		}
	}

	// Individual command subsections (h3s inside the Commands section)
	for _, part := range parts {
		id := anchorFromHeading(part.Heading)
		if id == "commands" {
			cmdSubs := splitByH3(part.Body)
			cmdMap := map[string]struct {
				Title string
				Icon  string
			}{
				"build":      {"build — Generate PDF", "📦"},
				"check":      {"check — Validate MDX", "🔍"},
				"init":       {"init — Scaffold Project", "✨"},
				"serve":      {"serve — Live Preview", "🌐"},
				"watch":      {"watch — Auto-Rebuild", "👀"},
				"version":    {"version — Print Version", "📌"},
				"completion": {"completion — Shell Scripts", "🔤"},
			}
			for _, sub := range cmdSubs {
				subID := anchorFromHeading(sub.Heading)
				if info, ok := cmdMap[subID]; ok {
					sections = append(sections, Section{
						ID:      "cmd-" + subID,
						Title:   info.Title,
						Icon:    info.Icon,
						Content: sub.Body,
					})
				}
			}
		}
	}
	return sections
}

func changelogSection(src []byte, md goldmark.Markdown) []Section {
	html := renderMarkdown(src, md)
	return []Section{{
		ID:      "changelog",
		Title:   "Changelog",
		Icon:    "📜",
		Content: html,
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

		var contentEnd int
		if i+1 < len(openMatches) {
			contentEnd = openMatches[i+1][0]
		} else {
			contentEnd = len(html)
		}

		headingMatch := headingTextRe.FindStringSubmatch(html[om[0]:contentStart])
		heading := ""
		if len(headingMatch) >= 2 {
			heading = headingMatch[1]
		}

		body := strings.TrimSpace(html[contentStart:contentEnd])
		parts[i] = headingPart{Heading: heading, Level: 3, Body: body}
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

		var contentEnd int
		if i+1 < len(openMatches) {
			contentEnd = openMatches[i+1][0]
		} else {
			contentEnd = len(html)
		}

		headingMatch := headingTextRe.FindStringSubmatch(html[om[0]:contentStart])
		heading := ""
		if len(headingMatch) >= 2 {
			heading = headingMatch[1]
		}

		body := strings.TrimSpace(html[contentStart:contentEnd])
		parts[i] = headingPart{Heading: heading, Level: 2, Body: body}
	}
	return parts
}

func anchorFromHeading(html string) string {
	lower := strings.ToLower(html)
	lower = stripHTMLTags(lower)
	lower = strings.TrimSpace(lower)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(s, "")
}

func buildHTML(sections []Section) string {
	n := len(sections)
	navItems := make([]string, n)
	for i, s := range sections {
		navItems[i] = fmt.Sprintf(
			`<a href="#%s">%s %s</a>`, s.ID, s.Icon, s.Title)
	}

	bodyParts := make([]string, n)
	for i, s := range sections {
		wrapperClass := "section"
		if s.ID == "hero" {
			wrapperClass = "section hero-section"
		}
		bodyParts[i] = fmt.Sprintf(
			`<section id="%s" class="%s"><h2>%s %s</h2><div class="section-content">%s</div></section>`,
			s.ID, wrapperClass, s.Icon, s.Title, s.Content)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>go-pretty-pdf — Documentation</title>
<meta name="description" content="Transform MDX files into beautiful, print-ready PDFs via headless Chrome. Library + CLI for Go.">
<style>
%s
</style>
</head>
<body>
<nav class="sidebar">
  <div class="sidebar-header">
    <a href="#hero" class="sidebar-logo">go-pretty-pdf</a>
  </div>
  <div class="sidebar-links">
    %s
  </div>
</nav>
<main class="content">
  %s
  <footer class="page-footer">
    <p>Generated from source. <a href="https://github.com/sazardev/go-pretty-pdf">GitHub</a> &middot; <a href="https://pkg.go.dev/github.com/sazardev/go-pretty-pdf">Go Package</a></p>
  </footer>
</main>
</body>
</html>`, css(), strings.Join(navItems, "\n    "), strings.Join(bodyParts, "\n  "))
}

func css() string {
	return strings.TrimSpace(`
:root {
  --bg: #0d1117;
  --bg-secondary: #161b22;
  --bg-tertiary: #21262d;
  --border: #30363d;
  --text: #c9d1d9;
  --text-muted: #8b949e;
  --text-heading: #f0f6fc;
  --accent: #58a6ff;
  --accent-hover: #79c0ff;
  --green: #3fb950;
  --orange: #d2991d;
  --red: #f85149;
  --code-bg: #1c2128;
  --code-text: #c9d1d9;
  --sidebar-width: 260px;
  --font-mono: 'SF Mono', 'Fira Code', 'Fira Mono', Menlo, Consolas, monospace;
  --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

html { scroll-behavior: smooth; scroll-padding-top: 2rem; }

body {
  font-family: var(--font-sans);
  background: var(--bg);
  color: var(--text);
  line-height: 1.7;
  display: flex;
  min-height: 100vh;
}

a { color: var(--accent); text-decoration: none; }
a:hover { color: var(--accent-hover); text-decoration: underline; }

.sidebar {
  position: fixed;
  top: 0; left: 0;
  width: var(--sidebar-width);
  height: 100vh;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border);
  overflow-y: auto;
  z-index: 10;
  padding: 1.5rem 0;
}
.sidebar-header {
  padding: 0.5rem 1.25rem 1rem;
  border-bottom: 1px solid var(--border);
  margin-bottom: 0.5rem;
}
.sidebar-logo {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-heading) !important;
  text-decoration: none !important;
}
.sidebar-links { padding: 0.5rem 0; }
.sidebar-links a {
  display: block;
  padding: 0.4rem 1.25rem;
  font-size: 0.85rem;
  color: var(--text-muted);
  border-left: 2px solid transparent;
  transition: all 0.15s;
}
.sidebar-links a:hover {
  color: var(--text);
  background: var(--bg-tertiary);
  border-left-color: var(--accent);
  text-decoration: none;
}

.content {
  margin-left: var(--sidebar-width);
  flex: 1;
  max-width: 860px;
  padding: 3rem 3rem 3rem 4rem;
}

.hero-section {
  padding: 4rem 0;
  border-bottom: 1px solid var(--border);
  margin-bottom: 2rem;
}
.hero-section h2 {
  font-size: 2.5rem;
  font-weight: 800;
  color: var(--text-heading);
  letter-spacing: -0.02em;
}
.hero-tagline {
  font-size: 1.15rem;
  color: var(--text-muted);
  margin: 1rem 0 1.5rem;
  line-height: 1.6;
}
.hero-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 2rem;
}
.hero-badges img { height: 20px; }
.hero-install p {
  color: var(--text-muted);
  font-size: 0.9rem;
  margin: 0.75rem 0 0.375rem;
}
.install-cmd {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0.75rem 1rem;
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text);
  overflow-x: auto;
}
.install-cmd code {
  background: none;
  padding: 0;
  border: none;
  font-size: inherit;
}

.section {
  margin-bottom: 3rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid var(--border);
}
.section:last-of-type { border-bottom: none; }
.section h2 {
  font-size: 1.5rem;
  color: var(--text-heading);
  margin-bottom: 1.25rem;
  font-weight: 700;
}
.section h3 {
  font-size: 1.15rem;
  color: var(--text-heading);
  margin: 1.75rem 0 0.75rem;
  font-weight: 600;
}
.section h4 {
  font-size: 1rem;
  color: var(--text-heading);
  margin: 1.25rem 0 0.5rem;
  font-weight: 600;
}
.section p {
  margin: 0.75rem 0;
  color: var(--text);
}

.section-content table {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0 1.5rem;
  font-size: 0.9rem;
}
.section-content table th,
.section-content table td {
  border: 1px solid var(--border);
  padding: 0.5rem 0.85rem;
  text-align: left;
}
.section-content table th {
  background: var(--bg-tertiary);
  color: var(--text-heading);
  font-weight: 600;
}
.section-content table tr:nth-child(even) { background: var(--bg-secondary); }

.section-content pre {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 1rem;
  overflow-x: auto;
  margin: 1rem 0;
}
.section-content pre code {
  font-family: var(--font-mono);
  font-size: 0.825rem;
  line-height: 1.55;
  color: var(--code-text);
  background: none;
  padding: 0;
}
.section-content code {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  background: var(--code-bg);
  padding: 0.15em 0.4em;
  border-radius: 3px;
  border: 1px solid var(--border);
}
.section-content pre code {
  border: none;
  padding: 0;
  background: none;
}
.section-content ul, .section-content ol {
  padding-left: 1.5rem;
  margin: 0.75rem 0;
}
.section-content li { margin: 0.3rem 0; }
.section-content blockquote {
  border-left: 3px solid var(--accent);
  padding: 0.5rem 1rem;
  margin: 1rem 0;
  color: var(--text-muted);
  background: var(--bg-secondary);
  border-radius: 0 4px 4px 0;
}
.section-content hr {
  border: none;
  border-top: 1px solid var(--border);
  margin: 2rem 0;
}
.section-content img { max-width: 100%; }
.section-content input[type="checkbox"] {
  margin-right: 0.4rem;
  accent-color: var(--accent);
}

.section-content a[href*="github.com"]::after {
  content: " ↗";
  font-size: 0.7rem;
  opacity: 0.5;
}
.section-content h2 a[href^="#"],
.section-content h3 a[href^="#"],
.sidebar-links a::after,
.sidebar-logo::after,
.page-footer a::after {
  content: none !important;
}

.page-footer {
  margin-top: 3rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border);
  text-align: center;
  color: var(--text-muted);
  font-size: 0.85rem;
}

@media (max-width: 900px) {
  body { flex-direction: column; }
  .sidebar {
    position: relative;
    width: 100%;
    height: auto;
    border-right: none;
    border-bottom: 1px solid var(--border);
    padding: 0.75rem 1rem;
  }
  .sidebar-header {
    padding: 0;
    border-bottom: none;
    margin-bottom: 0;
  }
  .sidebar-links {
    display: none;
  }
  .content {
    margin-left: 0;
    padding: 1.5rem;
  }
  .hero-section h2 { font-size: 1.75rem; }
}
`)
}


