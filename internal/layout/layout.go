// Package layout provides dimension calculation for YAML-configured components.
package layout

import (
	"github.com/SevcikMichal/yamtui/internal/loader"
)

// Size holds the computed dimensions for a component.
type Size struct {
	Width  int
	Height int
}

// Default spacing and padding constants.
const (
	defaultColSpacing = 1 // gap between columns in chars
	defaultRowSpacing = 1 // gap between rows in lines
	outerHPad         = 2 // horizontal padding (1 char each side)
	outerVPad         = 1 // vertical padding (1 line top+bottom)
)

// CalculateLayout computes dimensions for all components based on terminal size
// and YAML grid layout configuration.
func CalculateLayout(termWidth, termHeight int, layoutConfig loader.LayoutConfig) map[string]Size {
	return calculateGridLayout(termWidth, termHeight, layoutConfig)
}

// calculateGridLayout implements the 2D grid layout algorithm.
func calculateGridLayout(termWidth, termHeight int, layoutConfig loader.LayoutConfig) map[string]Size {
	availWidth := termWidth - outerHPad
	availHeight := termHeight - outerVPad

	result := make(map[string]Size)

	// Phase 1: Allocate row heights.
	// Subtract row gaps before allocation so heights are computed on the actual usable space.
	availHeight -= rowGaps(len(layoutConfig.Rows))
	if availHeight < 0 {
		availHeight = 0
	}
	rowHeights := allocateRowHeights(availHeight, layoutConfig)

	// Phase 2: Allocate column width within each row.
	for rowIndex, row := range layoutConfig.Rows {
		spacing := row.Spacing
		if spacing <= 0 {
			spacing = defaultColSpacing
		}

		// Calculate total column gaps for this row.
		colCount := len(row.Components)
		var colGaps int
		if colCount > 1 {
			colGaps = (colCount - 1) * int(spacing)
		}

		availColWidth := availWidth - colGaps
		if availColWidth < 0 {
			availColWidth = 0
		}

		colWidths := allocateColWidths(availColWidth, row.Components, layoutConfig.Sizing)

		for colIndex, compName := range row.Components {
			h := rowHeights[rowIndex]
			if h < 1 {
				h = 1
			}
			w := colWidths[colIndex]
			if w < 0 {
				w = 0
			}
			result[compName] = Size{Width: w, Height: h}
		}
	}

	return result
}

// allocateRowHeights distributes height among rows based on their sizing configs.
func allocateRowHeights(availHeight int, layoutConfig loader.LayoutConfig) []int {
	rowHeights := make([]int, len(layoutConfig.Rows))

	// Separate fixed-height rows from flexible ones.
	var fixedRows, flexibleRows []int
	fixedHeight := 0

	for i, row := range layoutConfig.Rows {
		rowHeight := getRowHeight(row.Components, layoutConfig.Sizing)
		switch rowHeight.typ {
		case "fixed":
			fixedRows = append(fixedRows, i)
			fixedHeight += rowHeight.value
		case "ratio", "fill":
			flexibleRows = append(flexibleRows, i)
		}
	}

	// Allocate fixed heights.
	for _, i := range fixedRows {
		row := layoutConfig.Rows[i]
		h := getRowHeight(row.Components, layoutConfig.Sizing)
		if h.value < 1 {
			h.value = 1
		}
		rowHeights[i] = h.value
	}

	// Distribute remaining height to flexible rows.
	remainingHeight := availHeight - fixedHeight
	if len(flexibleRows) > 0 && remainingHeight > 0 {
		// Calculate total weight.
		totalWeight := 0.0
		for _, i := range flexibleRows {
			row := layoutConfig.Rows[i]
			w := getRowHeight(row.Components, layoutConfig.Sizing)
			totalWeight += rowWeight(w)
		}

		allocated := 0
		for idx, i := range flexibleRows {
			row := layoutConfig.Rows[i]
			h := getRowHeight(row.Components, layoutConfig.Sizing)
			weight := rowWeight(h)

			var allocatedH int
			if totalWeight > 0 {
				allocatedH = int(float64(remainingHeight) * weight / totalWeight)
			}

			if allocatedH < 1 {
				allocatedH = 1
			}

			// Last flexible row gets the remainder to avoid rounding errors.
			if idx == len(flexibleRows)-1 {
				allocatedH = remainingHeight - allocated
				if allocatedH < 1 {
					allocatedH = 1
				}
			}

			rowHeights[i] = allocatedH
			allocated += allocatedH
		}
	}

	return rowHeights
}

