package mdx

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

type ComponentHandler func(attrs map[string]string, innerHTML string) string

type ComponentRegistry struct {
	handlers map[string]ComponentHandler
}

func NewComponentRegistry() *ComponentRegistry {
	r := &ComponentRegistry{
		handlers: make(map[string]ComponentHandler),
	}
	r.Register("DeepDive", deepDiveHandler)
	r.Register("Warning", warningHandler)
	r.Register("Axiom", axiomHandler)
	return r
}

func (r *ComponentRegistry) Register(name string, handler ComponentHandler) {
	r.handlers[name] = handler
}

func (r *ComponentRegistry) Transpile(input string) string {
	result := input
	for name, handler := range r.handlers {
		re := buildComponentRegex(name)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			matches := re.FindStringSubmatch(match)
			title := matches[1]
			content := strings.TrimSpace(matches[2])
			attrs := map[string]string{}
			if title != "" {
				attrs["title"] = title
			}
			return handler(attrs, content)
		})
	}
	return result
}

func buildComponentRegex(name string) *regexp.Regexp {
	pattern := fmt.Sprintf(`(?s)<%s(?:\s+title="([^"]*)")?\s*>(.*?)</%s>`, name, name)
	return regexp.MustCompile(pattern)
}

func deepDiveHandler(attrs map[string]string, inner string) string {
	var buf strings.Builder
	buf.WriteString(`<aside class="component-deep-dive">`)
	if attrs["title"] != "" {
		fmt.Fprintf(&buf, `<div class="component-deep-dive-title">%s</div>`+"\n", html.EscapeString(attrs["title"]))
	}
	buf.WriteString(rewriteInlineContent(inner))
	buf.WriteString("</aside>")
	return buf.String()
}

func warningHandler(attrs map[string]string, inner string) string {
	var buf strings.Builder
	buf.WriteString(`<div class="component-warning">`)
	if attrs["title"] != "" {
		fmt.Fprintf(&buf, `<div class="component-warning-title">%s</div>`+"\n", html.EscapeString(attrs["title"]))
	}
	buf.WriteString(rewriteInlineContent(inner))
	buf.WriteString("</div>")
	return buf.String()
}

func axiomHandler(attrs map[string]string, inner string) string {
	var buf strings.Builder
	buf.WriteString(`<blockquote class="component-axiom">`)
	if attrs["title"] != "" {
		fmt.Fprintf(&buf, `<div class="component-axiom-title">%s</div>`+"\n", html.EscapeString(attrs["title"]))
	}
	buf.WriteString(rewriteInlineContent(inner))
	buf.WriteString("</blockquote>")
	return buf.String()
}

var codeSpanRe = regexp.MustCompile("`([^`]+)`")
var boldRe = regexp.MustCompile(`\*\*([^*]+)\*\*`)

func rewriteInlineContent(content string) string {
	result := codeSpanRe.ReplaceAllString(content, "<code>$1</code>")
	result = boldRe.ReplaceAllString(result, "<strong>$1</strong>")
	result = strings.ReplaceAll(result, "\n", "<br>\n")
	return result
}
