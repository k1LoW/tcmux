# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Run all tests with coverage
make test

# Run a single test
go test -v ./claude -run TestParseStatus

# Build the binary
make build

# Install locally
make install

# Lint (requires golangci-lint and gostyle)
make lint

# Install dev dependencies
make depsdev
```

## Architecture

tcmux is a CLI tool that displays Claude Code instances running in tmux windows/sessions along with their status.

### Core Components

- **cmd/**: CLI commands using cobra
  - `root.go` - Root command with `--color` flag
  - `lsw.go` - `list-windows` (alias: `lsw`) - Lists windows with Claude Code status
  - `ls.go` - `list-sessions` (alias: `ls`) - Lists sessions with Claude Code stats

- **claude/**: Claude Code detection and status parsing
  - `detector.go` - Detects Claude Code instances by checking pane title (✳ prefix or Braille spinner) and process name (`node` or `claude`)
  - `status.go` - Parses pane content to determine status (Idle, Running, Waiting) using regex patterns for prompts, spinners, and permission dialogs

- **tmux/**: tmux interaction via shell commands
  - `tmux.go` - Wraps `tmux list-panes`, `tmux list-sessions`, `tmux capture-pane`

- **output/**: Output formatting and colorization
  - `format.go` - Expands format strings with tmux variables and custom `#{c_status}` variable
  - `color.go` - Terminal color handling with `termenv`

### Data Flow

1. CLI command parses user format string and extracts tmux variables
2. `tmux.ListPanes()` or `tmux.ListSessions()` fetches data from tmux
3. For each pane, `claude.MayBeTitle()` + `claude.MayBeProcess()` detect Claude Code
4. `tmux.CapturePane()` gets pane content, `claude.ParseStatus()` determines state
5. `output.ExpandFormat()` generates final output with colors

### Status Detection

Claude Code status is determined by regex matching the captured pane content:
- **Running**: Matches spinner symbols with time (e.g., `✻ Thinking… (1m 30s)`) or "esc to interrupt"
- **Waiting**: Contains permission prompts like "Yes, allow once", "Run this command?", etc.
- **Idle**: Prompt line starting with `❯`
