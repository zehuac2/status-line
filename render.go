package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

func rateStyle(remaining int) lipgloss.Style {
	switch {
	case remaining >= 50:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	case remaining >= 20:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	}
}

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

func render(in StatusInput) string {
	cyan := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	blue := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	red := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	green := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	magenta := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	gray := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8"))

	dir := filepath.Base(in.Cwd)
	if dir == "" || dir == "." {
		dir = in.Cwd
	}

	var gitSeg, modelSeg, ctxSeg string

	gitSeg = cyan.Render(dir)
	if in.Cwd != "" {
		if branch, ok := getGitBranch(in.Cwd); ok {
			gitSeg = row(gitSeg, blue.Render("git:(")+red.Render(branch)+blue.Render(")"))
		}
	}

	if name := in.Model.DisplayName; name != "" {
		modelSeg = green.Render("✦ " + name)
	}

	if p := in.ContextWindow.UsedPercentage; p != nil {
		ctxSeg = magenta.Render("ctx ") + bar(*p, 10, magenta)
	}

	line1 := row(gitSeg, modelSeg, ctxSeg)

	var diffSeg, sessionSeg string

	if in.Cost.TotalLinesAdded != nil && in.Cost.TotalLinesRemoved != nil {
		diffSeg = row(
			green.Render(fmt.Sprintf("▲%d", *in.Cost.TotalLinesAdded)),
			red.Render(fmt.Sprintf("▼%d", *in.Cost.TotalLinesRemoved)),
		)
	}

	if in.Cost.TotalDurationMs != nil {
		d := time.Duration(*in.Cost.TotalDurationMs) * time.Millisecond
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		sessionSeg = gray.Render(fmt.Sprintf("⧗ %dh%02dm", h, m))
	}

	line2 := row(diffSeg, sessionSeg)

	var costSeg, fiveHrSeg, sevenDSeg, resetSeg string

	if in.Cost.TotalCostUSD != nil {
		costSeg = green.Render(fmt.Sprintf("$%.2f", *in.Cost.TotalCostUSD))
	}

	if p := in.RateLimits.FiveHour.UsedPercentage; p != nil {
		rem := int(math.Round(100 - *p))
		style := rateStyle(rem)
		fiveHrSeg = style.Render("5h ") + bar(*p, 10, style)
	}

	if p := in.RateLimits.SevenDay.UsedPercentage; p != nil {
		rem := int(math.Round(100 - *p))
		style := rateStyle(rem)
		sevenDSeg = style.Render("7d ") + bar(*p, 10, style)
	}

	resetsAt := in.RateLimits.FiveHour.ResetsAt
	if resetsAt == nil {
		resetsAt = in.RateLimits.SevenDay.ResetsAt
	}
	if resetsAt != nil {
		resetSeg = gray.Render("↺ " + time.Unix(*resetsAt, 0).Format("3:04pm"))
	}

	line3 := row(costSeg, fiveHrSeg, sevenDSeg, resetSeg)

	var lines []string
	for _, l := range []string{line1, line2, line3} {
		if l != "" {
			lines = append(lines, l)
		}
	}

	return strings.Join(lines, "\n")
}
