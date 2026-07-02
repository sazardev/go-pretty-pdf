package mdx

import (
	"regexp"
	"strconv"
	"strings"
)

var idExtractRe = regexp.MustCompile(`^\[(\d+)\.(\d+)\.(\d+)\]$`)

type Document struct {
	Path        string
	Frontmatter map[string]interface{}
	HTML        string
}

func (d *Document) ID() string {
	v, _ := d.Frontmatter["id"].(string)
	return v
}

func (d *Document) Title() string {
	v, _ := d.Frontmatter["title"].(string)
	return v
}

func (d *Document) Subtitle() string {
	v, _ := d.Frontmatter["subtitle"].(string)
	return v
}

func (d *Document) Tags() []string {
	raw, ok := d.Frontmatter["tags"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	tags := make([]string, 0, len(arr))
	for _, t := range arr {
		if s, ok := t.(string); ok {
			tags = append(tags, s)
		}
	}
	return tags
}

func (d *Document) Difficulty() string {
	v, _ := d.Frontmatter["difficulty"].(string)
	return v
}

func (d *Document) Status() string {
	v, _ := d.Frontmatter["status"].(string)
	return v
}

func (d *Document) Completeness() int {
	raw, ok := d.Frontmatter["completeness"]
	if !ok {
		return 0
	}
	switch n := raw.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}

func (d *Document) DependsOn() []string {
	raw, ok := d.Frontmatter["depends_on"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	deps := make([]string, 0, len(arr))
	for _, dep := range arr {
		if s, ok := dep.(string); ok {
			deps = append(deps, s)
		}
	}
	return deps
}

func (d *Document) SortKey() [3]int {
	m := idExtractRe.FindStringSubmatch(d.ID())
	if m == nil {
		return [3]int{0, 0, 0}
	}
	return [3]int{atoi(m[1]), atoi(m[2]), atoi(m[3])}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func SplitID(id string) []int {
	var parts []int
	current := 0
	digits := false
	for _, c := range id {
		if c >= '0' && c <= '9' {
			current = current*10 + int(c-'0')
			digits = true
		} else {
			if digits {
				parts = append(parts, current)
				current = 0
				digits = false
			}
		}
	}
	if digits {
		parts = append(parts, current)
	}
	return parts
}

func AnchorID(id string) string {
	return "section-" + strings.ReplaceAll(id, "[", "")
}

func AnchorIDRaw(id string) string {
	return "section-" + id
}
