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

// Builtin theme names, usable with Get/ResolveByName and as extends: values
// in custom theme YAML.
const (
	NameDefault   = "default"
	NameMinimal   = "minimal"
	NameModern    = "modern"
	NameClassic   = "classic"
	NameCorporate = "corporate"
	NameDark      = "dark"
	NameAcademic  = "academic"
	NameEditorial = "editorial"
)

const (
	categoryProfessional = "professional"
	categoryMinimal      = "minimal"
	categoryEditorial    = "editorial"
	categoryDark         = "dark"
	categoryAcademic     = "academic"
)

var allSectionsOn = ResolvedSections{Cover: true, TOC: true, PageNumbers: true, Header: true}

// order lists builtin theme names in the order List() should return them.
var order = []string{
	NameDefault, NameMinimal, NameModern, NameClassic,
	NameCorporate, NameDark, NameAcademic, NameEditorial,
}

var registry = map[string]Theme{
	NameDefault: {
		Name:        NameDefault,
		Description: "Clean, professional look that fits any technical document.",
		Category:    categoryProfessional,
		CSS:         defaultCSS,
		Sections:    allSectionsOn,
	},
	NameMinimal: {
		Name:        NameMinimal,
		Description: "Stripped down: smaller type, no borders, maximum simplicity.",
		Category:    categoryMinimal,
		CSS:         minimalCSS,
		Sections:    allSectionsOn,
	},
	NameModern: {
		Name:        NameModern,
		Description: "Sans-serif with generous whitespace and bold accent underlines.",
		Category:    categoryProfessional,
		CSS:         modernCSS,
		Sections:    allSectionsOn,
	},
	NameClassic: {
		Name:        NameClassic,
		Description: "Serif, traditional book layout — ink on paper.",
		Category:    categoryEditorial,
		CSS:         classicCSS,
		Sections:    allSectionsOn,
	},
	NameCorporate: {
		Name:        NameCorporate,
		Description: "Structured blue/gray palette for client-facing reports.",
		Category:    categoryProfessional,
		CSS:         corporateCSS,
		Sections:    allSectionsOn,
	},
	NameDark: {
		Name:        NameDark,
		Description: "Dark background with light text. Best for on-screen PDFs.",
		Category:    categoryDark,
		CSS:         darkCSS,
		Sections:    allSectionsOn,
	},
	NameAcademic: {
		Name:        NameAcademic,
		Description: "Formal serif layout for theses, papers, and reports.",
		Category:    categoryAcademic,
		CSS:         academicCSS,
		Sections:    allSectionsOn,
	},
	NameEditorial: {
		Name:        NameEditorial,
		Description: "Magazine-style display headings and pull-quote blockquotes.",
		Category:    categoryEditorial,
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
