package command

import (
	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/builder"
)

// FocusCommand sets focus to a specified input component.
type FocusCommand struct {
	Target string
}

func (c FocusCommand) Execute(app *builder.App) tea.Cmd {
	if c.Target == "" {
		return nil
	}

	comp, ok := app.Components[c.Target]
	if !ok {
		return nil
	}

	_ = comp
	return nil
}

func init() {
	builder.RegisterCommand("focus", func() builder.Command {
		return FocusCommand{}
	})
}
