package config

import (
	"regexp"
	"strconv"
	"strings"
)

var cssUnitRe = regexp.MustCompile(`^(-?[0-9]*\.?[0-9]+)\s*([a-zA-Z]+)$`)

// PaperLetter is the config value for US Letter paper size.
const PaperLetter = "letter"

// ParsePaperSize maps a named paper size to its width/height in inches.
// ok is false for unrecognized names, leaving width/height unchanged.
func ParsePaperSize(name string) (width, height float64, ok bool) {
	switch strings.ToLower(name) {
	case PaperLetter:
		return 8.5, 11, true
	case "legal":
		return 8.5, 14, true
	case "a4":
		return 8.27, 11.69, true
	default:
		return 0, 0, false
	}
}

// ParseCSSUnit converts a CSS length string (e.g. "20mm", "0.8in") to
// inches. It returns 0 for empty or unrecognized input.
func ParseCSSUnit(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	m := cssUnitRe.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	value, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}

	switch strings.ToLower(m[2]) {
	case "in":
		return value
	case "mm":
		return value / 25.4
	case "cm":
		return value / 2.54
	case "pt":
		return value / 72.0
	case "px":
		return value / 96.0
	default:
		return 0
	}
}
