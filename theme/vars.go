package theme

import (
	"regexp"
	"strings"
)

var pdfVarDeclRe = regexp.MustCompile(`--pdf-([a-z0-9-]+):\s*([^;]+);`)

// ExtractCSSVars parses every `--pdf-<name>: <value>;` custom property
// *declaration* out of css (a single theme's CSS, or a fully composed
// document's <style> block) into a name->value map. It only matches
// declarations, never var(--pdf-x, ...) usages, since those have no colon
// right after the property name.
//
// When a property is declared more than once (e.g. Resolve() appends a
// :root{} override block after the base theme CSS for user-customized
// colors), the last occurrence wins — matching normal CSS cascade order
// for rules of equal specificity.
func ExtractCSSVars(css string) map[string]string {
	vars := make(map[string]string)
	for _, m := range pdfVarDeclRe.FindAllStringSubmatch(css, -1) {
		vars[m[1]] = strings.TrimSpace(m[2])
	}
	return vars
}
