// Package component provides a generic UI component system for YAML-configured
// Bubbletea bubbles. Components are created from YAML configuration using a
// registry of bubble types, with properties applied via reflection.
package component

import (
	"fmt"
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
	"charm.land/lipgloss/v2"

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
		reg("viewport", func() interface{} {
			m := viewport.New()
			m.SoftWrap = true
			return m
		})

		// Activity indicators
		reg("spinner", func() interface{} { return spinner.New() })
		reg("progress", func() interface{} { return progress.New() })
		reg("stopwatch", func() interface{} { return stopwatch.New() })

		// Data display — registered as a helper so YAML rows/columns can be
		// converted from []interface{} to the typed table.Row / table.Column
		// slices that SetRows/SetColumns require.
		globalRegistry.RegisterHelper("table", func(props map[string]any) (Component, error) {
			m := table.New()

			// Parse rows from YAML. YAML delivers []interface{} of []interface{}.
			// The first row is treated as column headers; the rest are data rows.
			var numCols int
			if raw, ok := props["rows"]; ok {
				if rawRows, ok := raw.([]interface{}); ok {
					var allRows []table.Row
					for _, r := range rawRows {
						if cells, ok := r.([]interface{}); ok {
							row := make(table.Row, len(cells))
							for i, c := range cells {
								row[i] = fmt.Sprintf("%v", c)
							}
							allRows = append(allRows, row)
						}
					}
					if len(allRows) > 0 {
						// First row → column definitions.
						headerRow := allRows[0]
						numCols = len(headerRow)
						cols := make([]table.Column, numCols)
						for i, h := range headerRow {
							cols[i] = table.Column{Title: h, Width: 10}
						}
						m.SetColumns(cols)
						if len(allRows) > 1 {
							m.SetRows(allRows[1:])
						}
					}
				}
			}

			// Clear the default hot-pink selected-row style so it doesn't bleed.
			neutral := table.DefaultStyles()
			neutral.Selected = lipgloss.NewStyle()
			neutral.Header = neutral.Header.Bold(false)
			m.SetStyles(neutral)

			// Wrap in a tableWrapper so SetWidth redistributes column widths evenly.
			return &tableWrapper{model: &m, numCols: numCols}, nil
		})

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
	value    string
	maxWidth int
}

func (t *textComponent) Name() string { return "text" }
func (t *textComponent) View() string {
	if t.maxWidth > 0 && lipgloss.Width(t.value) > t.maxWidth {
		// Truncate to maxWidth-1 and add ellipsis.
		runes := []rune(t.value)
		truncated := string(runes[:t.maxWidth-1]) + "…"
		return truncated
	}
	return t.value
}
func (t *textComponent) Update(_ tea.Msg) (Component, tea.Cmd) { return t, nil }
func (t *textComponent) SetWidth(w int)                        { t.maxWidth = w }

// tableWrapper wraps a bubbles table.Model and redistributes column widths
// evenly when SetWidth is called, so the table fills its allocated cell.
type tableWrapper struct {
	model   *table.Model
	numCols int
}

func (t *tableWrapper) Name() string { return "table" }
func (t *tableWrapper) View() string { return t.model.View() }
func (t *tableWrapper) Update(msg tea.Msg) (Component, tea.Cmd) {
	newM, cmd := t.model.Update(msg)
	t.model = &newM
	return t, cmd
}
func (t *tableWrapper) Focus() tea.Cmd {
	t.model.Focus()
	return nil
}
func (t *tableWrapper) Blur() { t.model.Blur() }
func (t *tableWrapper) SetWidth(w int) {
	t.model.SetWidth(w)
	if t.numCols > 0 {
		// Redistribute width evenly across columns.
		colW := w / t.numCols
		if colW < 1 {
			colW = 1
		}
		cols := t.model.Columns()
		for i := range cols {
			cols[i].Width = colW
		}
		t.model.SetColumns(cols)
	}
}
func (t *tableWrapper) SetHeight(h int) { t.model.SetHeight(h) }
func (t *tableWrapper) SetSize(w, h int) {
	t.SetWidth(w)
	t.SetHeight(h)
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
