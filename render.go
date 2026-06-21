package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

func rateStyle(remaining int) lipgloss.Style {
	switch {
	case remaining >= 50:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	case remaining >= 20:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	}
}

func render(in StatusInput) string {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	boldBlue := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	magenta := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	dir := filepath.Base(in.Cwd)
	if dir == "" || dir == "." {
		dir = in.Cwd
	}

	var line1 strings.Builder
	line1.WriteString(cyan.Render(dir))

	if in.Cwd != "" {
		if gs, ok := getGitStatus(in.Cwd); ok {
			gitPart := boldBlue.Render("git:(") + red.Render(gs.branch) + boldBlue.Render(")")
			if gs.dirty {
				gitPart += " " + yellow.Render("✗")
			}
			line1.WriteString(" " + gitPart)
		}
	}

	if p := in.ContextWindow.UsedPercentage; p != nil {
		used := int(math.Round(*p))
		line1.WriteString(" " + magenta.Render(fmt.Sprintf("ctx:%d%%", used)))
	}

	if in.ContextWindow.TotalInputTokens != nil && in.ContextWindow.TotalOutputTokens != nil {
		total := *in.ContextWindow.TotalInputTokens + *in.ContextWindow.TotalOutputTokens
		var display string
		if total >= 1000 {
			display = fmt.Sprintf("%.1fk", float64(total)/1000)
		} else {
			display = fmt.Sprintf("%d", total)
		}
		line1.WriteString(" " + yellow.Render("tok:"+display))
	}

	var line2Parts []string

	if name := in.Model.DisplayName; name != "" {
		line2Parts = append(line2Parts, cyan.Render(name))
	}

	if in.Cost.TotalCostUSD != nil {
		line2Parts = append(line2Parts, green.Render(fmt.Sprintf("$%.2f", *in.Cost.TotalCostUSD)))
	}

	var rateParts []string
	if p := in.RateLimits.FiveHour.UsedPercentage; p != nil {
		rem := int(math.Round(100 - *p))
		rateParts = append(rateParts, rateStyle(rem).Render(fmt.Sprintf("5h:%d%%", rem)))
	}
	if p := in.RateLimits.SevenDay.UsedPercentage; p != nil {
		rem := int(math.Round(100 - *p))
		rateParts = append(rateParts, rateStyle(rem).Render(fmt.Sprintf("7d:%d%%", rem)))
	}
	if len(rateParts) > 0 {
		line2Parts = append(line2Parts, strings.Join(rateParts, " "))
	}

	result := line1.String()
	if len(line2Parts) > 0 {
		result += "\n" + strings.Join(line2Parts, " ")
	}
	return result
}
