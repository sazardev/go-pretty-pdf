package theme

import (
	"fmt"
	"os"
	"strings"
)

// ResolveByName resolves a theme by name — a builtin ("default",
// "corporate", ...), the name of a custom theme discovered in the
// project-local or global themes directory, or a direct path to a
// ".theme.yml"/".css" file — and returns its final CSS plus the resolved
// section toggles.
func ResolveByName(name string, opts Options, cwd string) (string, ResolvedSections, error) {
	if name == "" {
		name = "default"
	}

	switch {
	case strings.HasSuffix(name, ".css"):
		return resolveRawCSSFile(name, opts)
	case strings.HasSuffix(name, ThemeFileSuffix), strings.HasSuffix(name, ".theme.yaml"):
		ct, err := LoadCustomTheme(name)
		if err != nil {
			return "", ResolvedSections{}, err
		}
		return ct.Resolve(opts)
	}

	if t, ok := Get(name); ok {
		return Resolve(t, opts)
	}

	if path, ok := FindCustom(cwd, name); ok {
		ct, err := LoadCustomTheme(path)
		if err != nil {
			return "", ResolvedSections{}, err
		}
		return ct.Resolve(opts)
	}

	return "", ResolvedSections{}, fmt.Errorf("unknown theme %q (not a builtin, custom theme, or file path)", name)
}

func resolveRawCSSFile(path string, opts Options) (string, ResolvedSections, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ResolvedSections{}, fmt.Errorf("reading theme CSS file %s: %w", path, err)
	}
	t := Theme{Name: path, CSS: string(data), Sections: allSectionsOn}
	return Resolve(t, opts)
}
