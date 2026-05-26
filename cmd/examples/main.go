// Package main provides example applications demonstrating the yamtui framework.
//
// These examples showcase the YAML-driven declarative UI approach, ranging from
// simple single-component layouts to complex multi-panel dashboards with
// advanced sizing and command routing.
//
// Usage:
//
//	go run cmd/examples/main.go simple        # Basic input component
//	go run cmd/examples/main.go advanced      # Multi-component with focus management
//	go run cmd/examples/main.go dashboard     # Complex dashboard with layout sizing
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	stdruntime "runtime"

	tea "charm.land/bubbletea/v2"

	"github.com/SevcikMichal/yamtui/command"
	"github.com/SevcikMichal/yamtui/component"
	yruntime "github.com/SevcikMichal/yamtui/runtime"
)

func main() {
	// Set up logging to file
	logFile, err := os.Create("/tmp/yamtui-debug.log")
	if err == nil {
		log.SetOutput(logFile)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	} else {
		fmt.Fprintf(os.Stderr, "Warning: could not open log file: %v\n", err)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	exampleName := os.Args[1]

	// Initialize the component registry with all built-in bubble types.
	component.Init()

	// Register custom action handlers for focus navigation.
	command.RegisterCustomAction("focusNext", func(ctx command.AppContext, cb command.CommandCallback) []tea.Cmd {
		order := ctx.ComponentOrder()
		if len(order) == 0 {
			return nil
		}
		current := ctx.FocusedComponent()
		for i, name := range order {
			if name == current {
				next := order[(i+1)%len(order)]
				cb.Focus(next)
				return nil
			}
		}
		// Current focus not found in order, focus first component
		cb.Focus(order[0])
		return nil
	})

	command.RegisterCustomAction("focusPrev", func(ctx command.AppContext, cb command.CommandCallback) []tea.Cmd {
		order := ctx.ComponentOrder()
		if len(order) == 0 {
			return nil
		}
		current := ctx.FocusedComponent()
		for i, name := range order {
			if name == current {
				prev := order[(i-1+len(order))%len(order)]
				cb.Focus(prev)
				return nil
			}
		}
		// Current focus not found in order, focus last component
		if len(order) > 0 {
			cb.Focus(order[len(order)-1])
		}
		return nil
	})

	// Resolve the examples directory path relative to this source file.
	examplesDir := findExamplesDir()

	configPath := filepath.Join(examplesDir, exampleName+".yaml")

	// Check if the config file exists.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: example %q not found\n\n", exampleName)
		printUsage()
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Starting yamtui example: %s\n", exampleName)
	fmt.Fprintf(os.Stderr, "Config: %s\n\n", configPath)

	if err := yruntime.Run(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error running example: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	examplesDir := findExamplesDir()
	fmt.Fprintln(os.Stderr, "Available examples:")
	fmt.Fprintln(os.Stderr)

	files, err := os.ReadDir(examplesDir)
	if err == nil {
		for _, file := range files {
			if file.Name() != "main.go" && filepath.Ext(file.Name()) == ".yaml" {
				name := file.Name()[:len(file.Name())-5] // strip .yaml
				fmt.Fprintf(os.Stderr, "  %s\n", name)
			}
		}
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage: go run cmd/examples/main.go <example-name>")
}

// findExamplesDir returns the directory containing this source file.
// It uses runtime.Caller to resolve the path relative to the compiled binary
// or source location.
func findExamplesDir() string {
	// Get the directory of this source file.
	_, file, _, _ := stdruntime.Caller(0)
	return filepath.Dir(file)
}
