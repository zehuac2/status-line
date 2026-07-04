package components

import "charm.land/lipgloss/v2"

// row joins non-empty segments with a single space using lipgloss.JoinHorizontal.
func Row(segments ...string) string {
	var present []string
	for _, s := range segments {
		if s != "" {
			present = append(present, s)
		}
	}
	if len(present) == 0 {
		return ""
	}

	parts := make([]string, 0, len(present)*2-1)
	for i, s := range present {
		if i > 0 {
			parts = append(parts, " ")
		}
		parts = append(parts, s)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
