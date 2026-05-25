// Package loader provides YAML configuration parsing and schema types
// for the declarative UI framework.
package loader

// Config is the root YAML configuration structure.
type Config struct {
	Theme       ThemeConfig           `yaml:"theme"`
	Components  map[string]CompConfig  `yaml:"components"`
	Layout      LayoutConfig           `yaml:"layout"`
	Keybindings map[string]string      `yaml:"keybindings"`
	Commands    map[string]CmdConfig   `yaml:"commands"`
}

// ThemeConfig specifies which theme to use.
type ThemeConfig struct {
	Name string `yaml:"name"`
}

// CompConfig defines a UI component from YAML.
type CompConfig struct {
	Type        string      `yaml:"type"`
	Placeholder string      `yaml:"placeholder"`
	Text        string      `yaml:"text"` // for helpbar
	Config      CompOptions `yaml:"config"`
}

// CompOptions holds component-specific configuration options.
type CompOptions struct {
	CharLimit int  `yaml:"char_limit"`
	ReadOnly  bool `yaml:"readonly"`
	Focus     bool `yaml:"focus"`
}

// LayoutConfig defines component ordering and sizing.
type LayoutConfig struct {
	Order  []string        `yaml:"order"`
	Sizing map[string]SizeConfig `yaml:"sizing"`
}

// SizeConfig defines how a component is sized.
type SizeConfig struct {
	Height string  `yaml:"height"` // "fixed", "ratio", "fill"
	Value  float64 `yaml:"value"`  // fixed: line count; ratio: 0.0-1.0
}

// CmdConfig defines a command from YAML.
type CmdConfig struct {
	Type   string       `yaml:"type"`
	Target string       `yaml:"target"`    // component target (for clear, focus)
	Steps  []StepConfig `yaml:"steps"`     // for sequence commands
}

// StepConfig defines a step within a sequence command.
type StepConfig struct {
	Action string `yaml:"action"`
	Target string `yaml:"target"`
	Value  string `yaml:"value"`
	Into   string `yaml:"into"`
}
