// Package component provides a generic UI component system for YAML-configured
// Bubbletea bubbles. Components are created from YAML configuration using a
// registry of bubble types, with properties applied via reflection.
package component

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"

	tea "charm.land/bubbletea/v2"
)

// Component is a UI component that renders part of the application.
type Component interface {
	View() string
	Update(msg tea.Msg) (Component, tea.Cmd)
	Name() string
}

// Constructor creates a component from a map of properties.
type Constructor func(properties map[string]any) (Component, error)

// globalRegistry holds the default component builder.
var globalRegistry *Registry

// Init initializes the global registry with built-in component types.
func Init() {
	globalRegistry = NewRegistry()

	// Register bubble types with the factory (generic instantiation).
	// Constructors return pointers because PropertySetter needs pointers
	// to modify struct fields via reflection.
	globalRegistry.RegisterBubbleType("textarea", func() interface{} {
		m := textarea.New()
		return &m
	})
	globalRegistry.RegisterBubbleType("input", func() interface{} {
		m := textinput.New()
		return &m
	})
}

// Build creates a component from a type name and properties using the global registry.
func Build(componentType string, properties map[string]any) (Component, error) {
	if globalRegistry == nil {
		Init()
	}
	return globalRegistry.Build(componentType, properties)
}

// Register adds a new component type to the global registry.
func Register(name string, constructor Constructor) {
	if globalRegistry == nil {
		Init()
	}
	globalRegistry.RegisterHelper(name, constructor)
}

// RegisterBubbleType adds a bubble type to the global registry's factory.
func RegisterBubbleType(name string, constructor BubbleConstructor) {
	if globalRegistry == nil {
		Init()
	}
	globalRegistry.RegisterBubbleType(name, constructor)
}
