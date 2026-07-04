package components

import "charm.land/lipgloss/v2"

// box frames non-empty lines with corner brackets only (no connecting edges),
// left-padded to line up under the top-left corner.
func Box(lines ...string) string {
	border := lipgloss.Border{
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		// U+2800 (blank braille pattern) renders as a blank cell but, unlike
		// a literal space, isn't whitespace — Claude Code's status line
		// strips leading whitespace per line, which collapses a real " "
		// left border and misaligns content under the top-left corner.
		Left:   "\u2800",
		Top:    "\u2800",
		Right:  "\u2800",
		Bottom: "\u2800",
	}

	borderStyle := lipgloss.NewStyle().
		Border(border).
		PaddingLeft(1).
		PaddingRight(1)

	return borderStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
