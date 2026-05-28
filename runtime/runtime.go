// Package runtime provides the entry point for running YAML-configured
// TUI applications built on Bubbletea.
package runtime

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/app"
	"github.com/SevcikMichal/yamtui/internal/loader"
	"github.com/SevcikMichal/yamtui/theme"
)

// globalRegistry is the shared theme registry initialized once.
var globalRegistry *theme.ThemeRegistry

// globalLogFile is used to flush logs at the end of Run().
var globalLogFile *os.File

func init() {
	// Set up log file FIRST so all theme construction logs go to file, not console.
	var err error
	globalLogFile, err = os.OpenFile("/tmp/yamtui-theme-debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		log.SetOutput(globalLogFile)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	globalRegistry = theme.NewThemeRegistry()
}

// Run creates and runs the application from a YAML config file.
// It loads the config, builds the app, and starts the Bubbletea program.
// External consumers should import the command package to register
// built-in commands via init() functions before calling Run().
func Run(configPath string) error {
	log.Printf("[runtime] Run START: configPath=%s", configPath)
	defer func() {
		if globalLogFile != nil {
			globalLogFile.Close()
		}
	}()

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

	a, err := app.BuildApp(cfg, globalRegistry)
	if err != nil {
		return err
	}

	p := tea.NewProgram(a)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
