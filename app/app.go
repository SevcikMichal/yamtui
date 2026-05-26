// Package app provides a complete declarative UI application framework
// built on Bubbletea. It encapsulates the application model, component
// system, command registry, and entry point — external consumers do not
// need to import Bubbletea directly.
package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/command"
	"github.com/SevcikMichal/yamtui/component"
	"github.com/SevcikMichal/yamtui/layout"
	"github.com/SevcikMichal/yamtui/loader"
)

// App is the main application struct that implements tea.Model.
type App struct {
	AltScreen  bool
	Components map[string]component.Component
	Commands   map[string]command.Command
	KeyMap     map[string]string // key string -> command name
	Layout     loader.LayoutConfig

	// Runtime state
	TermWidth        int
	TermHeight       int
	focusedComponent string // currently focused component name
}

// appContext wraps *App to provide the command.AppContext interface.
type appContext struct {
	app *App
}

func (c *appContext) GetComponent(name string) (component.Component, bool) {
	comp, ok := c.app.Components[name]
	return comp, ok
}

func (c *appContext) ComponentNames() []string {
	names := make([]string, 0, len(c.app.Components))
	for name := range c.app.Components {
		names = append(names, name)
	}
	return names
}

// BuildApp creates an App from YAML configuration.
func BuildApp(cfg *loader.Configuration) (*App, error) {
	// Build components using the component registry.
	components, err := buildComponents(cfg.Components)
	if err != nil {
		return nil, err
	}

	// Build commands.
	commands, err := buildCommands(cfg.Commands)
	if err != nil {
		return nil, err
	}

	// Build keybindings map, merging explicit keybindings with inline bind fields.
	keyMap := make(map[string]string)
	for k, v := range cfg.Keybindings {
		keyMap[k] = v
	}
	for name, cmdCfg := range cfg.Commands {
		if cmdCfg.Bind != "" {
			keyMap[cmdCfg.Bind] = name
		}
	}

	return &App{
		AltScreen:  cfg.AltScreen,
		Components: components,
		Commands:   commands,
		KeyMap:     keyMap,
		Layout:     cfg.Layout,
	}, nil
}

// buildComponents creates components from YAML configuration.
func buildComponents(components map[string]loader.ComponentConfig) (map[string]component.Component, error) {
	result := make(map[string]component.Component, len(components))
	for name, cfg := range components {
		comp, err := component.Build(cfg.Type, cfg.Properties)
		if err != nil {
			return nil, fmt.Errorf("building component %q (type %q): %w", name, cfg.Type, err)
		}
		result[name] = comp
	}
	return result, nil
}

// buildCommands creates commands from YAML configuration.
func buildCommands(commands map[string]loader.CommandConfig) (map[string]command.Command, error) {
	result := make(map[string]command.Command, len(commands))
	for name, cfg := range commands {
		switch cfg.Type {
		case "quit":
			result[name] = command.QuitCommand{}
		case "focus":
			result[name] = command.FocusCommand{Target: cfg.Target}
		default:
			constructor, ok := command.CommandConstructors[cfg.Type]
			if !ok {
				continue
			}
			result[name] = constructor()
		}
	}
	return result, nil
}

