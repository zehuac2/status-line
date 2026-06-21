package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

type StatusInput struct {
	Model         Model         `json:"model"`
	Cwd           string        `json:"cwd"`
	ContextWindow ContextWindow `json:"context_window"`
	Cost          Cost          `json:"cost"`
	RateLimits    RateLimits    `json:"rate_limits"`
}

type Model struct {
	DisplayName string `json:"display_name"`
}

type ContextWindow struct {
	UsedPercentage    *float64 `json:"used_percentage"`
	TotalInputTokens  *int64   `json:"total_input_tokens"`
	TotalOutputTokens *int64   `json:"total_output_tokens"`
}

type Cost struct {
	TotalCostUSD *float64 `json:"total_cost_usd"`
}

type RateLimits struct {
	FiveHour RateLimit `json:"five_hour"`
	SevenDay RateLimit `json:"seven_day"`
}

type RateLimit struct {
	UsedPercentage *float64 `json:"used_percentage"`
}

const sampleInput = `{"model":{"display_name":"Opus"},"cwd":"/Users/zehuachen/Developer/others/status-line","context_window":{"used_percentage":42.5,"total_input_tokens":15000,"total_output_tokens":3200},"cost":{"total_cost_usd":0.0123},"rate_limits":{"five_hour":{"used_percentage":30},"seven_day":{"used_percentage":15}}}`

type gitStatus struct {
	branch string
	dirty  bool
}

func getGitStatus(dir string) (gitStatus, bool) {
	if err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run(); err != nil {
		return gitStatus{}, false
	}

	var branch string
	out, err := exec.Command("git", "-C", dir, "symbolic-ref", "--short", "HEAD").Output()
	if err == nil {
		branch = strings.TrimSpace(string(out))
	} else {
		out, err = exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
		if err == nil {
			branch = strings.TrimSpace(string(out))
		}
	}

	if branch == "" {
		return gitStatus{}, false
	}

	out, _ = exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	dirty := strings.TrimSpace(string(out)) != ""

	return gitStatus{branch: branch, dirty: dirty}, true
}

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

func main() {
	claude := flag.Bool("claude", false, "render the status line from built-in sample JSON instead of stdin")
	flag.Parse()

	var r io.Reader
	if *claude {
		r = strings.NewReader(sampleInput)
	} else {
		r = os.Stdin
	}

	var in StatusInput
	if err := json.NewDecoder(r).Decode(&in); err != nil {
		fmt.Println("status-line: failed to read input")
		return
	}

	fmt.Println(render(in))
}
