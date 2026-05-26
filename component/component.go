// Package component provides a generic UI component system for YAML-configured
// Bubbletea bubbles. Components are created from YAML configuration using a
// registry of bubble types, with properties applied via reflection.
package component

import (
	"sync"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"

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
var (
	globalRegistry *Registry
	initOnce       sync.Once
)

// Init initializes the global registry with built-in component types.
// Calling Init more than once is safe and has no effect after the first call.
func Init() {
	initOnce.Do(func() {
		globalRegistry = NewRegistry()

		globalRegistry.RegisterBubbleType("textarea", BubbleType{
			New: func() interface{} {
				m := textarea.New()
				return &m
			},
			DoUpdate: func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd) {
				m := model.(*textarea.Model)
				newM, cmd := m.Update(msg)
				return &newM, cmd
			},
		})

		globalRegistry.RegisterBubbleType("input", BubbleType{
			New: func() interface{} {
				m := textinput.New()
				return &m
			},
			DoUpdate: func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd) {
				m := model.(*textinput.Model)
				newM, cmd := m.Update(msg)
				return &newM, cmd
			},
		})

		globalRegistry.RegisterBubbleType("viewport", BubbleType{
			New: func() interface{} {
				m := viewport.New()
				return &m
			},
			DoUpdate: func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd) {
				m := model.(*viewport.Model)
				newM, cmd := m.Update(msg)
				return &newM, cmd
			},
		})
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
func RegisterBubbleType(name string, bt BubbleType) {
	if globalRegistry == nil {
		Init()
	}
	globalRegistry.RegisterBubbleType(name, bt)
}
