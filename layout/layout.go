// Package layout provides dimension calculation for YAML-configured components.
package layout

import (
	"github.com/SevcikMichal/yamtui/loader"
)

// Size holds the computed dimensions for a component.
type Size struct {
	Width  int
	Height int
}

// CalculateLayout computes dimensions for all components based on terminal size
// and YAML layout configuration.
func CalculateLayout(termWidth, termHeight int, layoutConfig loader.LayoutConfig) map[string]Size {
	result := make(map[string]Size, len(layoutConfig.Order))

	// Subtract theme padding/borders from available space.
	// Default estimates: outer padding 2h+2v, border 2h, gaps 1 each
	outerHPad := 4
	outerVPad := 2
	borderHPad := 2

	availWidth := termWidth - outerHPad - borderHPad - 2
	availHeight := termHeight - outerVPad

	// Separate fixed and ratio components.
	var fixedOrder, ratioOrder []string
	for _, name := range layoutConfig.Order {
		sizing, hasSizing := layoutConfig.Sizing[name]
		if !hasSizing {
			// Default: fill remaining space (like ratio 1.0).
			ratioOrder = append(ratioOrder, name)
			continue
		}
		switch sizing.Height {
		case "fixed":
			fixedOrder = append(fixedOrder, name)
		case "ratio", "fill":
			ratioOrder = append(ratioOrder, name)
		}
	}

	// Allocate fixed-height components first.
	fixedHeight := 0
	for _, name := range fixedOrder {
		sizing := layoutConfig.Sizing[name]
		h := int(sizing.Value)
		if h < 1 {
			h = 1
		}
		result[name] = Size{Width: availWidth, Height: h}
		fixedHeight += h
	}

	// Account for gaps between components (1 newline each).
	gaps := len(fixedOrder) + len(ratioOrder) - 1
	if gaps < 0 {
		gaps = 0
	}
	availHeight -= gaps

	// Distribute remaining height to ratio components.
	if len(ratioOrder) > 0 && availHeight > 0 {
		// Calculate total ratio weight.
		totalRatio := 0.0
		for _, name := range ratioOrder {
			sizing := layoutConfig.Sizing[name]
			if sizing.Height == "fill" {
				totalRatio += 1.0
			} else {
				totalRatio += sizing.Value
			}
		}

		remainingHeight := availHeight
		for i, name := range ratioOrder {
			sizing := layoutConfig.Sizing[name]
			var weight float64
			if sizing.Height == "fill" {
				weight = 1.0
			} else {
				weight = sizing.Value
			}

			h := int(float64(availHeight) * weight / totalRatio)
			if h < 1 {
				h = 1
			}

			// Give remaining components any fractional remainder.
			if i == len(ratioOrder)-1 {
				h = remainingHeight - (fixedHeight + calculateFixedHeight(layoutConfig, ratioOrder[:i+1]))
				if h < 1 {
					h = 1
				}
			}

			result[name] = Size{Width: availWidth, Height: h}
			remainingHeight -= h
		}
	}

	return result
}

// calculateFixedHeight returns the sum of fixed heights for the given component names.
func calculateFixedHeight(layoutConfig loader.LayoutConfig, names []string) int {
	total := 0
	for _, name := range names {
		sizing, ok := layoutConfig.Sizing[name]
		if !ok {
			continue
		}
		if sizing.Height == "fixed" {
			h := int(sizing.Value)
			if h < 1 {
				h = 1
			}
			total += h
		}
	}
	return total
}
