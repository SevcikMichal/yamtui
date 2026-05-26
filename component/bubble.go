package component

import (
	"fmt"
	"reflect"

	tea "charm.land/bubbletea/v2"
)

// bubbleComponent is a generic wrapper for any bubble tea model.
// It delegates View and Update calls to the underlying model.
type bubbleComponent struct {
	name  string
	model interface{}
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
	// Always call Update on the pointer - this gives us access to both
	// value and pointer receiver methods via reflection.
	ptr := reflect.ValueOf(b.model)
	if ptr.Kind() != reflect.Ptr {
		return b, nil
	}

	updateMethod := ptr.MethodByName("Update")
	if !updateMethod.IsValid() {
		return b, nil
	}

	args := []reflect.Value{reflect.ValueOf(msg)}
	results := updateMethod.Call(args)

	if len(results) >= 1 {
		resultVal := results[0]
		// Create a new pointer to the updated model value
		newPtr := reflect.New(resultVal.Type())
		newPtr.Elem().Set(resultVal)
		b.model = newPtr.Interface()
	}

	var cmd tea.Cmd
	if len(results) >= 2 {
		if cmdVal, ok := results[1].Interface().(tea.Cmd); ok {
			cmd = cmdVal
		}
	}

	return b, cmd
}

// BubbleConstructor is a function that creates a new bubble tea model.
type BubbleConstructor func() interface{}

// BubbleFactory maps bubble type names to constructor functions.
type BubbleFactory struct {
	constructors map[string]BubbleConstructor
}

// NewBubbleFactory creates a new empty factory.
func NewBubbleFactory() *BubbleFactory {
	return &BubbleFactory{
		constructors: make(map[string]BubbleConstructor),
	}
}

// Register adds a bubble type constructor to the factory.
func (f *BubbleFactory) Register(name string, constructor BubbleConstructor) {
	f.constructors[name] = constructor
}

// Create instantiates a bubble by name and wraps it in a bubbleComponent.
func (f *BubbleFactory) Create(name string) (Component, error) {
	constructor, ok := f.constructors[name]
	if !ok {
		return nil, fmt.Errorf("unknown bubble type %q", name)
	}
	model := constructor()
	return &bubbleComponent{
		name:  name,
		model: model,
	}, nil
}
