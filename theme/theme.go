package theme

type Theme struct {
	Name     string
	CSS      string
	Template string
}

var Default = Theme{
	Name: "default",
}

var Minimal = Theme{
	Name: "minimal",
	CSS:  minimalCSS,
}

var minimalCSS = `@page {
  size: A4;
  margin: 18mm 14mm 18mm 14mm;
}

body {
  font-family: system-ui, sans-serif;
  font-size: 10pt;
  line-height: 1.5;
  color: #222;
}

h1 { font-size: 16pt; page-break-before: always; }
h2 { font-size: 13pt; }
h3 { font-size: 11pt; }
h4 { font-size: 10.5pt; }
h5 { font-size: 10pt; }

pre {
  background: #fafafa;
  padding: 8px 10px;
  font-size: 8.5pt;
  page-break-inside: avoid;
}

code {
  font-family: monospace;
  font-size: 9pt;
  background: #f4f4f4;
  padding: 1px 3px;
}

table {
  border-collapse: collapse;
  width: 100%;
  font-size: 9pt;
}

th, td {
  border: 1px solid #ddd;
  padding: 4px 8px;
}

th { background: #f0f0f0; }

.cover { min-height: 70vh; }
.cover h1 { font-size: 24pt; page-break-before: avoid; }
.cover-subtitle { font-size: 11pt; color: #666; }

.toc h1 { font-size: 18pt; }
.toc-h2 { padding-left: 16pt; }
.toc-h3 { padding-left: 32pt; }
.toc-h4 { padding-left: 48pt; }
.toc-h5 { padding-left: 64pt; }

.component-deep-dive {
  background: #f5f8ff;
  border-left: 3px solid #4a6cf7;
  padding: 8px 12px;
}

.component-warning {
  background: #fffbed;
  border-left: 3px solid #f7a84a;
  padding: 8px 12px;
}

.component-axiom {
  background: #f5faf5;
  border-left: 3px solid #4caf50;
  padding: 8px 12px;
  font-style: italic;
}
`
