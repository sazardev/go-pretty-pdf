package mdx

import "fmt"

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
