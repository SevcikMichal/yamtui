package component

import (
	"fmt"
	"reflect"

	tea "charm.land/bubbletea/v2"
)

// BubbleType groups the construction and typed update logic for a bubble model.
// New creates a pointer to the initial model value. DoUpdate receives the current
// model pointer, applies the message, and returns the updated pointer and any command.
type BubbleType struct {
	New      func() interface{}
	DoUpdate func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd)
}

// Updatable is the constraint satisfied by all Bubble Tea model types whose
// Update method returns the same type (the standard pattern for bubbles).
type Updatable[M any] interface {
	Update(tea.Msg) (M, tea.Cmd)
}

// MakeBubbleType creates a BubbleType for any bubble model that satisfies
// Updatable. This eliminates per-type DoUpdate boilerplate: any bubble whose
// Update returns its own type can be registered with a single line:
//
//	component.RegisterBubbleType("spinner", component.MakeBubbleType(spinner.New))
func MakeBubbleType[M Updatable[M]](newFn func() M) BubbleType {
	return BubbleType{
		New: func() interface{} {
			m := newFn()
			return &m
		},
		DoUpdate: func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd) {
			m := model.(*M)
			newM, cmd := (*m).Update(msg)
			return &newM, cmd
		},
	}
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

// Init calls the underlying model's Init() if it exists, returning the
// initial command (e.g. the first tick for spinner/stopwatch/progress).
// For components that use Tick() tea.Msg instead of Init() tea.Cmd
// (e.g. charm.land/bubbles/v2 spinner, stopwatch), Tick is used as
// the initial command so the animation loop starts immediately.
func (b *bubbleComponent) Init() tea.Cmd {
	if initer, ok := b.model.(interface{ Init() tea.Cmd }); ok {
		return initer.Init()
	}
	// Bubbles v2 animated components (spinner, stopwatch) expose Tick() tea.Msg.
	// A method value matching func() tea.Msg satisfies tea.Cmd.
	if ticker, ok := b.model.(interface{ Tick() tea.Msg }); ok {
		return ticker.Tick
	}
	return nil
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

// makeReflectiveDoUpdate builds an update closure for any bubble model using
// reflection. It is called once at registration/creation time: element type and
// method index are pre-cached so the per-call cost is two reflect.Value ops
// rather than a full MethodByName lookup every frame.
// model must be a pointer to the bubble struct.
func makeReflectiveDoUpdate(model interface{}) func(interface{}, tea.Msg) (interface{}, tea.Cmd) {
	modelType := reflect.TypeOf(model) // *M
	elemType := modelType.Elem()       // M

	updateIdx := -1
	for i := 0; i < modelType.NumMethod(); i++ {
		if modelType.Method(i).Name == "Update" {
			updateIdx = i
			break
		}
	}
	if updateIdx < 0 {
		return func(m interface{}, _ tea.Msg) (interface{}, tea.Cmd) { return m, nil }
	}

	return func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd) {
		results := reflect.ValueOf(model).Method(updateIdx).Call(
			[]reflect.Value{reflect.ValueOf(msg)},
		)
		newPtr := reflect.New(elemType)
		newPtr.Elem().Set(results[0])
		var cmd tea.Cmd
		if c, ok := results[1].Interface().(tea.Cmd); ok {
			cmd = c
		}
		return newPtr.Interface(), cmd
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
// wires the update closure. If BubbleType.DoUpdate is nil the closure is
// derived automatically via reflection (element type and method index are
// pre-cached here, not on every frame).
func (f *BubbleFactory) Create(name string) (Component, error) {
	bt, ok := f.types[name]
	if !ok {
		return nil, fmt.Errorf("unknown bubble type %q", name)
	}
	model := bt.New()
	// Ensure the stored model is always a pointer so pointer-receiver
	// methods (Focus, Blur, SetWidth, SetContent …) work correctly.
	if v := reflect.ValueOf(model); v.Kind() != reflect.Ptr {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		model = ptr.Interface()
	}
	comp := &bubbleComponent{name: name, model: model}
	doUpdateFn := bt.DoUpdate
	if doUpdateFn == nil {
		doUpdateFn = makeReflectiveDoUpdate(model)
	}
	comp.doUpdate = func(msg tea.Msg) tea.Cmd {
		newModel, cmd := doUpdateFn(comp.model, msg)
		comp.model = newModel
		return cmd
	}
	return comp, nil
}
