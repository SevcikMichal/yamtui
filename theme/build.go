package theme

import (
	"fmt"

	"github.com/SevcikMichal/yamtui/internal/loader"
)

// BuildFromConfig creates a Theme from a loader.ThemeConfig.
// If base is specified, it inherits from the named built-in theme.
func BuildFromConfig(registry *ThemeRegistry, cfg loader.ThemeConfig) (*Theme, error) {
	var base *Theme
	if cfg.Base != "" {
		base = registry.Get(cfg.Base)
		if base == nil {
			// Try loading by name as a built-in
			base = registry.Get(cfg.Name)
			if base == nil {
				return nil, fmt.Errorf("base theme %q not found", cfg.Base)
			}
		}
	}

	name := cfg.Name
	if name == "" {
		if cfg.Base != "" {
			name = cfg.Base + "-custom"
		} else {
			name = "custom"
		}
	}

	th := &Theme{
		Name:            name,
		Colors:          NewColorPalette(),
		Default:         NewStyle(),
		Focused:         NewStyle(),
		Error:           NewStyle(),
		Styles:          make(map[string]Style),
		ComponentStyles: make(map[string]Style),
	}

	// If we have a base, deep copy it
	if base != nil {
		*th = *base.Copy()
		th.Name = name
	}

	// Apply colors
	for name, color := range cfg.Colors {
		if err := th.Colors.Set(name, color); err != nil {
			return nil, fmt.Errorf("setting color %q: %w", name, err)
		}
	}

	// Apply default style
	for prop, value := range cfg.Default {
		if err := th.Default.SetProperty(prop, value); err != nil {
			// Log warning but don't fail - allows for future properties
			// fmt.Printf("warning: theme %q: unknown property %q: %v\n", name, prop, err)
		}
	}

	// Apply focused style
	for prop, value := range cfg.Focused {
		if err := th.Focused.SetProperty(prop, value); err != nil {
			// fmt.Printf("warning: theme %q: unknown property %q: %v\n", name, prop, err)
		}
	}

	// Apply error style
	for prop, value := range cfg.Error {
		if err := th.Error.SetProperty(prop, value); err != nil {
			// fmt.Printf("warning: theme %q: unknown property %q: %v\n", name, prop, err)
		}
	}

	// Apply named styles
	for styleName, props := range cfg.Styles {
		s := NewStyle()
		for prop, value := range props {
			if err := s.SetProperty(prop, value); err != nil {
				// fmt.Printf("warning: theme %q: unknown property %q in style %q: %v\n", name, prop, styleName, err)
			}
		}
		th.Styles[styleName] = s
	}

	// Apply component-specific styles
	for compName, props := range cfg.Components {
		s := NewStyle()
		for prop, value := range props {
			if err := s.SetProperty(prop, value); err != nil {
				// fmt.Printf("warning: theme %q: unknown property %q for component %q: %v\n", name, prop, compName, err)
			}
		}
		th.ComponentStyles[compName] = s
	}

	return th, nil
}
