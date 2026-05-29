package theme

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Style wraps lipgloss.Style and provides reflection-based property setting.
type Style struct {
	s       lipgloss.Style
	palette *ColorPalette
	defined bool
}

// NewStyle creates a new empty Style.
func NewStyle() Style {
	return Style{s: lipgloss.NewStyle()}
}

// SetPalette sets the color palette for this style.
func (s *Style) SetPalette(p *ColorPalette) {
	s.palette = p
}

// Render returns the styled string.
func (s Style) Render(strs ...string) string {
	return s.s.Render(strs...)
}

// GetStyle returns the underlying lipgloss.Style.
func (s Style) GetStyle() lipgloss.Style {
	return s.s
}

// colorProps lists property names that accept color values.
var colorProps = map[string]bool{
	"color":             true,
	"foreground":        true,
	"background":        true,
	"border_color":      true,
	"border_background": true,
}

// resolveColorValue converts a color value string to an ansi.Color using the palette.
// If the value is already an ansi.Color, it is returned as-is.
func (s Style) resolveColorValue(v any) any {
	switch val := v.(type) {
	case string:
		// Check palette first
		if s.palette != nil && s.palette.Has(val) {
			return s.palette.Get(val)
		}
		// Check known colors
		if hex, ok := knownColors[strings.ToLower(val)]; ok {
			return lipgloss.Color(hex)
		}
		// Try parsing as hex or ANSI index
		return lipgloss.Color(val)
	case ansi.Color:
		return val
	default:
		return v
	}
}

// SetProperty sets a style property by name using reflection-like dispatch.
// Supported properties:
//
//	color / foreground - text foreground color (color name or hex)
//	background - background color (color name or hex)
//	bold - bold text
//	italic - italic text
//	underline - underlined text
//	strikethrough - strikethrough text
//	reverse - reverse colors
//	blink - blinking text
//	dim / faint - dim/faint text
//	padding - [top, right, bottom, left] int array
//	margin - [top, right, bottom, left] int array
//	border - border style: none, rounded, bold, hidden, thick, double, inner
//	border_color - border foreground color
//	border_background - border background color
//	border_top / border_left / border_right / border_bottom - show specific border sides
//	width - fixed width
//	height - fixed height
//	align - text alignment: left, center, right, top, bottom
//	inline - render inline (no newlines)
//	max_width - maximum width
//	max_height - maximum height
func (s *Style) SetProperty(name string, value any) error {
	if colorProps[name] {
		value = s.resolveColorValue(value)
	}
	handler, ok := propertyHandlers[name]
	if !ok {
		return fmt.Errorf("unknown style property %q", name)
	}
	s.defined = true
	return handler(&s.s, value)
}

// IsDefined returns true if any property was set on this style.
func (s Style) IsDefined() bool {
	return s.defined
}

// Merge applies other on top of s — other's properties win where set.
// This is used to layer component overrides on top of the default style.
func (s Style) Merge(other Style) Style {
	return Style{
		s:       other.s.Inherit(s.s),
		palette: s.palette,
		defined: s.defined || other.defined,
	}
}

// Copy returns a deep copy of the Style.
func (s Style) Copy() Style {
	return Style{
		s:       s.s.Copy(),
		palette: s.palette,
		defined: s.defined,
	}
}

// propertyHandlers maps property names to functions that apply them to a lipgloss.Style.
var propertyHandlers = map[string]func(s *lipgloss.Style, v any) error{
	// Colors
	"color":             setForeground,
	"foreground":        setForeground,
	"background":        setBackground,
	"border_color":      setBorderColor,
	"border_background": setBorderBackground,

	// Text formatting
	"bold":          setBool("Bold"),
	"italic":        setBool("Italic"),
	"underline":     setUnderline,
	"strikethrough": setBool("Strikethrough"),
	"reverse":       setBool("Reverse"),
	"blink":         setBool("Blink"),
	"dim":           setBool("Dim"),
	"faint":         setBool("Faint"),

	// Spacing
	"padding": setPadding,
	"margin":  setMargin,

	// Border
	"border":        setBorder,
	"border_top":    setBorderTop,
	"border_left":   setBorderLeft,
	"border_right":  setBorderRight,
	"border_bottom": setBorderBottom,

	// Dimensions
	"width":      setInt("Width"),
	"height":     setInt("Height"),
	"max_width":  setInt("MaxWidth"),
	"max_height": setInt("MaxHeight"),

	// Alignment
	"align": setAlign,

	// Other
	"inline":    setBool("Inline"),
	"tab_width": setInt("TabWidth"),
}

