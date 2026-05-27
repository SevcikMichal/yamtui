// Package theme provides a declarative theme system for yamtui using lipgloss/v2.
// Themes are defined in YAML and applied globally to all components.
package theme

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// knownColors maps named colors to their hex string values for lipgloss.Color().
var knownColors = map[string]string{
	// ANSI basic
	"black":   "#000000",
	"red":     "#AA0000",
	"green":   "#00AA00",
	"yellow":  "#AA5500",
	"blue":    "#0000AA",
	"magenta": "#AA00AA",
	"cyan":    "#00AAAA",
	"white":   "#BBBBBB",

	// ANSI bright
	"bright_black":   "#666666",
	"bright_red":     "#FF5555",
	"bright_green":   "#55FF55",
	"bright_yellow":  "#FFFF55",
	"bright_blue":    "#5555FF",
	"bright_magenta": "#FF55FF",
	"bright_cyan":    "#55FFFF",
	"bright_white":   "#FFFFFF",

	// Common tea/lipgloss names (aliases)
	"purple": "#AA00AA",
	"pink":   "#FF66CC",
	"orange": "#FF8800",
	"teal":   "#00AAAA",
}

// ColorPalette holds named color values resolved from YAML.
type ColorPalette struct {
	colors map[string]ansi.Color
}

// NewColorPalette creates a new empty ColorPalette.
func NewColorPalette() *ColorPalette {
	return &ColorPalette{
		colors: make(map[string]ansi.Color),
	}
}

// Set adds a color to the palette. The value is a string that can be:
// - A hex color: "#FF5733" or "#FF573380" (with alpha)
// - A named color: "red", "blue", "teal", etc.
// - An ANSI index: "1", "21", etc.
func (p *ColorPalette) Set(name string, value string) error {
	color := lipgloss.Color(value)
	p.colors[name] = color
	return nil
}

// Get resolves a color name to ansi.Color.
// It first checks the palette, then known colors, then tries to parse as hex/ANSI.
func (p *ColorPalette) Get(name string) ansi.Color {
	// Check palette first
	if c, ok := p.colors[name]; ok {
		return c
	}

	// Check known colors
	if hex, ok := knownColors[strings.ToLower(name)]; ok {
		return lipgloss.Color(hex)
	}

	// Try parsing as hex or ANSI
	return lipgloss.Color(name)
}

// Has checks if a color name exists in the palette.
func (p *ColorPalette) Has(name string) bool {
	_, ok := p.colors[name]
	return ok
}
