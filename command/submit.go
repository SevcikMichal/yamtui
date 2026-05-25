package command

import (
	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/builder"
)

// SubmitCommand appends input text to the output textarea.
type SubmitCommand struct{}

func (c SubmitCommand) Execute(app *builder.App) tea.Cmd {
	inputComp, ok := app.Components["input"]
	if !ok {
		return nil
	}
	outputComp, ok := app.Components["textarea"]
	if !ok {
		return nil
	}

	inputVal := inputComp.(builder.TextEditor).Value()
	if inputVal == "" {
		return nil
	}

	outputVal := outputComp.(builder.TextEditor).Value()
	if outputVal != "" {
		outputComp.(builder.TextEditor).SetValue(outputVal + "\n" + inputVal)
	} else {
		outputComp.(builder.TextEditor).SetValue(inputVal)
	}
	outputComp.(builder.TextScroller).MoveToEnd()
	inputComp.(builder.TextEditor).SetValue("")

	return nil
}

func init() {
	builder.RegisterCommand("submit", func() builder.Command {
		return SubmitCommand{}
	})
}
