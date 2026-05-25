package builder

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/layout"
	"github.com/SevcikMichal/yamtui/loader"
)

// App is the main application struct that implements tea.Model.
type App struct {
	Components map[string]Component
	Commands   map[string]Command
	KeyMap     map[string]string // key string -> command name
	Layout     loader.LayoutConfig
	Order      []string

	// Runtime state
	TermWidth  int
	TermHeight int
}

// Command represents an action that can be executed.
type Command interface {
	Execute(app *App) tea.Cmd
}

// Component is a UI component that renders part of the application.
type Component interface {
	View() string
	Update(msg tea.Msg) (Component, tea.Cmd)
	Name() string
}

// TextEditor is a component that supports text editing operations.
type TextEditor interface {
	Component
	Value() string
	SetValue(string)
}

// TextScroller is a component that supports scrolling to end.
type TextScroller interface {
	TextEditor
	MoveToEnd()
}

// commandConstructors maps command type strings to constructor functions.
var commandConstructors = make(map[string]func() Command)

// RegisterCommand registers a command constructor by type name.
func RegisterCommand(name string, constructor func() Command) {
	commandConstructors[name] = constructor
}

// BuildCommands creates commands from a YAML config map using registered constructors.
func BuildCommands(commands map[string]loader.CmdConfig) (map[string]Command, error) {
	result := make(map[string]Command, len(commands))
	for name, cfg := range commands {
		constructor, ok := commandConstructors[cfg.Type]
		if !ok {
			return nil, fmt.Errorf("unknown command type %q for %q", cfg.Type, name)
		}
		result[name] = constructor()
	}
	return result, nil
}

// Build creates an App from a YAML config.
func Build(cfg *loader.Config) (*App, error) {
	// Build components.
	components, err := BuildComponents(cfg.Components)
	if err != nil {
		return nil, err
	}

	// Build commands.
	commands, err := BuildCommands(cfg.Commands)
	if err != nil {
		return nil, err
	}

	// Keybindings map is already from YAML.
	keyMap := cfg.Keybindings

	return &App{
		Components: components,
		Commands:   commands,
		KeyMap:     keyMap,
		Layout:     cfg.Layout,
		Order:      cfg.Layout.Order,
	}, nil
}

// Init returns the initialization command for the program.
func (a *App) Init() tea.Cmd {
	return nil
}

// Update processes messages and delegates to handlers and components.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var allCmds []tea.Cmd

	// Handle window resize.
	if wmsg, ok := msg.(tea.WindowSizeMsg); ok {
		a.TermWidth = wmsg.Width
		a.TermHeight = wmsg.Height

		// Calculate and apply layout dimensions.
		sizes := layout.CalculateLayout(a.TermWidth, a.TermHeight, a.Layout)
		for name, size := range sizes {
			comp, ok := a.Components[name]
			if !ok {
				continue
			}
			if sc, ok := comp.(interface{ SetSize(int, int) }); ok {
				sc.SetSize(size.Width, size.Height)
			}
		}

		return a, nil
	}

	// Handle key presses.
	if kmsg, ok := msg.(tea.KeyPressMsg); ok {
		cmdName, exists := a.KeyMap[kmsg.String()]
		if exists {
			if cmd, ok := a.Commands[cmdName]; ok {
				teaCmd := cmd.Execute(a)
				if teaCmd != nil {
					allCmds = append(allCmds, teaCmd)
				}
			}
		}
	}

	// Delegate to components.
	for _, c := range a.Components {
		var cmd tea.Cmd
		c, cmd = c.Update(msg)
		if cmd != nil {
			allCmds = append(allCmds, cmd)
		}
	}

	return a, tea.Batch(allCmds...)
}

// View renders the UI using the component layout.
func (a *App) View() tea.View {
	var parts []string
	for _, name := range a.Order {
		if c, ok := a.Components[name]; ok {
			parts = append(parts, c.View())
		}
	}

	layoutStr := ""
	if len(parts) > 0 {
		layoutStr = parts[0]
		for _, p := range parts[1:] {
			layoutStr = layoutStr + "\n" + p
		}
	}

	return tea.View{
		Content:   layoutStr,
		AltScreen: true,
	}
}
