package component

import (
	"fmt"
	"log"
	"sync"
)

// Registry manages bubble type registrations and component construction.
type Registry struct {
	mu         sync.RWMutex
	factory    *BubbleFactory
	helperCons map[string]Constructor
	setter     *PropertySetter
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		factory:    NewBubbleFactory(),
		helperCons: make(map[string]Constructor),
		setter:     &PropertySetter{},
	}
}

// RegisterBubbleType registers a bubble type with its constructor.
func (r *Registry) RegisterBubbleType(name string, bt BubbleType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factory.Register(name, bt)
}

// RegisterHelper registers a helper component type with its constructor.
func (r *Registry) RegisterHelper(name string, constructor Constructor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.helperCons[name] = constructor
}

// Register adds a new component type to the registry.
// It registers the constructor as a helper component.
func (r *Registry) Register(name string, constructor Constructor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.helperCons[name] = constructor
}

// Build creates a component from a name and a map of properties.
func (r *Registry) Build(name string, properties map[string]any) (Component, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Try bubble factory first (for typed bubbles like input, textarea)
	if comp, err := r.factory.Create(name); err == nil {
		// Apply generic properties.
		if err := r.applyProperties(comp, properties); err != nil {
			return nil, fmt.Errorf("applying properties to %q: %w", name, err)
		}
		return comp, nil
	}

	// Try helper component constructors.
	constructor, ok := r.helperCons[name]
	if !ok {
		return nil, fmt.Errorf("unknown component type %q", name)
	}

	comp, err := constructor(properties)
	if err != nil {
		return nil, fmt.Errorf("constructing component %q: %w", name, err)
	}

	// Apply generic properties.
	if err := r.applyProperties(comp, properties); err != nil {
		return nil, fmt.Errorf("applying properties to %q: %w", name, err)
	}

	return comp, nil
}

// applyProperties applies all properties from the config to the component.
func (r *Registry) applyProperties(comp Component, properties map[string]any) error {
	if properties == nil {
		return nil
	}

	// Get the underlying model from the component.
	model := getModelProperty(comp)
	if model == nil {
		return nil
	}

	for prop, value := range properties {
		if err := r.setter.SetProperty(model, prop, value); err != nil {
			log.Printf("yamtui: property %q not applicable: %v", prop, err)
		}
	}

	return nil
}

// getModelProperty extracts the underlying model from a Component.
// This is a bit hacky but necessary for reflection-based property setting.
func getModelProperty(comp Component) interface{} {
	switch c := comp.(type) {
	case *bubbleComponent:
		return c.model
	default:
		// Try to find a model field via reflection
		return nil
	}
}
