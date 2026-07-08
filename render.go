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
	primary := lipgloss.NewStyle().Bold(true).Foreground(t.Primary)
	warmGray := lipgloss.NewStyle().Bold(true).Foreground(t.WarmGray)
	warmGrayNormal := lipgloss.NewStyle().Foreground(t.WarmGray)
	dimGray := lipgloss.NewStyle().Bold(true).Foreground(t.DimGray)
	dimGrayNormal := lipgloss.NewStyle().Foreground(t.DimGray)
	divider := lipgloss.NewStyle().Foreground(t.Divider)

	dir := filepath.Base(in.Cwd)
	if dir == "" || dir == "." {
		dir = in.Cwd
	}

	var dirSeg, gitSeg, modelSeg, ctxSeg string

	if in.Cwd != "" {
		dirSeg = warmGrayNormal.Render(dir)
		if branch, ok := getGitBranch(in.Cwd); ok {
			gitSeg = warmGray.Render("git:(") + primary.Render(branch) + warmGray.Render(")")
		}
	}

	if name := in.Model.DisplayName; name != "" {
		modelSeg = warmGray.Render("✦ ") + primary.Render(name)
	}

	if p := in.ContextWindow.UsedPercentage; p != nil {
		ctxSeg = warmGrayNormal.Render("ctx ") + components.Bar(*p, 10, warmGrayNormal)
	}

	line1 := components.Row(dirSeg, gitSeg, modelSeg, ctxSeg)

	var diffSeg, sessionSeg string

	if in.Cost.TotalLinesAdded != nil && in.Cost.TotalLinesRemoved != nil {
		diffSeg = warmGray.Render(fmt.Sprintf("▲%d ▼%d", *in.Cost.TotalLinesAdded, *in.Cost.TotalLinesRemoved))
	}

	if in.Cost.TotalDurationMs != nil {
		d := time.Duration(*in.Cost.TotalDurationMs) * time.Millisecond
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		sessionSeg = dimGray.Render(fmt.Sprintf("⧗ %dh%02dm", h, m))
	}

	line2 := components.Row(diffSeg, sessionSeg)

	var costSeg, fiveHrSeg, sevenDSeg, resetSeg string

	if in.Cost.TotalCostUSD != nil {
		costSeg = warmGray.Render(fmt.Sprintf("$%.2f", *in.Cost.TotalCostUSD))
	}

	if p := in.RateLimits.FiveHour.UsedPercentage; p != nil {
		fiveHrSeg = primary.Render("5h ") + components.Bar(*p, 10, primary)
	}

	if p := in.RateLimits.SevenDay.UsedPercentage; p != nil {
		sevenDSeg = dimGray.Render("7d ") + components.Bar(*p, 10, dimGray)
	}

	resetsAt := in.RateLimits.FiveHour.ResetsAt
	if resetsAt == nil {
		resetsAt = in.RateLimits.SevenDay.ResetsAt
	}
	if resetsAt != nil {
		resetSeg = warmGray.Render("↺ ") + primary.Render(time.Unix(*resetsAt, 0).Format("3:04pm"))
	}

	line3 := components.Row(costSeg, fiveHrSeg, sevenDSeg, resetSeg)

	var modeRow, dividerRow string
	if mode := in.Vim.Mode; mode != "" {
		modeColor := lipgloss.NewStyle().Bold(true).Foreground(t.Vim.color(mode))
		modeRow = components.Row(dimGrayNormal.Render("mode"), modeColor.Render(mode))

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
