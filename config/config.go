package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultConfigFile = "go-pretty-pdf.yml"

type Config struct {
	Title    string            `yaml:"title"`
	Subtitle string            `yaml:"subtitle"`
	Author   string            `yaml:"author"`
	Source   string            `yaml:"source"`
	Output   string            `yaml:"output"`
	Theme    string            `yaml:"theme"`
	CSS      string            `yaml:"css"`
	Template string            `yaml:"template"`
	Vars     map[string]string `yaml:"vars"`
	Lint     LintConfig        `yaml:"lint"`
	Render   RenderConfig      `yaml:"render"`
}

type LintConfig struct {
	RequireFrontmatter []string `yaml:"require_frontmatter"`
	NoDuplicateIDs     bool     `yaml:"no_duplicate_ids"`
	MaxHeadingDepth    int      `yaml:"max_heading_depth"`
}

type RenderConfig struct {
	Timeout     string `yaml:"timeout"`
	Paper       string `yaml:"paper"`
	MarginTop   string `yaml:"margin_top"`
	MarginBot   string `yaml:"margin_bottom"`
	MarginLeft  string `yaml:"margin_left"`
	MarginRight string `yaml:"margin_right"`
	HeaderTitle string `yaml:"header_title"`
}

func Default() *Config {
	return &Config{
		Source: "book",
		Output: "out.pdf",
		Title:  "Document",
		Author: "go-pretty-pdf",
		Lint: LintConfig{
			RequireFrontmatter: []string{"id", "title"},
			NoDuplicateIDs:     true,
			MaxHeadingDepth:    3,
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return cfg, nil
}

func FindConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		path := filepath.Join(dir, DefaultConfigFile)
		if _, err := os.Stat(path); err == nil {
			return filepath.Abs(path)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", nil
}
