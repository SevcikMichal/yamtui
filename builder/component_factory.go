package builder

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/loader"
)

// componentConstructors maps YAML component type strings to constructor functions.
var componentConstructors = map[string]func(loader.CompConfig) Component{
	"textarea": func(cfg loader.CompConfig) Component {
		ta := textarea.New()
		if cfg.Placeholder != "" {
			ta.Placeholder = cfg.Placeholder
		}
		if cfg.Config.CharLimit > 0 {
			ta.CharLimit = cfg.Config.CharLimit
		}
		if cfg.Config.ReadOnly {
			ta.Blur()
		}
		return &yamlTextareaComponent{model: ta}
	},
	"input": func(cfg loader.CompConfig) Component {
		ti := textinput.New()
		if cfg.Placeholder != "" {
			ti.Placeholder = cfg.Placeholder
		}
		if cfg.Config.CharLimit > 0 {
			ti.CharLimit = cfg.Config.CharLimit
		}
		if cfg.Config.Focus {
			ti.Focus()
		}
		return &yamlInputComponent{model: ti}
	},
	"helpbar": func(cfg loader.CompConfig) Component {
		text := cfg.Text
		if text == "" {
			text = "(ctrl+c / esc to quit)"
		}
		return &yamlHelpBarComponent{text: text}
	},
}

// BuildComponents creates components from a YAML config map.
func BuildComponents(components map[string]loader.CompConfig) (map[string]Component, error) {
	result := make(map[string]Component, len(components))
	for name, cfg := range components {
		constructor, ok := componentConstructors[cfg.Type]
		if !ok {
			return nil, nil
		}
		result[name] = constructor(cfg)
	}
	return result, nil
}

// yamlTextareaComponent wraps the bubbletea textarea for YAML-configured output display.
type yamlTextareaComponent struct {
	model textarea.Model
}

func (c *yamlTextareaComponent) Name() string { return "textarea" }

func (c *yamlTextareaComponent) View() string {
	return c.model.View()
}

func (c *yamlTextareaComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	c.model, cmd = c.model.Update(msg)
	return c, cmd
}

func (c *yamlTextareaComponent) Value() string     { return c.model.Value() }
func (c *yamlTextareaComponent) SetValue(v string) { c.model.SetValue(v) }
func (c *yamlTextareaComponent) MoveToEnd()        { c.model.MoveToEnd() }
func (c *yamlTextareaComponent) SetSize(w, h int) {
	c.model.SetWidth(w)
	c.model.SetHeight(h)
}

// yamlInputComponent wraps the bubbletea textinput for YAML-configured user input.
type yamlInputComponent struct {
	model textinput.Model
}

func (c *yamlInputComponent) Name() string { return "input" }

func (c *yamlInputComponent) View() string {
	return c.model.View()
}

func (c *yamlInputComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	c.model, cmd = c.model.Update(msg)
	return c, cmd
}

func (c *yamlInputComponent) Value() string     { return c.model.Value() }
func (c *yamlInputComponent) SetValue(v string) { c.model.SetValue(v) }
func (c *yamlInputComponent) SetSize(w, h int) {
	c.model.SetWidth(w)
}

// yamlHelpBarComponent displays YAML-configured help text at the bottom of the UI.
type yamlHelpBarComponent struct {
	text string
}

func (c *yamlHelpBarComponent) Name() string { return "helpbar" }

func (c *yamlHelpBarComponent) View() string {
	return c.text
}

func (c *yamlHelpBarComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	return c, nil
}

func (c *yamlHelpBarComponent) SetSize(w, h int) {}
