package loader

import (
	"fmt"
	"os"

	yaml "go.yaml.in/yaml/v4"
)

// Load reads and parses a YAML configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config %q: %w", path, err)
	}

	return &cfg, nil
}

// LoadDefaults returns a Config populated with built-in defaults.
func LoadDefaults() *Config {
	return &Config{
		Theme: ThemeConfig{Name: "default"},
		Components: map[string]CompConfig{
			"textarea": {
				Type:        "textarea",
				Placeholder: " Fullscreen log area. Type text below...",
				Config:      CompOptions{},
			},
			"input": {
				Type:        "input",
				Placeholder: "Type structural content and press Enter...",
				Config: CompOptions{
					CharLimit: 256,
					Focus:     true,
				},
			},
			"helpbar": {
				Type: "helpbar",
				Text: "(ctrl+c / esc to quit)",
			},
		},
		Layout: LayoutConfig{
			Order: []string{"textarea", "input", "helpbar"},
			Sizing: map[string]SizeConfig{
				"textarea": {Height: "ratio", Value: 0.7},
				"input":    {Height: "fixed", Value: 3},
				"helpbar":  {Height: "fixed", Value: 1},
			},
		},
		Keybindings: map[string]string{
			"ctrl+c": "quit",
			"esc":    "quit",
			"enter":  "submit",
		},
		Commands: map[string]CmdConfig{
			"quit":   {Type: "quit"},
			"submit": {Type: "submit"},
		},
	}
}

// Validate checks the config for required fields and consistency.
func (c *Config) Validate() error {
	if c.Components == nil {
		return fmt.Errorf("components must be defined")
	}

	for name, comp := range c.Components {
		if comp.Type == "" {
			return fmt.Errorf("component %q: type is required", name)
		}
		if comp.Type != "textarea" && comp.Type != "input" && comp.Type != "helpbar" {
			return fmt.Errorf("component %q: unknown type %q", name, comp.Type)
		}
	}

	for key, cmdName := range c.Keybindings {
		if _, ok := c.Commands[cmdName]; !ok {
			return fmt.Errorf("keybinding %q references unknown command %q", key, cmdName)
		}
	}

	return nil
}