// rowHeightResult holds the computed type and value for a row's height.
type rowHeightResult struct {
	typ   string
	value int
}

// getRowHeight determines the dominant sizing for a row based on its components.
// If any component has a fixed height, the row is fixed.
// Otherwise, if any component has a ratio height, the row uses ratio.
// If all components have fill height, the row is fill.
func getRowHeight(components []string, sizing map[string]loader.SizeConfig) rowHeightResult {
	hasRatio := false

	for _, name := range components {
		cfg, ok := sizing[name]
		if !ok || cfg.Height == nil {
			// Default: fill for rows.
			continue
		}

		switch cfg.Height.Type {
		case "fixed":
			return rowHeightResult{typ: "fixed", value: int(cfg.Height.Value)}
		case "ratio":
			hasRatio = true
		case "fill":
			// Continue checking other components.
		}
	}

	if hasRatio {
		// Return the max ratio value among components.
		maxRatio := 0.0
		for _, name := range components {
			cfg, ok := sizing[name]
			if ok && cfg.Height != nil && cfg.Height.Type == "ratio" {
				if cfg.Height.Value > maxRatio {
					maxRatio = cfg.Height.Value
				}
			}
		}
		return rowHeightResult{typ: "ratio", value: int(maxRatio * 10)} // scale for ratio tracking
	}

	return rowHeightResult{typ: "fill", value: 1}
}

// rowWeight converts a rowHeightResult to a float64 weight for distribution.
func rowWeight(r rowHeightResult) float64 {
	switch r.typ {
	case "fixed":
		return float64(r.value)
	case "ratio":
		return float64(r.value) / 10.0
	case "fill":
		return 1.0
	default:
		return 1.0
	}
}

// allocateColWidths distributes width among columns in a row based on their sizing configs.
func allocateColWidths(availWidth int, components []string, sizing map[string]loader.SizeConfig) []int {
	colWidths := make([]int, len(components))

	// Separate fixed-width columns from flexible ones.
	var fixedCols, flexibleCols []int
	fixedWidth := 0

	for i, name := range components {
		cfg, ok := sizing[name]
		if !ok || cfg.Width == nil {
			// Default: fill.
			flexibleCols = append(flexibleCols, i)
			continue
		}

		switch cfg.Width.Type {
		case "fixed":
			fixedCols = append(fixedCols, i)
			fixedWidth += int(cfg.Width.Value)
		case "ratio", "fill":
			flexibleCols = append(flexibleCols, i)
		}
	}

	// Allocate fixed widths.
	for _, i := range fixedCols {
		cfg := sizing[components[i]]
		w := int(cfg.Width.Value)
		if w < 1 {
			w = 1
		}
		colWidths[i] = w
	}

	// Distribute remaining width to flexible columns.
	remainingWidth := availWidth - fixedWidth
	if len(flexibleCols) > 0 && remainingWidth > 0 {
		// Calculate total weight.
		totalWeight := 0.0
		for _, i := range flexibleCols {
			name := components[i]
			cfg := sizing[name]
			totalWeight += colWidthWeight(cfg)
		}

		allocated := 0
		for idx, i := range flexibleCols {
			name := components[i]
			cfg := sizing[name]
			weight := colWidthWeight(cfg)

			var allocatedW int
			if totalWeight > 0 {
				allocatedW = int(float64(remainingWidth) * weight / totalWeight)
			}

			if allocatedW < 1 {
				allocatedW = 1
			}

			// Last flexible column gets the remainder to avoid rounding errors.
			if idx == len(flexibleCols)-1 {
				allocatedW = remainingWidth - allocated
				if allocatedW < 1 {
					allocatedW = 1
				}
			}

			colWidths[i] = allocatedW
			allocated += allocatedW
		}
	}

	return colWidths
}

// colWidthWeight returns the weight for a component's column width.
func colWidthWeight(cfg loader.SizeConfig) float64 {
	if cfg.Width == nil {
		return 1.0 // default: fill
	}

	switch cfg.Width.Type {
	case "fixed":
		return float64(int(cfg.Width.Value))
	case "ratio":
		return cfg.Width.Value
	case "fill":
		return 1.0
	default:
		return 1.0
	}
}

// rowGaps returns the total spacing between rows.
func rowGaps(numRows int) int {
	if numRows <= 1 {
		return 0
	}
	return (numRows - 1) * defaultRowSpacing
}
