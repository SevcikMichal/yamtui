// Package runtime provides the entry point for running YAML-configured
// TUI applications built on Bubbletea.
package runtime

import (
	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/app"
	"github.com/SevcikMichal/yamtui/loader"
)

// Run creates and runs the application from a YAML config file.
// It loads the config, builds the app, and starts the Bubbletea program.
// External consumers should import the command package to register
// built-in commands via init() functions before calling Run().
func Run(configPath string) error {
	var cfg *loader.Configuration
	if configPath != "" {
		var err error
		cfg, err = loader.Load(configPath)
		if err != nil {
			return err
		}
	} else {
		cfg = &loader.Configuration{}
	}

	app, err := app.BuildApp(cfg)
	if err != nil {
		return err
	}

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
