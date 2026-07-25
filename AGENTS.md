# status-line

A Go CLI binary that renders a three-line styled terminal status bar for Claude
Code, with an optional vim-mode row prepended.

## What it does

Reads a JSON blob from stdin, parses it, and prints a lipgloss-styled status
line to stdout. Claude Code pipes a JSON payload to this binary on each turn;
the output becomes the status line shown below the prompt.

The lines are framed by a corner-bracket box (`components.Box()`) — just the
four rounded corners (`╭ ╮ ╰ ╯`), no connecting edges.

**Mode row (optional):** `mode <VimMode>` — only rendered when vim mode is
enabled, followed by a `─` divider rule before the identity row. **Identity row
(line 1):** `<cwd-basename> git:(<branch>) ✦ <ModelName> ctx <percentage>`.
**Usage row (line 2):**
`$<cost> <effort> 5h <percentage> 7d <percentage> ↺ <rate-limit-reset-time>`
**Activity row (line 3):**
`▲<lines-added> ▼<lines-removed> ⧗ <session-duration>`

`<percentage>` is the whole-number percentage (rounded to the nearest integer)
for the relevant field — for example, `ctx 8%`, `5h 93%`, or `7d 96%`.

Any segment whose backing field is absent is omitted, and a whole line collapses
(not printed as blank) if every one of its segments is absent. If all lines
collapse, the box itself is omitted too — no empty frame is printed.

## Input schema

The full JSON schema is published by Anthropic and drifts over time, so it is
**not** duplicated here. Before changing how the payload is parsed or rendered,
fetch the authoritative schema from
<https://code.claude.com/docs/en/statusline#full-json-schema> (use the WebFetch
tool) and reconcile `types.go` against it. The notes below describe how this
binary interprets the payload today and are not a substitute for the live
schema.

All numeric fields are pointers (`*float64` / `*int64`) and are omitted from
output when absent. `branch` overrides the git branch lookup; if omitted, the
branch is resolved by shelling out to `git` from `cwd`. `effort` is optional and
renders on the usage row when present. `resets_at` is Unix epoch seconds; the
reset-time segment prefers `five_hour.resets_at`, falling back to
`seven_day.resets_at`.

`vim` is absent from the payload entirely when vim mode is disabled — not just
`vim.mode` being empty. `vim.mode` is one of `NORMAL`, `INSERT`, `VISUAL`, or
`VISUAL LINE`. Set `"hideVimModeIndicator": true` in the status-line settings so
Claude Code's built-in `-- INSERT --` text isn't shown twice alongside this
binary's own mode row.

## Key files

| File                           | Purpose                                                                                                                                           |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `main.go`                      | Entrypoint: parses flags, reads stdin or sample JSON, builds the theme, calls `render()`                                                          |
| `types.go`                     | All input types (`StatusInput`, `Model`, etc.) and `sampleInput` const                                                                            |
| `git.go`                       | `getGitBranch()` function                                                                                                                         |
| `render.go`                    | `render()` plus `renderIdentityRow()`, `renderUsageRow()`, `renderActivityRow()`, one per line, assembling segments and calling into `components` |
| `theme.go`                     | `theme`/`vimTheme` structs and `claudeTheme()` constructor centralizing every color                                                               |
| `components/bar.go`            | `components.Bar()` block-gauge helper                                                                                                             |
| `components/row.go`            | `components.Row()` segment-joining helper                                                                                                         |
| `components/box.go`            | `components.Box()` corner-bracket framing helper                                                                                                  |
| `build.go`                     | Cross-compile + package script (`go run build.go`); tagged `//go:build ignore`                                                                    |
| `go.mod`                       | Module `github.com/zehuac2/status-line`, Go 1.26, uses `charm.land/lipgloss/v2`                                                                   |
| `homebrew/status-line.rb.tmpl` | Homebrew formula template, rendered and pushed to the `zehuac2/homebrew-tools` tap on each release                                                |

## Development

```sh
# Run with sample data (no stdin needed)
go run . -claude

# Run with real JSON
echo '{"model":{"display_name":"Opus"},...}' | go run .

# Cross-compile + package for darwin-arm64, linux-x64, linux-arm64,
# windows-x64, windows-arm64 → dist/*.tar.gz / dist/*.zip + dist/SHA256SUMS.txt
go run build.go

# Build locally
go build -o status-line .
```

- Always format code with `go fmt`
- Chnages to markdown files should be formatted with `bunx prettier`

## Color conventions (lipgloss truecolor)

Colors are centralized in the `theme` struct (`theme.go`), constructed by
`claudeTheme()` and passed into `render()` — not hardcoded `lipgloss.Color`
literals scattered through `render.go`.

Most segments render bold, matching the design's block-wide `font-weight:700`;
the cwd basename and `ctx <percentage>` segment are normal weight (the design
overrides those to `400`).

