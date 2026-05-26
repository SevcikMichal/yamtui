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
	"os"
	"path/filepath"
	stdruntime "runtime"

	"github.com/SevcikMichal/yamtui/component"
	yruntime "github.com/SevcikMichal/yamtui/runtime"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	exampleName := os.Args[1]

	// Initialize the component registry with built-in types.
	component.Init()

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
