package command

import (
	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/builder"
)

// ClearCommand clears the text content of a specified component.
type ClearCommand struct {
	Target string
}

func (c ClearCommand) Execute(app *builder.App) tea.Cmd {
	if c.Target == "" {
		return nil
	}

	comp, ok := app.Components[c.Target]
	if !ok {
		return nil
	}

	editor, ok := comp.(builder.TextEditor)
	if !ok {
		return nil
	}

	editor.SetValue("")
	return nil
}

func init() {
	builder.RegisterCommand("clear", func() builder.Command {
		return ClearCommand{}
	})
}
