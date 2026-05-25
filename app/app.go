// Package app provides the high-level entry point for running
// the declarative UI application. It encapsulates all Bubbletea
// logic so external consumers do not need to import Bubbletea directly.
package app

import (
	"log"

	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/builder"
	"github.com/SevcikMichal/yamtui/command"
	"github.com/SevcikMichal/yamtui/loader"
)

// Run creates and runs the application from a YAML config file.
// It registers all built-in commands, loads the config, and starts
// the Bubbletea program. This is the stable entry point for external
// consumers — yalca and other apps should call this instead of
// importing Bubbletea directly.
func Run(configPath string) error {
	command.RegisterBuiltinCommands()

	cfg := loader.LoadDefaults()
	if configPath != "" {
		var err error
		cfg, err = loader.Load(configPath)
		if err != nil {
			log.Printf("Warning: failed to load config %q: %v (using defaults)", configPath, err)
			cfg = loader.LoadDefaults()
		}
	}

	app, err := builder.Build(cfg)
	if err != nil {
		return err
	}

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