| Color        | Theme field | Used for                                                                                          |
| ------------ | ----------- | ------------------------------------------------------------------------------------------------- |
| text         | `Text`      | cwd basename, `git:(…)` brackets, `✦`, `ctx <percentage>`, `▲added ▼removed`, `$cost`, `↺`        |
| text dim     | `TextDim`   | session duration, whole `7d <percentage>` segment, `mode` label (normal weight)                   |
| Claude coral | `Primary`   | branch name, model name, `effort`, whole `5h <percentage>` segment, reset time, `NORMAL` vim mode |
| divider      | `Divider`   | the `─` rule between the mode row and the identity row (same hue as `TextDim`)                    |

Box corners (`components.Box()`) are unstyled — they render in the terminal's
default foreground, not a themed color.

`5h` and `7d` are fixed colors (coral / dim gray) rather than keyed off
remaining rate-limit percentage — there's no severity coloring.

The vim mode value itself is colored per-mode (bold): `NORMAL`, `INSERT`,
`VISUAL` / `VISUAL LINE`, and `REPLACE` (kept for design fidelity even though
Claude Code doesn't currently emit it). The per-mode accent for each is in
`vimTheme` in `theme.go`. An unrecognized mode string falls back to the `NORMAL`
coral.

## Architecture notes

- `main` holds the input types, git lookup, the theme, and `render()`; the
  presentational helpers (`Bar`, `Row`, `Box`) live in the `components`
  sub-package and are imported as `github.com/zehuac2/status-line/components`.
- `render()` builds the mode row and divider, then delegates each of the three
  lines to its own function — `renderIdentityRow()`, `renderUsageRow()`,
  `renderActivityRow()` — before framing them all with `components.Box()`. Each
  row function builds its own `lipgloss.Style`s from the `*theme` it's passed,
  rather than sharing styles built in `render()`.
- Colors are a `*theme` argument to `render()` rather than package-level
  constants, so an alternate palette could be swapped in by constructing a
  different `*theme` — `main()` currently always builds `claudeTheme()`.
- `render()` and the row functions are pure (no side effects) — unit-testable
  without file I/O.
- `getGitBranch()` shells out to `git`; it gracefully returns `false` when the
  cwd is not a repo. It only resolves a branch (or short SHA for detached HEAD)
  — no dirty-tree check, since the design has no dirty indicator.
- Segments within a line are assembled with `lipgloss.JoinHorizontal`
  (`components.Row()` wraps it, skipping empty segments); the lines are then
  framed by `components.Box()`, which uses a `lipgloss.Border` of just the four
  corner glyphs and `lipgloss.JoinVertical` — the connecting edges are set to
  U+2800 (blank braille pattern) rather than a literal space, since Claude
  Code's status line strips leading whitespace per line, which would otherwise
  collapse the left border and misalign content under the top-left corner.
  `Box()` also filters out empty lines before joining, so a collapsed line (or
  the divider next to one) never prints a blank row.
- The divider's width is `lipgloss.Width(row)` maxed over the mode row and the
  identity/usage/activity rows — not `len()`, since the rows carry ANSI styling
  and wide/multibyte glyphs (`▲ ✦ ↺`) whose byte length doesn't match their
  terminal cell width. `lipgloss.Width` strips ANSI and measures true cell
  width, so the rule spans exactly the box's widest content row regardless of
  which segments are present.
- The `-claude` flag is a preview mode that feeds `sampleInput` instead of stdin
  — useful for iterating on styling without a live Claude session.

## Releases

`build.go` packages each target as an archive (`.tar.gz` on darwin/linux, `.zip`
on windows) containing a single binary named `status-line` (`status-line.exe` on
windows), plus a `dist/SHA256SUMS.txt` covering all archives. This naming/format
is deliberate so the GitHub release assets are installable via
[mise's `github:` backend](https://mise.jdx.dev/dev-tools/backends/github.html)
(`mise use github:zehuac2/status-line`), which autodetects platform from OS/arch
tokens in the filename and scores archive formats over bare binaries.
`.github/workflows/create-release-assets.yml` uploads everything `build.go`
emits on `release: published`.

`.github/workflows/publish-brew-release.yml` runs after
`create-release-assets.yml` completes successfully (`workflow_run`) and
publishes a Homebrew formula to the `zehuac2/homebrew-tools` tap
(`Formula/status-line.rb`). It reads the just-published release's
`SHA256SUMS.txt`, renders `homebrew/status-line.rb.tmpl` with the version, tag,
and per-archive checksums, then commits and pushes the rendered formula to the
tap repo. The formula only covers `darwin-arm64`, `linux-arm64`, and `linux-x64`
— the archives `build.go` actually produces for those OSes; Homebrew doesn't run
on Windows, and there's no `darwin-amd64` (Intel Mac) build. Pushing to the tap
repo requires a `HOMEBREW_TAP_TOKEN` secret (a personal access token with
`contents: write` on `zehuac2/homebrew-tools`).
