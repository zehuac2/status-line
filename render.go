package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/zehuac2/status-line/components"
)

func render(in StatusInput, t *theme) string {
	accent := lipgloss.NewStyle().Bold(true).Foreground(t.Coral)
	label := lipgloss.NewStyle().Bold(true).Foreground(t.WarmGray)
	labelNorm := lipgloss.NewStyle().Foreground(t.WarmGray)
	dim := lipgloss.NewStyle().Bold(true).Foreground(t.DimGray)
	dimNorm := lipgloss.NewStyle().Foreground(t.DimGray)
	divider := lipgloss.NewStyle().Foreground(t.Divider)

	dir := filepath.Base(in.Cwd)
	if dir == "" || dir == "." {
		dir = in.Cwd
	}

	var dirSeg, gitSeg, modelSeg, ctxSeg string

	if in.Cwd != "" {
		dirSeg = labelNorm.Render(dir)
		if branch, ok := getGitBranch(in.Cwd); ok {
			gitSeg = label.Render("git:(") + accent.Render(branch) + label.Render(")")
		}
	}

	if name := in.Model.DisplayName; name != "" {
		modelSeg = label.Render("✦ ") + accent.Render(name)
	}

	if p := in.ContextWindow.UsedPercentage; p != nil {
		ctxSeg = labelNorm.Render("ctx ") + components.Bar(*p, 10, labelNorm)
	}

	line1 := components.Row(dirSeg, gitSeg, modelSeg, ctxSeg)

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

	line2 := components.Row(diffSeg, sessionSeg)

	var costSeg, fiveHrSeg, sevenDSeg, resetSeg string

	if in.Cost.TotalCostUSD != nil {
		costSeg = label.Render(fmt.Sprintf("$%.2f", *in.Cost.TotalCostUSD))
	}

	if p := in.RateLimits.FiveHour.UsedPercentage; p != nil {
		fiveHrSeg = accent.Render("5h ") + components.Bar(*p, 10, accent)
	}

	if p := in.RateLimits.SevenDay.UsedPercentage; p != nil {
		sevenDSeg = dim.Render("7d ") + components.Bar(*p, 10, dim)
	}

	resetsAt := in.RateLimits.FiveHour.ResetsAt
	if resetsAt == nil {
		resetsAt = in.RateLimits.SevenDay.ResetsAt
	}
	if resetsAt != nil {
		resetSeg = label.Render("↺ ") + accent.Render(time.Unix(*resetsAt, 0).Format("3:04pm"))
	}

	line3 := components.Row(costSeg, fiveHrSeg, sevenDSeg, resetSeg)

	var modeRow, dividerRow string
	if mode := in.Vim.Mode; mode != "" {
		modeColor := lipgloss.NewStyle().Bold(true).Foreground(t.Vim.color(mode))
		modeRow = components.Row(dimNorm.Render("mode"), modeColor.Render(mode))

		w := lipgloss.Width(modeRow)
		for _, l := range []string{line1, line2, line3} {
			if lw := lipgloss.Width(l); lw > w {
				w = lw
			}
		}
		dividerRow = divider.Render(strings.Repeat("─", w))
	}

	return components.Box(modeRow, dividerRow, line1, line2, line3)
}
