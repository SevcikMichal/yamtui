package command

import (
	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/builder"
)

// QuitCommand quits the application.
type QuitCommand struct{}

func (c QuitCommand) Execute(app *builder.App) tea.Cmd {
	return tea.Quit
}

func init() {
	builder.RegisterCommand("quit", func() builder.Command {
		return QuitCommand{}
	})
}
