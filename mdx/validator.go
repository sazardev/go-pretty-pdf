package mdx

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationError struct {
	File    string
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s: %s: %s", e.File, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type Validator interface {
	Validate(doc *Document) []ValidationError
}

type DefaultValidator struct {
	RequireFrontmatter        []string
	RequireIDFormat           string
	NoDuplicateIDs            bool
	MaxHeadingDepth           int
	RequireLowercaseFilenames bool
	CheckBrokenLinks          bool
}

func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{
		RequireFrontmatter: []string{"id", "title"},
		NoDuplicateIDs:     true,
		MaxHeadingDepth:    3,
	}
}

var headingRe = regexp.MustCompile(`<h([1-6])[ >]`)
var linkRe = regexp.MustCompile(`href="#([^"]+)"`)

func (v *DefaultValidator) Validate(doc *Document) []ValidationError {
	var errs []ValidationError

	for _, field := range v.RequireFrontmatter {
		val, ok := doc.Frontmatter[field]
		if !ok || isEmptyValue(val) {
			errs = append(errs, ValidationError{
				File:    doc.Path,
				Field:   field,
				Message: "required frontmatter field is missing",
			})
		}
	}

	id := doc.ID()
	if id == "" {
		errs = append(errs, ValidationError{
			File:    doc.Path,
			Field:   "id",
			Message: "id field is required and must be in [X.Y.Z] format",
		})
	} else if !idExtractRe.MatchString(id) {
		errs = append(errs, ValidationError{
			File:    doc.Path,
			Field:   "id",
			Message: "id must match format [X.Y.Z] (e.g. [1.0.0])",
		})
	}

	if v.MaxHeadingDepth > 0 {
		depth := countMaxHeadingDepth(doc.HTML)
		if depth > v.MaxHeadingDepth {
			errs = append(errs, ValidationError{
				File:    doc.Path,
				Field:   "content",
				Message: fmt.Sprintf("heading depth exceeds maximum of h%d (found h%d)", v.MaxHeadingDepth, depth),
			})
		}
	}

	return errs
}

func (v *DefaultValidator) ValidateAll(docs []*Document) []ValidationError {
	var errs []ValidationError

	for _, doc := range docs {
		errs = append(errs, v.Validate(doc)...)
	}

	if v.NoDuplicateIDs {
		seen := make(map[string]string)
		for _, doc := range docs {
			id := doc.ID()
			if id == "" {
				continue
			}
			if prev, ok := seen[id]; ok {
				errs = append(errs, ValidationError{
					File:    doc.Path,
					Field:   "id",
					Message: fmt.Sprintf("duplicate ID %q (also in %s)", id, prev),
				})
			} else {
				seen[id] = doc.Path
			}
		}
	}

	return errs
}

func countMaxHeadingDepth(html string) int {
	maxDepth := 0
	matches := headingRe.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		d := atoi(m[1])
		if d > maxDepth {
			maxDepth = d
		}
	}
	return maxDepth
}

func isEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []interface{}:
		return len(val) == 0
	}
	return false
}
