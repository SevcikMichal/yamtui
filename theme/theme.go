package theme

import (
	"sync"
)

// Theme holds a collection of named styles.
type Theme struct {
	Name            string
	Colors          *ColorPalette
	Default         Style
	Focused         Style
	Error           Style
	Styles          map[string]Style
	ComponentStyles map[string]Style
}

// GetStyle returns the style for a component name.
// It merges Default with any component-specific overrides.
func (t *Theme) GetStyle(componentName string) Style {
	s := t.Default.Copy()
	if overrides, ok := t.ComponentStyles[componentName]; ok {
		s = s.Merge(overrides)
	}
	return s
}

// GetFocusedStyle returns the style for a focused component.
// It layers Default → component overrides → focused style, so all three contribute.
func (t *Theme) GetFocusedStyle(componentName string) Style {
	s := t.GetStyle(componentName)
	if t.Focused.IsDefined() {
		s = s.Merge(t.Focused)
	}
	return s
}

// GetNamedStyle returns a named style (e.g., "focused", "error").
// Returns the Default style if the named style doesn't exist.
func (t *Theme) GetNamedStyle(name string) Style {
	switch name {
	case "default":
		return t.Default
	case "focused":
		if t.Focused.IsDefined() {
			return t.Focused
		}
		return t.Default
	case "error":
		return t.Error
	default:
		if s, ok := t.Styles[name]; ok {
			return s
		}
		return t.Default
	}
}

// Copy returns a deep copy of the Theme.
func (t *Theme) Copy() *Theme {
	cp := &Theme{
		Name:            t.Name,
		Colors:          t.Colors.Copy(),
		Default:         t.Default.Copy(),
		Focused:         t.Focused.Copy(),
		Error:           t.Error.Copy(),
		Styles:          make(map[string]Style),
		ComponentStyles: make(map[string]Style),
	}
	for k, v := range t.Styles {
		cp.Styles[k] = v.Copy()
	}
	for k, v := range t.ComponentStyles {
		cp.ComponentStyles[k] = v.Copy()
	}
	return cp
}

// applyPalettes sets the theme's color palette on all styles.
// This should be called after the theme is fully constructed.
func (t *Theme) applyPalettes() {
	t.Default.SetPalette(t.Colors)
	t.Focused.SetPalette(t.Colors)
	t.Error.SetPalette(t.Colors)
	for k, s := range t.Styles {
		s.SetPalette(t.Colors)
		t.Styles[k] = s
	}
	for k, s := range t.ComponentStyles {
		s.SetPalette(t.Colors)
		t.ComponentStyles[k] = s
	}
}

// ThemeRegistry stores and looks up themes by name.
type ThemeRegistry struct {
	themes map[string]*Theme
	mu     sync.RWMutex
}

// NewThemeRegistry creates a new registry with built-in themes loaded.
func NewThemeRegistry() *ThemeRegistry {
	r := &ThemeRegistry{
		themes: make(map[string]*Theme),
	}
	r.registerBuiltins()
	return r
}

// Register adds a theme to the registry.
func (r *ThemeRegistry) Register(t *Theme) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.themes[t.Name] = t
}

// Get retrieves a theme by name. Returns nil if not found.
func (r *ThemeRegistry) Get(name string) *Theme {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.themes[name]
}

// registerBuiltins registers all built-in themes.
func (r *ThemeRegistry) registerBuiltins() {
	r.Register(DefaultTheme())
	r.Register(CatppuccinTheme())
	r.Register(DraculaTheme())
	r.Register(NordTheme())
}
