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

//go:embed assets/sepia.css
var sepiaCSS string

//go:embed assets/terminal.css
var terminalCSS string

//go:embed assets/blueprint.css
var blueprintCSS string

//go:embed assets/ivy.css
var ivyCSS string

//go:embed assets/government.css
var governmentCSS string

//go:embed assets/resume.css
var resumeCSS string

//go:embed assets/legal.css
var legalCSS string

//go:embed assets/latex.css
var latexCSS string

//go:embed assets/gruvbox.css
var gruvboxCSS string

// Builtin theme names, usable with Get/ResolveByName and as extends: values
// in custom theme YAML.
const (
	NameDefault    = "default"
	NameMinimal    = "minimal"
	NameModern     = "modern"
	NameClassic    = "classic"
	NameCorporate  = "corporate"
	NameDark       = "dark"
	NameAcademic   = "academic"
	NameEditorial  = "editorial"
	NameSepia      = "sepia"
	NameTerminal   = "terminal"
	NameBlueprint  = "blueprint"
	NameIvy        = "ivy"
	NameGovernment = "government"
	NameResume     = "resume"
	NameLegal      = "legal"
	NameLatex      = "latex"
	NameGruvbox    = "gruvbox"
)

const (
	categoryProfessional  = "professional"
	categoryMinimal       = "minimal"
	categoryEditorial     = "editorial"
	categoryDark          = "dark"
	categoryAcademic      = "academic"
	categoryWarm          = "warm"
	categoryTechnical     = "technical"
	categoryInstitutional = "institutional"
	categoryResume        = "resume"
	categoryFormal        = "formal"
)

var allSectionsOn = ResolvedSections{Cover: true, TOC: true, PageNumbers: true, Header: true}

// resumeSections turns off every section a CV/one-pager has no use for —
// there's no cover page, no table of contents, no running header, and
// (typically single-page) no page numbers.
var resumeSections = ResolvedSections{Cover: false, TOC: false, PageNumbers: false, Header: false}

// order lists builtin theme names in the order List() should return them.
var order = []string{
	NameDefault, NameMinimal, NameModern, NameClassic,
	NameCorporate, NameDark, NameAcademic, NameEditorial,
	NameSepia, NameTerminal, NameBlueprint,
	NameIvy, NameGovernment, NameResume, NameLegal, NameLatex,
	NameGruvbox,
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
		Accented:    true,
	},
	NameClassic: {
		Name:        NameClassic,
		Description: "Serif, traditional book layout — ink on paper.",
		Category:    categoryEditorial,
		CSS:         classicCSS,
		Sections:    allSectionsOn,
		Accented:    true,
	},
	NameCorporate: {
		Name:        NameCorporate,
		Description: "Structured blue/gray palette for client-facing reports.",
		Category:    categoryProfessional,
		CSS:         corporateCSS,
		Sections:    allSectionsOn,
		Accented:    true,
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
		Accented:    true,
	},
	NameSepia: {
		Name:        NameSepia,
		Description: "Warm, sepia-toned palette for long, comfortable reading sessions.",
		Category:    categoryWarm,
		CSS:         sepiaCSS,
		Sections:    allSectionsOn,
	},
	NameTerminal: {
		Name:        NameTerminal,
		Description: "All-monospace, terminal-inspired look for technical references.",
		Category:    categoryTechnical,
		CSS:         terminalCSS,
		Sections:    allSectionsOn,
		Accented:    true,
	},
	NameBlueprint: {
		Name:        NameBlueprint,
		Description: "Dark technical blueprint palette with monospace type and cyan highlights.",
		Category:    categoryTechnical,
		CSS:         blueprintCSS,
		Sections:    allSectionsOn,
		Accented:    true,
	},
	NameIvy: {
		Name:        NameIvy,
		Description: "Classic Ivy League university letterhead: forest green and gold on cream.",
		Category:    categoryInstitutional,
		CSS:         ivyCSS,
		Sections:    allSectionsOn,
		Accented:    true,
	},
	NameGovernment: {
		Name:        NameGovernment,
		Description: "Formal official-document palette: navy and bronze, centered headings.",
		Category:    categoryInstitutional,
		CSS:         governmentCSS,
		Sections:    allSectionsOn,
		Accented:    true,
	},
	NameResume: {
		Name:        NameResume,
		Description: "Clean, ATS-friendly sans-serif for CVs and one-pagers — no cover or TOC.",
		Category:    categoryResume,
		CSS:         resumeCSS,
		Sections:    resumeSections,
	},
	NameLegal: {
		Name:        NameLegal,
		Description: "Stark, formal brief style: black ink, no color as decoration.",
		Category:    categoryFormal,
		CSS:         legalCSS,
		Sections:    allSectionsOn,
	},
	NameLatex: {
		Name:        NameLatex,
		Description: "Mathematical/scientific paper look with automatic section numbering.",
		Category:    categoryAcademic,
		CSS:         latexCSS,
		Sections:    allSectionsOn,
	},
	NameGruvbox: {
		Name:        NameGruvbox,
		Description: "Retro warm dark palette inspired by the popular Gruvbox editor theme.",
		Category:    categoryTechnical,
		CSS:         gruvboxCSS,
		Sections:    allSectionsOn,
		Accented:    true,
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
