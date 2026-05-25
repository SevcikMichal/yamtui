package command

import (
	"github.com/SevcikMichal/yamtui/builder"
)

// RegisterBuiltinCommands registers all built-in commands with the builder.
func RegisterBuiltinCommands() {
	builder.RegisterCommand("quit", func() builder.Command {
		return QuitCommand{}
	})
	builder.RegisterCommand("submit", func() builder.Command {
		return SubmitCommand{}
	})
	builder.RegisterCommand("clear", func() builder.Command {
		return ClearCommand{}
	})
	builder.RegisterCommand("focus", func() builder.Command {
		return FocusCommand{}
	})
	builder.RegisterCommand("sequence", func() builder.Command {
		return SequenceCommand{}
	})
}