// Init returns the initialization command for the program.
// It focuses the first component by default so it can receive keyboard input.
func (a *App) Init() tea.Cmd {
	for name, comp := range a.Components {
		if focuser, ok := comp.(interface{ Focus() tea.Cmd }); ok {
			a.focusedComponent = name
			return focuser.Focus()
		}
		// Fallback: check for Focus() without Cmd return
		if focuser, ok := comp.(interface{ Focus() }); ok {
			a.focusedComponent = name
			focuser.Focus()
		}
		break
	}
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
			// Set width via SetWidth interface (for textinput/textarea).
			// This enables responsive layouts defined in YAML using sizing.ratio or sizing.fill.
			if sw, ok := comp.(interface{ SetWidth(int) }); ok {
				sw.SetWidth(size.Width)
			}
		}

		return a, nil
	}

	// Handle key presses. Track whether the key was consumed by a command.
	keyConsumed := false
	if kmsg, ok := msg.(tea.KeyPressMsg); ok {
		cmdName, exists := a.KeyMap[strings.ToLower(kmsg.String())]
		if exists {
			if cmd, ok := a.Commands[cmdName]; ok {
				ctx := &appContext{app: a}
				cb := &commandCallback{app: a, cmds: &allCmds}
				cmd.Execute(ctx, cb)
				keyConsumed = true
			} else if handler, ok := command.CustomActionHandlers[cmdName]; ok {
				ctx := &appContext{app: a}
				cb := &commandCallback{app: a, cmds: &allCmds}
				cmds := handler(ctx, cb)
				allCmds = append(allCmds, cmds...)
				keyConsumed = true
			}
		} else {
			// Also try the raw string (in case keymap uses different casing)
			cmdNameRaw, existsRaw := a.KeyMap[kmsg.String()]
			if existsRaw {
				if cmd, ok := a.Commands[cmdNameRaw]; ok {
					ctx := &appContext{app: a}
					cb := &commandCallback{app: a, cmds: &allCmds}
					cmd.Execute(ctx, cb)
					keyConsumed = true
				} else if handler, ok := command.CustomActionHandlers[cmdNameRaw]; ok {
					ctx := &appContext{app: a}
					cb := &commandCallback{app: a, cmds: &allCmds}
					cmds := handler(ctx, cb)
					allCmds = append(allCmds, cmds...)
					keyConsumed = true
				}
			}
		}
	}

	// Only delegate to components if the key was not consumed by a command.
	// This prevents consumed keys (like tab/shift+tab for focus navigation)
	// from being processed as text input by unfocused components.
	if !keyConsumed {
		for _, c := range a.Components {
			var cmd tea.Cmd
			newComp, cmd := c.Update(msg)
			if newComp != nil {
				c = newComp
			}
			if cmd != nil {
				allCmds = append(allCmds, cmd)
			}
		}
	}

	return a, tea.Batch(allCmds...)
}

// View renders the UI using the component grid layout.
func (a *App) View() tea.View {
	var rows []string

	for _, row := range a.Layout.Rows {
		var cols []string
		for _, name := range row.Components {
			if c, ok := a.Components[name]; ok {
				cols = append(cols, c.View())
			}
		}
		rows = append(rows, joinCols(cols, row.Spacing))
	}

	layoutStr := joinRows(rows)

	return tea.View{
		Content:   layoutStr,
		AltScreen: a.AltScreen,
	}
}

// joinCols joins column views with spacing between them.
func joinCols(cols []string, spacing float64) string {
	if len(cols) == 0 {
		return ""
	}
	if spacing <= 0 {
		spacing = 1
	}
	sep := string(' ')
	result := cols[0]
	for _, col := range cols[1:] {
		result += sep + col
	}
	return result
}

// joinRows joins row strings with newlines between them.
func joinRows(rows []string) string {
	if len(rows) == 0 {
		return ""
	}
	result := rows[0]
	for _, row := range rows[1:] {
		result += "\n" + row
	}
	return result
}

// commandCallback implements CommandCallback and bridges to Bubbletea.
type commandCallback struct {
	app  *App
	cmds *[]tea.Cmd
}

func (cb *commandCallback) Quit() {
	*cb.cmds = append(*cb.cmds, tea.Quit)
}

func (cb *commandCallback) Focus(name string) {
	comp, ok := cb.app.Components[name]
	if !ok {
		return
	}
	// Blur the previously focused component first.
	// This ensures only one component shows its cursor at a time.
	prevName := cb.app.focusedComponent
	if prevName != "" && prevName != name {
		if prevComp, ok := cb.app.Components[prevName]; ok {
			if blurrer, ok := prevComp.(interface{ Blur() }); ok {
				blurrer.Blur()
			}
		}
	}
	cb.app.focusedComponent = name
	// Check if Focus() returns a tea.Cmd - Bubbletea needs this to actually set focus
	if focuser, ok := comp.(interface{ Focus() tea.Cmd }); ok {
		cmd := focuser.Focus()
		if cmd != nil {
			*cb.cmds = append(*cb.cmds, cmd)
		}
	} else if focuser, ok := comp.(interface{ Focus() }); ok {
		focuser.Focus()
	}
}

func (cb *commandCallback) Custom(actionName string, data map[string]string) {
	handler, ok := command.CustomActionHandlers[actionName]
	if !ok {
		return
	}
	cmds := handler(&appContext{app: cb.app}, cb)
	*cb.cmds = append(*cb.cmds, cmds...)
}

// ComponentOrder returns components in layout order (for focus navigation).
func (c *appContext) ComponentOrder() []string {
	var order []string
	for _, row := range c.app.Layout.Rows {
		order = append(order, row.Components...)
	}
	return order
}

// FocusedComponent returns the currently focused component name.
func (c *appContext) FocusedComponent() string {
	return c.app.focusedComponent
}
