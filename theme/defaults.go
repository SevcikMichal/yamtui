package theme

// DefaultTheme returns the default minimal theme.
func DefaultTheme() *Theme {
	p := newDefaultPalette()
	def := NewStyle()
	def.SetPalette(p)
	def.SetProperty("color", "text")
	def.SetProperty("background", "background")

	foc := NewStyle()
	foc.SetPalette(p)
	foc.SetProperty("border", "rounded")
	foc.SetProperty("border_color", "accent")

	return &Theme{
		Name:    "default",
		Colors:  p,
		Default: def,
		Focused: foc,
	}
}

// CatppuccinTheme returns the Catppuccin Mocha theme.
func CatppuccinTheme() *Theme {
	p := newCatppuccinPalette()
	def := NewStyle()
	def.SetPalette(p)
	def.SetProperty("color", "text")
	def.SetProperty("background", "background")

	foc := NewStyle()
	foc.SetPalette(p)
	foc.SetProperty("border", "rounded")
	foc.SetProperty("border_color", "blue")
	foc.SetProperty("color", "lavender")

	return &Theme{
		Name:    "catppuccin",
		Colors:  p,
		Default: def,
		Focused: foc,
	}
}

// DraculaTheme returns the Dracula theme.
func DraculaTheme() *Theme {
	p := newDraculaPalette()
	def := NewStyle()
	def.SetPalette(p)
	def.SetProperty("color", "text")
	def.SetProperty("background", "background")

	foc := NewStyle()
	foc.SetPalette(p)
	foc.SetProperty("border", "rounded")
	foc.SetProperty("border_color", "purple")

	return &Theme{
		Name:    "dracula",
		Colors:  p,
		Default: def,
		Focused: foc,
	}
}

// NordTheme returns the Nord theme.
func NordTheme() *Theme {
	p := newNordPalette()
	def := NewStyle()
	def.SetPalette(p)
	def.SetProperty("color", "text")
	def.SetProperty("background", "background")

	foc := NewStyle()
	foc.SetPalette(p)
	foc.SetProperty("border", "rounded")
	foc.SetProperty("border_color", "cyan")

	return &Theme{
		Name:    "nord",
		Colors:  p,
		Default: def,
		Focused: foc,
	}
}

// --- Color Palettes ---

func newDefaultPalette() *ColorPalette {
	p := NewColorPalette()
	p.Set("text", "#FFFFFF")
	p.Set("background", "#1A1B26")
	p.Set("muted", "#5C6166")
	p.Set("accent", "#007ACC")
	return p
}

func newCatppuccinPalette() *ColorPalette {
	p := NewColorPalette()
	p.Set("text", "#CDD6F4")
	p.Set("background", "#1E1E2E")
	p.Set("surface", "#313244")
	p.Set("muted", "#6C7086")
	p.Set("overlay", "#737992")
	p.Set("subtle", "#9399B2")
	p.Set("blue", "#89B4FA")
	p.Set("lavender", "#B4BEFE")
	p.Set("sapphire", "#74C7EC")
	p.Set("sky", "#89DCEB")
	p.Set("teal", "#94E2D5")
	p.Set("green", "#A6E3A1")
	p.Set("yellow", "#F9E2AF")
	p.Set("peach", "#FAB387")
	p.Set("maroon", "#F38BA8")
	p.Set("red", "#F38BA8")
	p.Set("mauve", "#BA5D83")
	p.Set("pink", "#F5C2E7")
	p.Set("flamingo", "#F2D5CF")
	p.Set("rosewater", "#F5E0DC")
	return p
}

func newDraculaPalette() *ColorPalette {
	p := NewColorPalette()
	p.Set("text", "#F8F8F2")
	p.Set("background", "#282A36")
	p.Set("current_line", "#44475A")
	p.Set("selection", "#44475A")
	p.Set("comment", "#6272A4")
	p.Set("cyan", "#8BE9FD")
	p.Set("green", "#50FA7B")
	p.Set("orange", "#FFB86C")
	p.Set("pink", "#FF79C6")
	p.Set("purple", "#BD93F9")
	p.Set("red", "#FF5555")
	p.Set("yellow", "#F1FA8C")
	return p
}

func newNordPalette() *ColorPalette {
	p := NewColorPalette()
	p.Set("text", "#ECEFF4")
	p.Set("background", "#2E3440")
	p.Set("comment", "#4C566A")
	p.Set("cyan", "#88C0D0")
	p.Set("dark_cyan", "#81A1C1")
	p.Set("green", "#A3BE8C")
	p.Set("orange", "#D08770")
	p.Set("pink", "#B48EAD")
	p.Set("purple", "#B48EAD")
	p.Set("red", "#BF616A")
	p.Set("yellow", "#EBCB8B")
	return p
}
