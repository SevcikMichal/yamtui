// Package component provides a generic UI component system for YAML-configured
// Bubbletea bubbles. Components are created from YAML configuration using a
// registry of bubble types, with properties applied via reflection.
package component

import (
	"sync"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/paginator"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/stopwatch"
	"charm.land/bubbles/v2/table"
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

// Init initializes the global registry with all built-in Bubble Tea component
// types. Calling Init more than once is safe and has no effect after the first call.
func Init() {
	initOnce.Do(func() {
		globalRegistry = NewRegistry()

		// reg is a shorthand: registers a bubble using reflection-based update
		// dispatch so no per-type DoUpdate boilerplate is needed.
		reg := func(name string, newFn func() interface{}) {
			globalRegistry.RegisterBubbleType(name, BubbleType{New: newFn})
		}

		// Text input
		reg("input", func() interface{} { return textinput.New() })
		reg("textarea", func() interface{} { return textarea.New() })

		// Scrollable read-only content
		reg("viewport", func() interface{} { return viewport.New() })

		// Activity indicators
		reg("spinner", func() interface{} { return spinner.New() })
		reg("progress", func() interface{} { return progress.New() })
		reg("stopwatch", func() interface{} { return stopwatch.New() })

		// Data display
		reg("table", func() interface{} { return table.New() })

		// Navigation / selection
		reg("paginator", func() interface{} { return paginator.New() })
		reg("filepicker", func() interface{} { return filepicker.New() })

		// Static display — not focusable, not interactive.
		globalRegistry.RegisterHelper("text", func(props map[string]any) (Component, error) {
			val := ""
			if v, ok := props["value"]; ok {
				if s, ok := v.(string); ok {
					val = s
				}
			}
			return &textComponent{value: val}, nil
		})
	})
}

// textComponent is a non-interactive label that renders a static string.
// It deliberately does not implement Focus() so it is excluded from focus traversal.
type textComponent struct {
	value string
}

func (t *textComponent) Name() string                          { return "text" }
func (t *textComponent) View() string                          { return t.value }
func (t *textComponent) Update(_ tea.Msg) (Component, tea.Cmd) { return t, nil }

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

// RegisterBubble registers any Bubble Tea model using only a constructor.
// Update dispatch is derived automatically via reflection so no DoUpdate
// boilerplate is required. The constructor may return a value or a pointer;
// the framework takes the address if necessary.
//
//	component.RegisterBubble("spinner", func() interface{} { return spinner.New() })
func RegisterBubble(name string, newFn func() interface{}) {
	if globalRegistry == nil {
		Init()
	}
	globalRegistry.RegisterBubbleType(name, BubbleType{New: newFn})
}
