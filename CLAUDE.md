# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Run all tests with coverage
make test

# Run a single test
go test -v ./agent -run TestClaudeAgent_ParseStatus

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

tcmux is a CLI tool that displays coding agent instances (Claude Code, GitHub Copilot CLI) running in tmux windows/sessions along with their status.

### Core Components

- **cmd/**: CLI commands using cobra
  - `root.go` - Root command with `--color` flag
  - `lsw.go` - `list-windows` (alias: `lsw`) - Lists windows with coding agent status
  - `ls.go` - `list-sessions` (alias: `ls`) - Lists sessions with coding agent stats

- **agent/**: Coding agent detection and status parsing
  - `agent.go` - `Detector` interface and `Detect()` function for agent detection
  - `claude.go` - Claude Code detection (pane title with `✳` prefix or Braille spinner, process `claude` or `node`)
  - `status_claude.go` - Claude Code status parsing
  - `copilot.go` - Copilot CLI detection (process `copilot`)
  - `status_copilot.go` - Copilot CLI status parsing
  - `status.go` - Common status utilities

- **tmux/**: tmux interaction via shell commands
  - `tmux.go` - Wraps `tmux list-panes`, `tmux list-sessions`, `tmux capture-pane`

- **output/**: Output formatting and colorization
  - `format.go` - Expands format strings with tmux variables and custom `#{agent_status}` variable
  - `color.go` - Terminal color handling with `termenv`

### Data Flow

1. CLI command parses user format string and extracts tmux variables
2. `tmux.ListPanes()` or `tmux.ListSessions()` fetches data from tmux
3. For each pane, `agent.Detect()` checks if it's a coding agent (Claude Code or Copilot CLI)
4. `tmux.CapturePane()` gets pane content, agent's `ParseStatus()` determines state
5. `output.ExpandFormat()` generates final output with colors

### Status Detection

**Claude Code** status is determined by regex matching the captured pane content:
- **Running**: Matches spinner symbols with time (e.g., `✻ Thinking… (1m 30s)`) or "esc to interrupt"
- **Waiting**: Contains permission prompts like "Yes, allow once", "Run this command?", etc.
- **Idle**: Prompt line starting with `❯`

**GitHub Copilot CLI** status:
- **Running**: Contains "Esc to cancel"
- **Waiting**: Contains "Do you want to run this command?", "Confirm with number keys", etc.
- **Idle**: Prompt line starting with `❯`
