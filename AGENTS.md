# status-line

A Go CLI binary that renders a two-line styled terminal status bar for Claude Code.

## What it does

Reads a JSON blob from stdin, parses it, and prints a lipgloss-styled status line to stdout. Claude Code pipes a JSON payload to this binary on each turn; the output becomes the status line shown below the prompt.

**Line 1:** `<cwd-basename> git:(<branch>) [✗] ctx:<N>% tok:<N>k`
**Line 2:** `<ModelName> $<cost> 5h:<rem>% 7d:<rem>%`

## Input schema

```json
{
  "model": { "display_name": "Sonnet" },
  "cwd": "/absolute/path",
  "context_window": {
    "used_percentage": 42.5,
    "total_input_tokens": 15000,
    "total_output_tokens": 3200
  },
  "cost": { "total_cost_usd": 0.0123 },
  "rate_limits": {
    "five_hour": { "used_percentage": 30 },
    "seven_day": { "used_percentage": 15 }
  }
}
```

All numeric fields are pointers (`*float64` / `*int64`) and are omitted from output when absent.

## Key files

| File         | Purpose                                                                         |
| ------------ | ------------------------------------------------------------------------------- |
| `main.go`    | Entrypoint: parses flags, reads stdin or sample JSON, calls `render()`          |
| `types.go`   | All input types (`StatusInput`, `Model`, etc.) and `sampleInput` const          |
| `git.go`     | `gitStatus` struct and `getGitStatus()` function                                |
| `render.go`  | `rateStyle()` helper and `render()` function                                    |
| `build.go`   | Cross-compile script (`go run build.go`); tagged `//go:build ignore`            |
| `go.mod`     | Module `github.com/zehuac2/status-line`, Go 1.26, uses `charm.land/lipgloss/v2` |

## Development

```sh
# Run with sample data (no stdin needed)
go run . -claude

# Run with real JSON
echo '{"model":{"display_name":"Opus"},...}' | go run .

# Cross-compile for darwin-arm64 and linux-amd64 → dist/
go run build.go

# Build locally
go build -o status-line .
```

- Always format code with `go fmt`

## Color conventions (lipgloss ANSI 256)

| Color     | Code       | Used for                                                   |
| --------- | ---------- | ---------------------------------------------------------- |
| cyan      | `"6"`      | cwd basename, model name                                   |
| bold blue | `"4"` bold | `git:(…)` brackets                                         |
| red       | `"1"`      | branch name, rate-limit ≤20% remaining                     |
| yellow    | `"3"`      | dirty marker `✗`, token count, rate-limit 20–49% remaining |
| magenta   | `"5"`      | context window percentage                                  |
| green     | `"2"`      | cost, rate-limit ≥50% remaining                            |

## Architecture notes

- Single-package binary; no sub-packages.
- `render()` is pure (no side effects) — unit-testable without file I/O.
- `getGitStatus()` shells out to `git`; it gracefully returns `false` when the cwd is not a repo.
- The `-claude` flag is a preview mode that feeds `sampleInput` instead of stdin — useful for iterating on styling without a live Claude session.