// resolveColor converts a value to lipgloss.Color.
// It handles color names (from palette), hex strings, and ANSI indices.
func resolveColor(palette *ColorPalette, v any) ansi.Color {
	switch val := v.(type) {
	case string:
		// Check if it's a named color in the palette
		if palette != nil && palette.Has(val) {
			return palette.Get(val)
		}
		// Try as hex or ANSI index
		return lipgloss.Color(val)
	case ansi.Color:
		return val
	case lipgloss.NoColor:
		return val
	default:
		// Try to convert to string
		return lipgloss.Color(fmt.Sprintf("%v", v))
	}
}

func setForeground(s *lipgloss.Style, v any) error {
	*s = s.Foreground(resolveColor(nil, v))
	return nil
}

func setBackground(s *lipgloss.Style, v any) error {
	*s = s.Background(resolveColor(nil, v))
	return nil
}

func setBorderColor(s *lipgloss.Style, v any) error {
	*s = s.BorderForeground(resolveColor(nil, v))
	return nil
}

func setBorderBackground(s *lipgloss.Style, v any) error {
	*s = s.BorderBackground(resolveColor(nil, v))
	return nil
}

func setBool(methodName string) func(s *lipgloss.Style, v any) error {
	return func(s *lipgloss.Style, v any) error {
		var val bool
		switch v := v.(type) {
		case bool:
			val = v
		case string:
			switch strings.ToLower(v) {
			case "true", "1", "yes", "on":
				val = true
			case "false", "0", "no", "off":
				val = false
			default:
				b, err := strconv.ParseBool(v)
				if err != nil {
					return fmt.Errorf("cannot parse %q as boolean for %s", v, methodName)
				}
				val = b
			}
		default:
			return fmt.Errorf("cannot convert %T to bool for %s", v, methodName)
		}
		switch methodName {
		case "Bold":
			*s = s.Bold(val)
		case "Italic":
			*s = s.Italic(val)
		case "Strikethrough":
			*s = s.Strikethrough(val)
		case "Reverse":
			*s = s.Reverse(val)
		case "Blink":
			*s = s.Blink(val)
		case "Dim":
			*s = s.Faint(val)
		case "Faint":
			*s = s.Faint(val)
		case "Inline":
			*s = s.Inline(val)
		default:
			return fmt.Errorf("unknown bool property %q", methodName)
		}
		return nil
	}
}

func setUnderline(s *lipgloss.Style, v any) error {
	var val bool
	switch v := v.(type) {
	case bool:
		// no-op
	case string:
		switch strings.ToLower(v) {
		case "true", "1", "yes", "on":
			val = true
		case "false", "0", "no", "off":
			val = false
		default:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("cannot parse %q as boolean for underline", v)
			}
			val = b
		}
	default:
		return fmt.Errorf("cannot convert %T to bool for underline", v)
	}
	if val {
		*s = s.Underline(true)
	} else {
		*s = s.Underline(false)
	}
	return nil
}

func setPadding(s *lipgloss.Style, v any) error {
	pad, err := parseIntSlice(v, 4)
	if err != nil {
		return fmt.Errorf("padding: %w", err)
	}
	switch len(pad) {
	case 4:
		*s = s.Padding(pad[0], pad[1], pad[2], pad[3])
	case 3:
		*s = s.Padding(pad[0], pad[1], pad[2])
	case 2:
		*s = s.Padding(pad[0], pad[1])
	case 1:
		*s = s.Padding(pad[0])
	default:
		return fmt.Errorf("padding must be 1-4 integers, got %d", len(pad))
	}
	return nil
}

func setMargin(s *lipgloss.Style, v any) error {
	mar, err := parseIntSlice(v, 4)
	if err != nil {
		return fmt.Errorf("margin: %w", err)
	}
	switch len(mar) {
	case 4:
		*s = s.Margin(mar[0], mar[1], mar[2], mar[3])
	case 3:
		*s = s.Margin(mar[0], mar[1], mar[2])
	case 2:
		*s = s.Margin(mar[0], mar[1])
	case 1:
		*s = s.Margin(mar[0])
	default:
		return fmt.Errorf("margin must be 1-4 integers, got %d", len(mar))
	}
	return nil
}

