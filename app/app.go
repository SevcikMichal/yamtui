// Package app provides a complete declarative UI application framework
// built on Bubbletea. It encapsulates the application model, component
// system, command registry, and entry point — external consumers do not
// need to import Bubbletea directly.
package app

import (
	"fmt"
	"image/color"
	"strings"

	ansicolor "github.com/charmbracelet/x/ansi"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/SevcikMichal/yamtui/command"
	"github.com/SevcikMichal/yamtui/component"
	"github.com/SevcikMichal/yamtui/internal/layout"
	"github.com/SevcikMichal/yamtui/internal/loader"
	"github.com/SevcikMichal/yamtui/theme"
)

// App is the main application struct that implements tea.Model.
type App struct {
	AltScreen  bool
	Components map[string]component.Component
	Commands   map[string]command.Command
	KeyMap     map[string]string // key string -> command name
	Layout     loader.LayoutConfig
	Theme      *theme.Theme // optional: theme for styling components

	// Runtime state
	TermWidth        int
	TermHeight       int
	focusedComponent string           // currently focused component name
	LayoutSizes      map[string]layout.Size // last computed sizes per component
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
// The themeRegistry is used to look up named themes or build from inline config.
// Pass nil to skip theming.
func BuildApp(cfg *loader.Configuration, themeRegistry *theme.ThemeRegistry) (*App, error) {
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

	// Build theme.
	var th *theme.Theme
	if themeRegistry != nil {
		th, err = buildTheme(themeRegistry, cfg.Theme)
		if err != nil {
			return nil, fmt.Errorf("building theme: %w", err)
		}
	}
	// log.Printf("[app] BuildApp: theme=%v, cfg.Theme.Name=%q, cfg.Theme.Default=%d, cfg.Theme.Focused=%d, cfg.Theme.Components=%d",
	//      th != nil, cfg.Theme.Name, len(cfg.Theme.Default), len(cfg.Theme.Focused), len(cfg.Theme.Components))

	return &App{
		AltScreen:  cfg.AltScreen,
		Components: components,
		Commands:   commands,
		KeyMap:     keyMap,
		Layout:     cfg.Layout,
		Theme:      th,
	}, nil
}

// buildTheme creates a theme from config.
func buildTheme(registry *theme.ThemeRegistry, cfg loader.ThemeConfig) (*theme.Theme, error) {
	// If name is set, try to load built-in theme first
	if cfg.Name != "" {
		if th := registry.Get(cfg.Name); th != nil {
			// If there are also inline overrides, build a derived theme
			if hasInlineOverrides(cfg) {
				return theme.BuildFromConfig(registry, cfg)
			}
			return th, nil
		}
	}

	// Build from inline config (or override)
	return theme.BuildFromConfig(registry, cfg)
}

// hasInlineOverrides checks if the theme config has any inline style definitions.
func hasInlineOverrides(cfg loader.ThemeConfig) bool {
	return len(cfg.Colors) > 0 ||
		len(cfg.Default) > 0 ||
		len(cfg.Focused) > 0 ||
		len(cfg.Error) > 0 ||
		len(cfg.Styles) > 0 ||
		len(cfg.Components) > 0 ||
		cfg.Base != ""
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
// It collects init cmds from every component (e.g. spinner tick) and focuses
// the first focusable component in layout order.
func (a *App) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Collect init commands from all components in layout order.
	// This ensures animated components (spinners, progress bars) start ticking.
	for _, row := range a.Layout.Rows {
		for _, name := range row.Components {
			comp, ok := a.Components[name]
			if !ok {
				continue
			}
			if initer, ok := comp.(interface{ Init() tea.Cmd }); ok {
				if cmd := initer.Init(); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	// Focus the first focusable component in layout order.
	for _, row := range a.Layout.Rows {
		for _, name := range row.Components {
			comp, ok := a.Components[name]
			if !ok {
				continue
			}
			if focuser, ok := comp.(interface{ Focus() tea.Cmd }); ok {
				a.focusedComponent = name
				if cmd := focuser.Focus(); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return tea.Batch(cmds...)
			}
			if focuser, ok := comp.(interface{ Focus() }); ok {
				a.focusedComponent = name
				focuser.Focus()
				return tea.Batch(cmds...)
			}
		}
	}
	return tea.Batch(cmds...)
}

// Update processes messages and delegates to handlers and components.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var allCmds []tea.Cmd

	// Handle window resize.
	if wmsg, ok := msg.(tea.WindowSizeMsg); ok {
		a.TermWidth = wmsg.Width
		a.TermHeight = wmsg.Height
		a.LayoutSizes = layout.CalculateLayout(a.TermWidth, a.TermHeight, a.Layout)
		a.applyComponentSizes(a.focusedComponent)
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
				handler(ctx, cb)
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
					handler(ctx, cb)
					keyConsumed = true
				}
			}
		}
	}

	// Dispatch messages to components.
	// Key messages consumed by a command (e.g. Tab for focus) are not forwarded
	// to components so they don't land in text inputs. All other messages —
	// ticks, window events, custom msgs — are always forwarded so that animated
	// components (spinners, progress bars, stopwatches) keep running regardless
	// of whether a key was handled.
	_, isKeyMsg := msg.(tea.KeyPressMsg)
	if !isKeyMsg || !keyConsumed {
		for name, c := range a.Components {
			newComp, cmd := c.Update(msg)
			if newComp != nil {
				a.Components[name] = newComp
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

	var canvasBg ansicolor.Color
	if a.Theme != nil {
		canvasBg = a.Theme.Default.GetStyle().GetBackground()
	}

	for _, row := range a.Layout.Rows {
		var cols []string
		for _, name := range row.Components {
			if c, ok := a.Components[name]; ok {
				view := c.View()

				// Determine the lipgloss style to apply (empty if no theme).
				var ls lipgloss.Style
				if a.Theme != nil {
					var s theme.Style
					if name == a.focusedComponent {
						s = a.Theme.GetFocusedStyle(name)
					} else {
						s = a.Theme.GetStyle(name)
					}
					ls = s.GetStyle()
				} else {
					ls = lipgloss.NewStyle()
				}

				// Set Width/Height on the style so the allocated cell is
				// filled exactly. The component was already sized to
				// contentW = sz.Width - frame, so ls.Width(contentW) matches
				// the rendered content and no clipping or wrapping occurs.
				// Height fills the cell with the background color for short
				// components (inputs, spinners).
				if sz, ok := a.LayoutSizes[name]; ok {
					cw := sz.Width - ls.GetHorizontalFrameSize()
					ch := sz.Height - ls.GetVerticalFrameSize()
					if cw < 0 {
						cw = 0
					}
					if ch < 0 {
						ch = 0
					}
					ls = ls.Width(cw).Height(ch)
				}

				view = ls.Render(view)
				cols = append(cols, view)
			}
		}
		rows = append(rows, joinCols(cols, row.Spacing, canvasBg))
	}

	layoutStr := joinRows(rows)

	// Wrap the entire layout in a full-terminal background-coloured container
	// so the theme's canvas colour fills margins, inter-row gaps, and any
	// unused space instead of letting the terminal's own background bleed
	// through. The padding (1 line top+bottom, 2 chars left+right) matches
	// the outerVPad/outerHPad constants used by the layout calculator.
	if a.Theme != nil && a.TermWidth > 0 && a.TermHeight > 0 {
		bg := a.Theme.Default.GetStyle().GetBackground()
		// Width/Height set the *content* area; Padding(1,2) adds 1 line
		// top+bottom (outerVPad=2) and 2 chars left+right (outerHPad=4)
		// on each side, so total rendered area fills the terminal exactly.
		rootStyle := lipgloss.NewStyle().
			Width(a.TermWidth).
			Height(a.TermHeight).
			Background(bg)
		layoutStr = rootStyle.Render(layoutStr)
	}

	return tea.View{
		Content:   layoutStr,
		AltScreen: a.AltScreen,
	}
}

// joinCols joins column views with spacing between them, aligning
// each column's lines side-by-side so multi-line components render
// correctly in a grid layout. bg is the canvas background colour used
// to paint the inter-column gap so the terminal's own colour doesn't bleed.
func joinCols(cols []string, spacing float64, bg ansicolor.Color) string {
	if len(cols) == 0 {
		return ""
	}
	if len(cols) == 1 {
		return cols[0]
	}
	if spacing <= 0 {
		spacing = 1
	}
	// Paint the separator with the canvas background colour so the gap
	// doesn't expose the terminal's own background.
	gapStyle := lipgloss.NewStyle()
	if bg != nil {
		gapStyle = gapStyle.Background(bg)
	}
	sep := gapStyle.Render(strings.Repeat(" ", int(spacing)))

	// Split each column into lines.
	lines := make([][]string, len(cols))
	for i, col := range cols {
		lines[i] = strings.Split(col, "\n")
	}

	// Measure the visible cell width of each column (max across all its lines).
	// This uses lipgloss.Width which strips ANSI sequences before measuring,
	// ensuring columns stay aligned even when styled with colors/borders.
	colWidths := make([]int, len(cols))
	for colIdx, colLines := range lines {
		for _, line := range colLines {
			if w := lipgloss.Width(line); w > colWidths[colIdx] {
				colWidths[colIdx] = w
			}
		}
	}

	// Find the maximum number of lines across all columns.
	maxLines := 0
	for _, l := range lines {
		if len(l) > maxLines {
			maxLines = len(l)
		}
	}

	// Build each output line by concatenating the corresponding line
	// from each column. Non-last columns are right-padded with spaces to
	// their measured width so subsequent columns stay horizontally aligned.
	var result []string
	lastCol := len(lines) - 1
	for row := 0; row < maxLines; row++ {
		var parts []string
		for colIdx := range lines {
			var line string
			if row < len(lines[colIdx]) {
				line = lines[colIdx][row]
			}
			// Pad non-last columns to their full width.
			if colIdx < lastCol {
				if pad := colWidths[colIdx] - lipgloss.Width(line); pad > 0 {
					line += strings.Repeat(" ", pad)
				}
			}
			parts = append(parts, line)
		}
		result = append(result, strings.Join(parts, sep))
	}

	return strings.Join(result, "\n")
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
	// Re-apply component sizes so the newly focused component is sized to
	// its focused-style content area (frame may differ, e.g. border added).
	cb.app.applyComponentSizes(name)
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
	handler(&appContext{app: cb.app}, cb)
}

func (cb *commandCallback) SetContent(componentName, content string) {
	comp, ok := cb.app.Components[componentName]
	if !ok {
		return
	}
	if setter, ok := comp.(interface{ SetContent(string) }); ok {
		setter.SetContent(content)
	}
}

// applyComponentSizes pushes content dimensions (allocated size minus the
// style frame) into each component that accepts SetSize or SetWidth.
// focusedName is used to select the correct style (focused vs unfocused)
// so the content area is computed against the actual frame that will be
// rendered, preventing the style from clipping or wrapping component content.
func (a *App) applyComponentSizes(focusedName string) {
	for name, size := range a.LayoutSizes {
		comp, ok := a.Components[name]
		if !ok {
			continue
		}

		contentW := size.Width
		contentH := size.Height

		if a.Theme != nil {
			var s theme.Style
			if name == focusedName {
				s = a.Theme.GetFocusedStyle(name)
			} else {
				s = a.Theme.GetStyle(name)
			}
			ls := s.GetStyle()
			contentW -= ls.GetHorizontalFrameSize()
			contentH -= ls.GetVerticalFrameSize()
			if contentW < 0 {
				contentW = 0
			}
			if contentH < 0 {
				contentH = 0
			}
		}

		if sc, ok := comp.(interface{ SetSize(int, int) }); ok {
			sc.SetSize(contentW, contentH)
		}
		if sw, ok := comp.(interface{ SetWidth(int) }); ok {
			sw.SetWidth(contentW)
		}
		if sh, ok := comp.(interface{ SetHeight(int) }); ok {
			sh.SetHeight(contentH)
		}

		// Push the panel background colour into the component's internal
		// styles (textarea, textinput) so their rendered text uses the
		// theme background instead of the terminal default.
		if a.Theme != nil {
			var s theme.Style
			if name == focusedName {
				s = a.Theme.GetFocusedStyle(name)
			} else {
				s = a.Theme.GetStyle(name)
			}
			bg := s.GetStyle().GetBackground()
			if bg != nil {
				if setter, ok := comp.(interface{ SetBackground(color.Color) }); ok {
					setter.SetBackground(bg)
				}
			}
		}
	}
}

// ComponentOrder returns focusable components in layout order.
// Non-focusable components (e.g. text labels) are excluded so Tab
// navigation only cycles through interactive elements.
func (c *appContext) ComponentOrder() []string {
	var order []string
	for _, row := range c.app.Layout.Rows {
		for _, name := range row.Components {
			comp, ok := c.app.Components[name]
			if !ok {
				continue
			}
			_, hasFocusCmd := comp.(interface{ Focus() tea.Cmd })
			_, hasFocus := comp.(interface{ Focus() })
			if hasFocusCmd || hasFocus {
				order = append(order, name)
			}
		}
	}
	return order
}

// FocusedComponent returns the currently focused component name.
func (c *appContext) FocusedComponent() string {
	return c.app.focusedComponent
}
