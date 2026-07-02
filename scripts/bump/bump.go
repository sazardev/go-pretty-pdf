package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: bump <patch|minor|major>")
		os.Exit(1)
	}

	level := os.Args[1]

	data, err := os.ReadFile("version/version.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read version file: %v\n", err)
		os.Exit(1)
	}

	re := regexp.MustCompile(`var Version = "([^"]+)"`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		fmt.Fprintln(os.Stderr, "version string not found in version/version.go")
		os.Exit(1)
	}

	current := matches[1]

	parts := strings.Split(strings.TrimPrefix(current, "v"), ".")
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "unexpected version format: %s\n", current)
		os.Exit(1)
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch level {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		fmt.Fprintf(os.Stderr, "unknown level: %s (use patch, minor, or major)\n", level)
		os.Exit(1)
	}

	newVer := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	newContent := re.ReplaceAllString(string(data), fmt.Sprintf(`var Version = "v%s"`, newVer))

	if err := os.WriteFile("version/version.go", []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write version file: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(newVer)
}
