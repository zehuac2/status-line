package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

// StatusInput is the JSON schema Claude Code feeds to the status line command on stdin.
type StatusInput struct {
	Model     Model     `json:"model"`
	Workspace Workspace `json:"workspace"`
	Cost      Cost      `json:"cost"`
}

// Model holds information about the active Claude model.
type Model struct {
	DisplayName string `json:"display_name"`
}

// Workspace holds information about the current workspace.
type Workspace struct {
	CurrentDir string `json:"current_dir"`
}

// Cost holds session cost and diff statistics.
type Cost struct {
	TotalCostUSD      float64 `json:"total_cost_usd"`
	TotalLinesAdded   int     `json:"total_lines_added"`
	TotalLinesRemoved int     `json:"total_lines_removed"`
}

// gitBranch returns the current git branch for the given directory,
// or an empty string if the directory is not inside a git repo.
func gitBranch(dir string) string {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func main() {
	var in StatusInput
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		// Never crash Claude Code's UI — emit a safe fallback line.
		fmt.Println("status-line: failed to read input")
		return
	}

	// Styles — use fmt.Println with style.Render, NOT lipgloss writer funcs,
	// because lipgloss writers strip ANSI when stdout is not a TTY (piped).
	styleModel   := lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	styleDir     := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleBranch  := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleCost    := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleAdded   := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleRemoved := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleSep     := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	sep := styleSep.Render("  ")

	var parts []string

	// Segment 1 — model name.
	if name := in.Model.DisplayName; name != "" {
		parts = append(parts, styleModel.Render(name))
	}

	// Segment 2 — current directory (basename only).
	if dir := in.Workspace.CurrentDir; dir != "" {
		parts = append(parts, styleDir.Render(filepath.Base(dir)))
	}

	// Segment 3 — git branch (omitted when not in a repo).
	if dir := in.Workspace.CurrentDir; dir != "" {
		if branch := gitBranch(dir); branch != "" {
			parts = append(parts, styleBranch.Render("⎇ "+branch))
		}
	}

	// Segment 4 — cost + lines added / removed.
	costStr   := fmt.Sprintf("$%.4f", in.Cost.TotalCostUSD)
	addedStr  := fmt.Sprintf("+%d", in.Cost.TotalLinesAdded)
	removedStr := fmt.Sprintf("-%d", in.Cost.TotalLinesRemoved)
	costSegment := styleCost.Render(costStr) + " " +
		styleAdded.Render(addedStr) + styleSep.Render("/") + styleRemoved.Render(removedStr)
	parts = append(parts, costSegment)

	fmt.Println(strings.Join(parts, sep))
}
