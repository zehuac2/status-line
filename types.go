package main

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
