package main

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// theme centralizes every color used by render() so the palette lives in one
// place and can be swapped by constructing a different *theme.
type theme struct {
	WarmGray color.Color // cwd basename, git brackets, ✦, ctx bar, ▲▼, $cost, ↺
	DimGray  color.Color // session duration, 7d bar, mode label
	Primary  color.Color // branch, model, 5h bar, reset time
	Divider  color.Color // ─ rule under the mode row

	Vim vimTheme // per-mode vim accents
}

// vimTheme holds the accent color for each vim.mode value.
type vimTheme struct {
	Normal     color.Color
	Insert     color.Color
	Visual     color.Color
	VisualLine color.Color
	Replace    color.Color
}

// color resolves a vim.mode string to its accent, falling back to Normal
// (coral) for NORMAL and any unrecognized mode.
func (v vimTheme) color(mode string) color.Color {
	switch mode {
	case "INSERT":
		return v.Insert
	case "VISUAL":
		return v.Visual
	case "VISUAL LINE":
		return v.VisualLine
	case "REPLACE":
		return v.Replace
	default:
		return v.Normal
	}
}

func claudeTheme() *theme {
	return &theme{
		WarmGray: lipgloss.Color("#8f8a80"),
		DimGray:  lipgloss.Color("#6f6b62"),
		Primary:  lipgloss.Color("#d97757"),
		Divider:  lipgloss.Color("#2a2a2a"),
		Vim: vimTheme{
			Normal:     lipgloss.Color("#d97757"),
			Insert:     lipgloss.Color("#69c27e"),
			Visual:     lipgloss.Color("#9792ec"),
			VisualLine: lipgloss.Color("#9792ec"),
			Replace:    lipgloss.Color("#e36b65"),
		},
	}
}
