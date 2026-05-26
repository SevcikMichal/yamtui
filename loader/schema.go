// Package loader provides YAML configuration parsing and schema types
// for the declarative UI framework.
package loader

import (
	"fmt"
	"os"

	yaml "go.yaml.in/yaml/v4"
)

// Configuration is the root YAML configuration structure.
type Configuration struct {
	AltScreen  bool                       `yaml:"alt_screen"`
	Theme      ThemeConfig                `yaml:"theme"`
	Components map[string]ComponentConfig `yaml:"components"`
	Layout     LayoutConfig               `yaml:"layout"`
	Keybindings map[string]string         `yaml:"keybindings"`
	Commands   map[string]CommandConfig   `yaml:"commands"`
}

// Validate checks the config for required fields and consistency.
func (c *Configuration) Validate() error {
	// TODO: once the API stabilizes
	return nil
}

// ThemeConfig specifies which theme to use.
type ThemeConfig struct {
	Name string `yaml:"name"`
}

// ComponentConfig defines a UI component from YAML.
type ComponentConfig struct {
	Type       string           `yaml:"type"`
	Properties map[string]any   `yaml:"properties"`
}

// LayoutConfig defines component ordering and sizing.
type LayoutConfig struct {
	Rows   []RowConfig          `yaml:"rows"`
	Sizing map[string]SizeConfig `yaml:"sizing"`
}

// RowConfig defines a single row of components in the grid layout.
type RowConfig struct {
	Components []string  `yaml:"components"`
	Spacing    float64   `yaml:"spacing"` // gap between columns in chars (default: 1)
}

// SizeConfig defines how a component is sized in both dimensions.
type SizeConfig struct {
	Width  *DimensionConfig `yaml:"width"`
	Height *DimensionConfig `yaml:"height"`
}

// DimensionConfig defines sizing for a single dimension.
type DimensionConfig struct {
	Type  string  `yaml:"type"`  // "fixed", "ratio", "fill"
	Value float64 `yaml:"value"` // required for "fixed" and "ratio"
}

// CommandConfig defines a command from YAML.
type CommandConfig struct {
	Type   string `yaml:"type"`
	Target string `yaml:"target"` // component target (for focus)
	Bind   string `yaml:"bind"`   // optional: auto-creates keybinding (e.g., "enter", "ctrl+c")
}

// Load reads and parses a YAML configuration file.
func Load(path string) (*Configuration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Configuration
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config %q: %w", path, err)
	}

	return &cfg, nil
}
