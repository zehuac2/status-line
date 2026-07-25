package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/zehuac2/status-line/components"
)

// percentage rounds a percentage value to the nearest whole number and
// returns it as a string with a trailing percent sign.
func percentage(p float64) string {
	return fmt.Sprintf("%d%%", int(math.Round(p)))
}

func render(in StatusInput, t *theme) string {
	textDimNormal := lipgloss.NewStyle().Foreground(t.TextDim)
	divider := lipgloss.NewStyle().Foreground(t.Divider)

	identityRow := renderIdentityRow(in, t)
	usageRow := renderUsageRow(in, t)
	activityRow := renderActivityRow(in, t)

	var modeRow, dividerRow string
	if mode := in.Vim.Mode; mode != "" {
		modeColor := lipgloss.NewStyle().Bold(true).Foreground(in.Vim.color(&t.Vim))
		modeRow = components.Row(textDimNormal.Render("mode"), modeColor.Render(mode))
	}

	w := lipgloss.Width(modeRow)
	for _, l := range []string{identityRow, usageRow, activityRow} {
		if lw := lipgloss.Width(l); lw > w {
			w = lw
		}
	}

	if modeRow != "" {
		dividerRow = divider.Render(strings.Repeat("─", w))
	}

	return components.Box(modeRow, dividerRow, identityRow, usageRow, activityRow)
}

// renderIdentityRow renders line 1: cwd basename, git branch, model name, and
// context-window percentage.
func renderIdentityRow(in StatusInput, t *theme) string {
	primary := lipgloss.NewStyle().Bold(true).Foreground(t.Primary)
	text := lipgloss.NewStyle().Bold(true).Foreground(t.Text)
	textNormal := lipgloss.NewStyle().Foreground(t.Text)

	dir := filepath.Base(in.Cwd)
	if dir == "" || dir == "." {
		dir = in.Cwd
	}

	var dirSeg, gitSeg, modelSeg, ctxSeg string

	if in.Cwd != "" {
		dirSeg = textNormal.Render(dir)
		branch := in.Branch
		if branch == "" {
			if b, ok := getGitBranch(in.Cwd); ok {
				branch = b
			}
		}
		if branch != "" {
			gitSeg = text.Render("(") + primary.Render(branch) + text.Render(")")
		}
	}

	if name := in.Model.DisplayName; name != "" {
		modelSeg = text.Render("✦ ") + primary.Render(name)
	}

	if p := in.ContextWindow.UsedPercentage; p != nil {
		ctxSeg = textNormal.Render("ctx " + percentage(*p))
	}

	return components.Row(dirSeg, gitSeg, modelSeg, ctxSeg)
}

// renderUsageRow renders line 2: session cost, model effort, 5h/7d
// rate-limit usage, and the next rate-limit reset time.
func renderUsageRow(in StatusInput, t *theme) string {
	primary := lipgloss.NewStyle().Bold(true).Foreground(t.Primary)
	text := lipgloss.NewStyle().Bold(true).Foreground(t.Text)
	textDim := lipgloss.NewStyle().Bold(true).Foreground(t.TextDim)

	var costSeg, effortSeg, fiveHrSeg, sevenDSeg, resetSeg string

	if in.Cost.TotalCostUSD != nil {
		costSeg = text.Render(fmt.Sprintf("$%.2f", *in.Cost.TotalCostUSD))
	}

	if effort := in.Effort.Level; effort != "" {
		effortSeg = primary.Render(effort)
	}

	if p := in.RateLimits.FiveHour.UsedPercentage; p != nil {
		fiveHrSeg = primary.Render("5h " + percentage(*p))
	}

	if p := in.RateLimits.SevenDay.UsedPercentage; p != nil {
		sevenDSeg = textDim.Render("7d " + percentage(*p))
	}

	resetsAt := in.RateLimits.FiveHour.ResetsAt
	if resetsAt == nil {
		resetsAt = in.RateLimits.SevenDay.ResetsAt
	}
	if resetsAt != nil {
		resetSeg = text.Render("↺ ") + primary.Render(time.Unix(*resetsAt, 0).Format("3:04pm"))
	}

	return components.Row(costSeg, effortSeg, fiveHrSeg, sevenDSeg, resetSeg)
}

// renderActivityRow renders line 3: lines added/removed and session
// duration.
func renderActivityRow(in StatusInput, t *theme) string {
	text := lipgloss.NewStyle().Bold(true).Foreground(t.Text)
	textDim := lipgloss.NewStyle().Bold(true).Foreground(t.TextDim)

	var diffSeg, sessionSeg string

	if in.Cost.TotalLinesAdded != nil && in.Cost.TotalLinesRemoved != nil {
		diffSeg = text.Render(fmt.Sprintf("▲%d ▼%d", *in.Cost.TotalLinesAdded, *in.Cost.TotalLinesRemoved))
	}

	if in.Cost.TotalDurationMs != nil {
		d := time.Duration(*in.Cost.TotalDurationMs) * time.Millisecond
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		sessionSeg = textDim.Render(fmt.Sprintf("⧗ %dh%02dm", h, m))
	}

	return components.Row(diffSeg, sessionSeg)
}
