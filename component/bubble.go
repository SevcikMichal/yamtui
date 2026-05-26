package component

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// BubbleType groups the construction and typed update logic for a bubble model.
// New creates a pointer to the initial model value. DoUpdate receives the current
// model pointer, applies the message, and returns the updated pointer and any command.
type BubbleType struct {
	New      func() interface{}
	DoUpdate func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd)
}

// bubbleComponent is a generic wrapper for any Bubble Tea model.
// It delegates View and Update calls to the underlying model.
type bubbleComponent struct {
	name     string
	model    interface{}
	doUpdate func(msg tea.Msg) tea.Cmd
}

// NewBubbleComponent creates a new generic bubble component wrapper.
func NewBubbleComponent(name string, model interface{}) Component {
	return &bubbleComponent{
		name:  name,
		model: model,
	}
}

func (b *bubbleComponent) Name() string {
	return b.name
}

// Focus delegates to the underlying model's Focus method.
// It implements the Focus() tea.Cmd interface for textinput/textarea.
func (b *bubbleComponent) Focus() tea.Cmd {
	if focuser, ok := b.model.(interface{ Focus() tea.Cmd }); ok {
		return focuser.Focus()
	}
	if focuser, ok := b.model.(interface{ Focus() }); ok {
		focuser.Focus()
	}
	return nil
}

// Blur delegates to the underlying model's Blur method.
func (b *bubbleComponent) Blur() {
	if blurrer, ok := b.model.(interface{ Blur() }); ok {
		blurrer.Blur()
	}
}

// SetWidth delegates to the underlying model's SetWidth method.
func (b *bubbleComponent) SetWidth(w int) {
	if setter, ok := b.model.(interface{ SetWidth(int) }); ok {
		setter.SetWidth(w)
	}
}

// SetSize delegates to the underlying model's SetSize method if available.
func (b *bubbleComponent) SetSize(width, height int) {
	if setter, ok := b.model.(interface{ SetSize(int, int) }); ok {
		setter.SetSize(width, height)
	} else {
		// Fallback: set width only.
		b.SetWidth(width)
	}
}

func (b *bubbleComponent) View() string {
	if viewer, ok := b.model.(interface{ View() string }); ok {
		return viewer.View()
	}
	return ""
}

func (b *bubbleComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	if b.doUpdate == nil {
		return b, nil
	}
	cmd := b.doUpdate(msg)
	return b, cmd
}

// SetContent sets the text content of the underlying model if it supports it.
// Used by CommandCallback.SetContent to push content into viewport components.
func (b *bubbleComponent) SetContent(content string) {
	if setter, ok := b.model.(interface{ SetContent(string) }); ok {
		setter.SetContent(content)
	}
}

// BubbleFactory maps bubble type names to BubbleType registrations.
type BubbleFactory struct {
	types map[string]BubbleType
}

// NewBubbleFactory creates a new empty factory.
func NewBubbleFactory() *BubbleFactory {
	return &BubbleFactory{
		types: make(map[string]BubbleType),
	}
}

// Register adds a BubbleType registration to the factory.
func (f *BubbleFactory) Register(name string, bt BubbleType) {
	f.types[name] = bt
}

// Create instantiates a bubble by name, wraps it in a bubbleComponent, and
// wires the typed update closure so Update never uses reflection.
func (f *BubbleFactory) Create(name string) (Component, error) {
	bt, ok := f.types[name]
	if !ok {
		return nil, fmt.Errorf("unknown bubble type %q", name)
	}
	model := bt.New()
	comp := &bubbleComponent{
		name:  name,
		model: model,
	}
	if bt.DoUpdate != nil {
		comp.doUpdate = func(msg tea.Msg) tea.Cmd {
			newModel, cmd := bt.DoUpdate(comp.model, msg)
			comp.model = newModel
			return cmd
		}
	}
	return comp, nil
}
