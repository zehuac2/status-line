package components

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// bar renders pct (0-100) as a length-cell block gauge in the given style.
// A cell that's only partially filled renders as a solid block in a faint
// variant of style rather than a fractional glyph — the eighth-block runes
// (▏▎▍▌▋▊▉) are inconsistently metriced across monospace fonts and can show
// a hairline gap next to the neighboring cells.
func Bar(pct float64, length int, style lipgloss.Style) string {
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}

	p := pct / 100 * float64(length)
	eighths := int(math.Round(p * 8))
	full := eighths / 8
	partial := eighths%8 != 0
	if full >= length {
		full = length
		partial = false
	}
	empty := length - full
	if partial {
		empty--
	}

	var s strings.Builder
	if full > 0 {
		s.WriteString(style.Render(strings.Repeat("█", full)))
	}
	if partial {
		s.WriteString(style.Faint(true).Render("█"))
	}
	if empty > 0 {
		s.WriteString(style.Render(strings.Repeat("░", empty)))
	}
	return s.String()
}
