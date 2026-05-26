// Package command provides a declarative command system for TUI applications
// built on Bubbletea. Commands interact with the app through interfaces
// (AppContext, CommandCallback) without importing the app package.
package command

import (
	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/component"
)

// AppContext provides access to app state for commands.
// Commands use this to read state without importing the app package.
type AppContext interface {
	// GetComponent returns a component by name.
	GetComponent(name string) (component.Component, bool)
	// ComponentNames returns all registered component names.
	ComponentNames() []string
	// ComponentOrder returns components in layout order (for focus navigation).
	ComponentOrder() []string
	// FocusedComponent returns the currently focused component name.
	FocusedComponent() string
}

// CommandCallback is provided by the app and invoked by commands to signal actions.
// Commands use this to communicate desired actions without knowing about Bubbletea.
type CommandCallback interface {
	// Quit signals the application should quit.
	Quit()
	// Focus signals focus should move to a named component.
	Focus(name string)
	// SetContent pushes text content into a named component (e.g. viewport).
	SetContent(componentName string, content string)
	// Custom allows commands to register arbitrary actions.
	Custom(actionName string, data map[string]string)
}

// Command represents an action that can be executed.
// Commands interact with the app through AppContext (state) and CommandCallback (actions).
// They never import Bubbletea or return tea.Cmd directly.
type Command interface {
	Execute(ctx AppContext, cb CommandCallback)
}

// CustomActionHandlers maps custom action names to handler functions.
var CustomActionHandlers = make(map[string]func(ctx AppContext, cb CommandCallback) []tea.Cmd)

// RegisterCustomAction registers a handler for a custom action name.
func RegisterCustomAction(name string, handler func(ctx AppContext, cb CommandCallback) []tea.Cmd) {
	CustomActionHandlers[name] = handler
}

// CommandConstructors maps command type strings to constructor functions.
var CommandConstructors = make(map[string]func() Command)

// RegisterCommand registers a command constructor by type name.
func RegisterCommand(name string, constructor func() Command) {
	CommandConstructors[name] = constructor
}

// Built-in commands (available directly from YAML without importing anything).

// QuitCommand quits the application.
type QuitCommand struct{}

func (c QuitCommand) Execute(ctx AppContext, cb CommandCallback) {
	cb.Quit()
}

// FocusCommand sets focus to a specified input component.
type FocusCommand struct {
	Target string
}

func (c FocusCommand) Execute(ctx AppContext, cb CommandCallback) {
	if c.Target == "" {
		return
	}
	cb.Focus(c.Target)
}