func setBorder(s *lipgloss.Style, v any) error {
	borderStr, ok := v.(string)
	if !ok {
		return fmt.Errorf("border must be a string, got %T", v)
	}
	borderStr = strings.ToLower(strings.TrimSpace(borderStr))
	var border lipgloss.Border
	switch borderStr {
	case "none", "null", "":
		border = lipgloss.HiddenBorder()
	case "rounded":
		border = lipgloss.RoundedBorder()
	case "bold", "thick":
		border = lipgloss.ThickBorder()
	case "hidden":
		border = lipgloss.HiddenBorder()
	case "double":
		border = lipgloss.DoubleBorder()
	case "inner":
		border = lipgloss.InnerHalfBlockBorder()
	case "block":
		border = lipgloss.BlockBorder()
	case "ascii":
		border = lipgloss.ASCIIBorder()
	case "markdown":
		border = lipgloss.MarkdownBorder()
	default:
		return fmt.Errorf("unknown border style %q (use none, rounded, bold, thick, hidden, double, inner, block, ascii, markdown)", borderStr)
	}
	*s = s.BorderStyle(border)
	return nil
}

func setBorderTop(s *lipgloss.Style, v any) error {
	return setBorderSide(s, v, func(b bool) { *s = s.BorderTop(b) })
}

func setBorderLeft(s *lipgloss.Style, v any) error {
	return setBorderSide(s, v, func(b bool) { *s = s.BorderLeft(b) })
}

func setBorderRight(s *lipgloss.Style, v any) error {
	return setBorderSide(s, v, func(b bool) { *s = s.BorderRight(b) })
}

func setBorderBottom(s *lipgloss.Style, v any) error {
	return setBorderSide(s, v, func(b bool) { *s = s.BorderBottom(b) })
}

func setBorderSide(s *lipgloss.Style, v any, setter func(bool)) error {
	var val bool
	switch v := v.(type) {
	case bool:
		// no-op
	case string:
		switch strings.ToLower(v) {
		case "true", "1", "yes", "on":
			val = true
		case "false", "0", "no", "off":
			val = false
		default:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("cannot parse %q as boolean for border side", v)
			}
			val = b
		}
	default:
		return fmt.Errorf("cannot convert %T to bool for border side", v)
	}
	setter(val)
	return nil
}

func setInt(methodName string) func(s *lipgloss.Style, v any) error {
	return func(s *lipgloss.Style, v any) error {
		val, err := toInt(v)
		if err != nil {
			return fmt.Errorf("%s: %w", methodName, err)
		}
		switch methodName {
		case "Width":
			*s = s.Width(val)
		case "Height":
			*s = s.Height(val)
		case "MaxWidth":
			*s = s.MaxWidth(val)
		case "MaxHeight":
			*s = s.MaxHeight(val)
		case "TabWidth":
			*s = s.TabWidth(val)
		default:
			return fmt.Errorf("unknown int property %q", methodName)
		}
		return nil
	}
}

func setAlign(s *lipgloss.Style, v any) error {
	alignStr, ok := v.(string)
	if !ok {
		return fmt.Errorf("align must be a string, got %T", v)
	}
	alignStr = strings.ToLower(strings.TrimSpace(alignStr))
	switch alignStr {
	case "left", "start":
		*s = s.Align(lipgloss.Left)
	case "center", "middle":
		*s = s.Align(lipgloss.Center)
	case "right", "end":
		*s = s.Align(lipgloss.Right)
	case "top":
		*s = s.AlignVertical(lipgloss.Top)
		*s = s.AlignHorizontal(lipgloss.Left)
	case "bottom":
		*s = s.AlignVertical(lipgloss.Bottom)
		*s = s.AlignHorizontal(lipgloss.Left)
	default:
		return fmt.Errorf("unknown alignment %q (use left, center, right, top, bottom)", alignStr)
	}
	return nil
}

// toInt converts a value to int.
func toInt(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int8:
		return int(val), nil
	case int16:
		return int(val), nil
	case int32:
		return int(val), nil
	case int64:
		return int(val), nil
	case uint:
		return int(val), nil
	case uint8:
		return int(val), nil
	case uint16:
		return int(val), nil
	case uint32:
		return int(val), nil
	case uint64:
		return int(val), nil
	case float32:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as integer", val)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

// parseIntSlice converts a value to a slice of ints.
// Supports YAML arrays (converted to []any) and comma-separated strings.
func parseIntSlice(v any, expected int) ([]int, error) {
	// Handle comma-separated string
	if str, ok := v.(string); ok {
		parts := strings.Split(str, ",")
		result := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			i, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("cannot parse %q as integer in slice", p)
			}
			result = append(result, i)
		}
		return result, nil
	}

	// Handle slice
	switch slice := v.(type) {
	case []int:
		return slice, nil
	case []any:
		result := make([]int, 0, len(slice))
		for _, item := range slice {
			i, err := toInt(item)
			if err != nil {
				return nil, fmt.Errorf("cannot convert element to int: %w", err)
			}
			result = append(result, i)
		}
		return result, nil
	}

	return nil, fmt.Errorf("value must be a slice or comma-separated string, got %T", v)
}
