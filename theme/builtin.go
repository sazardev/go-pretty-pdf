package theme

import _ "embed"

//go:embed assets/base.css
var baseCSS string

//go:embed assets/default.css
var defaultCSS string

//go:embed assets/minimal.css
var minimalCSS string

//go:embed assets/modern.css
var modernCSS string

//go:embed assets/classic.css
var classicCSS string

//go:embed assets/corporate.css
var corporateCSS string

//go:embed assets/dark.css
var darkCSS string

//go:embed assets/academic.css
var academicCSS string

//go:embed assets/editorial.css
var editorialCSS string

var allSectionsOn = ResolvedSections{Cover: true, TOC: true, PageNumbers: true, Header: true}

// order lists builtin theme names in the order List() should return them.
var order = []string{
	"default", "minimal", "modern", "classic",
	"corporate", "dark", "academic", "editorial",
}

var registry = map[string]Theme{
	"default": {
		Name:        "default",
		Description: "Clean, professional look that fits any technical document.",
		Category:    "professional",
		CSS:         defaultCSS,
		Sections:    allSectionsOn,
	},
	"minimal": {
		Name:        "minimal",
		Description: "Stripped down: smaller type, no borders, maximum simplicity.",
		Category:    "minimal",
		CSS:         minimalCSS,
		Sections:    allSectionsOn,
	},
	"modern": {
		Name:        "modern",
		Description: "Sans-serif with generous whitespace and bold accent underlines.",
		Category:    "professional",
		CSS:         modernCSS,
		Sections:    allSectionsOn,
	},
	"classic": {
		Name:        "classic",
		Description: "Serif, traditional book layout — ink on paper.",
		Category:    "editorial",
		CSS:         classicCSS,
		Sections:    allSectionsOn,
	},
	"corporate": {
		Name:        "corporate",
		Description: "Structured blue/gray palette for client-facing reports.",
		Category:    "professional",
		CSS:         corporateCSS,
		Sections:    allSectionsOn,
	},
	"dark": {
		Name:        "dark",
		Description: "Dark background with light text. Best for on-screen PDFs.",
		Category:    "dark",
		CSS:         darkCSS,
		Sections:    allSectionsOn,
	},
	"academic": {
		Name:        "academic",
		Description: "Formal serif layout for theses, papers, and reports.",
		Category:    "academic",
		CSS:         academicCSS,
		Sections:    allSectionsOn,
	},
	"editorial": {
		Name:        "editorial",
		Description: "Magazine-style display headings and pull-quote blockquotes.",
		Category:    "editorial",
		CSS:         editorialCSS,
		Sections:    allSectionsOn,
	},
}

// Get looks up a builtin theme by name.
func Get(name string) (Theme, bool) {
	t, ok := registry[name]
	return t, ok
}

// List returns every builtin theme in a stable, curated order.
func List() []Theme {
	out := make([]Theme, 0, len(order))
	for _, name := range order {
		out = append(out, registry[name])
	}
	return out
}
