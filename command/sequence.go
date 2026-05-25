package command

import (
	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/builder"
	"github.com/SevcikMichal/yamtui/loader"
)

// SequenceCommand executes a sequence of sub-commands in order.
type SequenceCommand struct {
	Steps []StepCommand
}

// StepCommand represents a single step within a sequence.
type StepCommand struct {
	Action string
	Target string
	Value  string
	Into   string
}

func (c SequenceCommand) Execute(app *builder.App) tea.Cmd {
	var cmds []tea.Cmd

	for _, step := range c.Steps {
		switch step.Action {
		case "clear":
			cmd := ClearCommand{Target: step.Target}.Execute(app)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "focus":
			cmd := FocusCommand{Target: step.Target}.Execute(app)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "set_value":
			cmd := SetValueCommand{Target: step.Target, Value: step.Value}.Execute(app)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "append_value":
			cmd := AppendValueCommand{Target: step.Target, Value: step.Value}.Execute(app)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return tea.Batch(cmds...)
}

// BuildSequenceCommands creates StepCommand slices from YAML step configs.
func BuildSequenceCommands(steps []loader.StepConfig) []StepCommand {
	result := make([]StepCommand, len(steps))
	for i, s := range steps {
		result[i] = StepCommand{
			Action: s.Action,
			Target: s.Target,
			Value:  s.Value,
			Into:   s.Into,
		}
	}
	return result
}

// SetValueCommand sets the value of a component.
type SetValueCommand struct {
	Target string
	Value  string
}

func (c SetValueCommand) Execute(app *builder.App) tea.Cmd {
	comp, ok := app.Components[c.Target]
	if !ok {
		return nil
	}

	editor, ok := comp.(builder.TextEditor)
	if !ok {
		return nil
	}

	editor.SetValue(c.Value)
	return nil
}

// AppendValueCommand appends a value to a component.
type AppendValueCommand struct {
	Target string
	Value  string
}

func (c AppendValueCommand) Execute(app *builder.App) tea.Cmd {
	comp, ok := app.Components[c.Target]
	if !ok {
		return nil
	}

	editor, ok := comp.(builder.TextEditor)
	if !ok {
		return nil
	}

	current := editor.Value()
	if current != "" {
		editor.SetValue(current + "\n" + c.Value)
	} else {
		editor.SetValue(c.Value)
	}

	return nil
}

func init() {
	builder.RegisterCommand("sequence", func() builder.Command {
		return SequenceCommand{}
	})
}
