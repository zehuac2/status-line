# status-line

A Go CLI binary that renders a three-line styled terminal status bar for Claude Code.

## What it does

Reads a JSON blob from stdin, parses it, and prints a lipgloss-styled status line to stdout. Claude Code pipes a JSON payload to this binary on each turn; the output becomes the status line shown below the prompt.

**Line 1:** `<cwd-basename> git:(<branch>) ✦ <ModelName> ctx <bar>`
**Line 2:** `▲<lines-added> ▼<lines-removed> ⧗ <session-duration>`
**Line 3:** `$<cost> 5h <bar> 7d <bar> ↺ <rate-limit-reset-time>`

`<bar>` is a 10-character block gauge (`bar()` in render.go) built from a percentage — full `█` blocks, one faint `█` remainder cell (rounded to the nearest eighth of a cell), `░` padding. The remainder cell is a dimmed solid block rather than a fractional glyph (▏▎▍▌▋▊▉), since those render inconsistently across monospace fonts.

Any segment whose backing field is absent is omitted, and a whole line collapses (not printed as blank) if every one of its segments is absent.

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
  "cost": {
    "total_cost_usd": 0.0123,
    "total_duration_ms": 7980000,
    "total_lines_added": 247,
    "total_lines_removed": 83
  },
  "rate_limits": {
    "five_hour": { "used_percentage": 30, "resets_at": 1751572500 },
    "seven_day": { "used_percentage": 15, "resets_at": 1752091200 }
  }
}
```

All numeric fields are pointers (`*float64` / `*int64`) and are omitted from output when absent. `resets_at` is Unix epoch seconds; the reset-time segment prefers `five_hour.resets_at`, falling back to `seven_day.resets_at`.

## Key files

| File         | Purpose                                                                         |
| ------------ | ------------------------------------------------------------------------------- |
| `main.go`    | Entrypoint: parses flags, reads stdin or sample JSON, calls `render()`          |
| `types.go`   | All input types (`StatusInput`, `Model`, etc.) and `sampleInput` const          |
| `git.go`     | `getGitBranch()` function                                                       |
| `render.go`  | `bar()`, `rateStyle()`, `row()` helpers and `render()` function                 |
| `build.go`   | Cross-compile + package script (`go run build.go`); tagged `//go:build ignore`  |
| `go.mod`     | Module `github.com/zehuac2/status-line`, Go 1.26, uses `charm.land/lipgloss/v2` |

## Development

```sh
# Run with sample data (no stdin needed)
go run . -claude

# Run with real JSON
echo '{"model":{"display_name":"Opus"},...}' | go run .

# Cross-compile + package for darwin-arm64, linux-amd64, linux-arm64,
# windows-amd64, windows-arm64 → dist/*.tar.gz / dist/*.zip + dist/SHA256SUMS.txt
go run build.go

# Build locally
go build -o status-line .
```

- Always format code with `go fmt`

## Color conventions (lipgloss ANSI 256)

Every segment renders bold, matching the design's block-wide `font-weight:700`.

| Color   | Code  | Used for                                             |
| ------- | ----- | ----------------------------------------------------- |
| cyan    | `"6"` | cwd basename                                           |
| blue    | `"4"` | `git:(…)` brackets                                     |
| red     | `"1"` | branch name, `▼` lines removed, rate-limit ≤20% remaining |
| yellow  | `"3"` | rate-limit 20–49% remaining                            |
| magenta | `"5"` | ctx bar                                                |
| green   | `"2"` | model name, `▲` lines added, cost, rate-limit ≥50% remaining |
| gray    | `"8"` | session duration, rate-limit reset time                |

The 5h/7d bars are colored by `rateStyle()`, keyed off *remaining* percentage (100 − used), not a fixed color — the whole `5h <bar>` / `7d <bar>` segment takes the severity color.

## Architecture notes

- Single-package binary; no sub-packages.
- `render()` is pure (no side effects) — unit-testable without file I/O.
- `getGitBranch()` shells out to `git`; it gracefully returns `false` when the cwd is not a repo. It only resolves a branch (or short SHA for detached HEAD) — no dirty-tree check, since the design has no dirty indicator.
- Segments within a line are assembled with `lipgloss.JoinHorizontal` (`row()` wraps it, skipping empty segments); the three lines are plain `strings.Join`ed with `"\n"` rather than `lipgloss.JoinVertical`, since that would pad shorter lines with trailing spaces to match the widest one.
- The `-claude` flag is a preview mode that feeds `sampleInput` instead of stdin — useful for iterating on styling without a live Claude session.

## Releases

`build.go` packages each target as an archive (`.tar.gz` on darwin/linux, `.zip` on
windows) containing a single binary named `status-line` (`status-line.exe` on windows),
plus a `dist/SHA256SUMS.txt` covering all archives. This naming/format is deliberate so the
GitHub release assets are installable via [mise's `github:` backend](https://mise.jdx.dev/dev-tools/backends/github.html)
(`mise use github:zehuac2/status-line`), which autodetects platform from OS/arch tokens in
the filename and scores archive formats over bare binaries. `.github/workflows/release.yml`
uploads everything `build.go` emits on `release: published`.
