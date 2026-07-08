package main

import "image/color"

type StatusInput struct {
	Model         Model         `json:"model"`
	Cwd           string        `json:"cwd"`
	ContextWindow ContextWindow `json:"context_window"`
	Cost          Cost          `json:"cost"`
	RateLimits    RateLimits    `json:"rate_limits"`
	Vim           Vim           `json:"vim"`
}

type Model struct {
	DisplayName string `json:"display_name"`
}

// Vim is absent from the input entirely when vim mode is disabled.
type Vim struct {
	Mode string `json:"mode"`
}

// color resolves v.Mode to its accent in t, falling back to Normal (coral)
// for NORMAL and any unrecognized mode.
func (v Vim) color(t *vimTheme) color.Color {
	switch v.Mode {
	case "INSERT":
		return t.Insert
	case "VISUAL":
		return t.Visual
	case "VISUAL LINE":
		return t.VisualLine
	case "REPLACE":
		return t.Replace
	default:
		return t.Normal
	}
}

type ContextWindow struct {
	UsedPercentage    *float64 `json:"used_percentage"`
	TotalInputTokens  *int64   `json:"total_input_tokens"`
	TotalOutputTokens *int64   `json:"total_output_tokens"`
}

type Cost struct {
	TotalCostUSD      *float64 `json:"total_cost_usd"`
	TotalDurationMs   *int64   `json:"total_duration_ms"`
	TotalLinesAdded   *int64   `json:"total_lines_added"`
	TotalLinesRemoved *int64   `json:"total_lines_removed"`
}

type RateLimits struct {
	FiveHour RateLimit `json:"five_hour"`
	SevenDay RateLimit `json:"seven_day"`
}

type RateLimit struct {
	UsedPercentage *float64 `json:"used_percentage"`
	ResetsAt       *int64   `json:"resets_at"`
}

const sampleInput = `{"model":{"display_name":"Opus"},"cwd":"/Users/zehuachen/Developer/others/status-line","context_window":{"used_percentage":8,"total_input_tokens":15000,"total_output_tokens":3200},"cost":{"total_cost_usd":0.0123,"total_duration_ms":7980000,"total_lines_added":247,"total_lines_removed":83},"rate_limits":{"five_hour":{"used_percentage":93,"resets_at":1751572500},"seven_day":{"used_percentage":96,"resets_at":1752091200}},"vim":{"mode":"NORMAL"}}`
