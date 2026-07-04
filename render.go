package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// bar renders pct (0-100) as a length-cell block gauge in the given style.
// A cell that's only partially filled renders as a solid block in a faint
// variant of style rather than a fractional glyph — the eighth-block runes
// (▏▎▍▌▋▊▉) are inconsistently metriced across monospace fonts and can show
// a hairline gap next to the neighboring cells.
func bar(pct float64, length int, style lipgloss.Style) string {
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

// row joins non-empty segments with a single space using lipgloss.JoinHorizontal.
func row(segments ...string) string {
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

// box frames non-empty lines with corner brackets only (no connecting edges),
// left-padded to line up under the top-left corner.
func box(lines []string) string {
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

func render(in StatusInput) string {
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#d97757"))
	label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8f8a80"))
	labelNorm := lipgloss.NewStyle().Foreground(lipgloss.Color("#8f8a80"))
	dim := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6f6b62"))

	dir := filepath.Base(in.Cwd)
	if dir == "" || dir == "." {
		dir = in.Cwd
	}

	var gitSeg, modelSeg, ctxSeg string

	if in.Cwd != "" {
		gitSeg = labelNorm.Render(dir)
		if branch, ok := getGitBranch(in.Cwd); ok {
			gitSeg = row(gitSeg, label.Render("git:(")+accent.Render(branch)+label.Render(")"))
		}
	}

	if name := in.Model.DisplayName; name != "" {
		modelSeg = label.Render("✦ ") + accent.Render(name)
	}

	if p := in.ContextWindow.UsedPercentage; p != nil {
		ctxSeg = labelNorm.Render("ctx ") + bar(*p, 10, labelNorm)
	}

	line1 := row(gitSeg, modelSeg, ctxSeg)

	var diffSeg, sessionSeg string

	if in.Cost.TotalLinesAdded != nil && in.Cost.TotalLinesRemoved != nil {
		diffSeg = label.Render(fmt.Sprintf("▲%d ▼%d", *in.Cost.TotalLinesAdded, *in.Cost.TotalLinesRemoved))
	}

	if in.Cost.TotalDurationMs != nil {
		d := time.Duration(*in.Cost.TotalDurationMs) * time.Millisecond
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		sessionSeg = dim.Render(fmt.Sprintf("⧗ %dh%02dm", h, m))
	}

	line2 := row(diffSeg, sessionSeg)

	var costSeg, fiveHrSeg, sevenDSeg, resetSeg string

	if in.Cost.TotalCostUSD != nil {
		costSeg = label.Render(fmt.Sprintf("$%.2f", *in.Cost.TotalCostUSD))
	}

	if p := in.RateLimits.FiveHour.UsedPercentage; p != nil {
		fiveHrSeg = accent.Render("5h ") + bar(*p, 10, accent)
	}

	if p := in.RateLimits.SevenDay.UsedPercentage; p != nil {
		sevenDSeg = dim.Render("7d ") + bar(*p, 10, dim)
	}

	resetsAt := in.RateLimits.FiveHour.ResetsAt
	if resetsAt == nil {
		resetsAt = in.RateLimits.SevenDay.ResetsAt
	}
	if resetsAt != nil {
		resetSeg = label.Render("↺ ") + accent.Render(time.Unix(*resetsAt, 0).Format("3:04pm"))
	}

	line3 := row(costSeg, fiveHrSeg, sevenDSeg, resetSeg)

	return box([]string{line1, line2, line3})
}
