package tmux

import (
	"context"
	"os/exec"
	"strings"
)

// ListPanesOptions specifies options for listing panes.
type ListPanesOptions struct {
	AllSessions bool   // If true, list panes from all sessions
	Target      string // Target session name (empty means current session)
}

// CurrentSession returns the name of the current tmux session.
func CurrentSession(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "tmux", "display-message", "-p", "#{session_name}")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CapturePane captures the content of a pane and returns the last lines.
func CapturePane(ctx context.Context, paneID string) (string, error) {
	// Capture the visible pane content
	cmd := exec.CommandContext(ctx, "tmux", "capture-pane", "-t", paneID, "-p")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// InternalPaneVars are tmux variables required internally for Claude Code detection.
var InternalPaneVars = []string{
	"session_name",
	"window_index",
	"window_name",
	"pane_id",
	"pane_current_command",
	"pane_title",
}

// InternalSessionVars are tmux variables required internally for session listing.
var InternalSessionVars = []string{
	"session_name",
	"session_windows",
	"session_attached",
}

// Pane represents a tmux pane with variable values.
type Pane struct {
	Vars map[string]string
}

// ListPanes returns tmux panes with variable values.
func ListPanes(ctx context.Context, format string, vars []string, opts ListPanesOptions) ([]Pane, error) {
	args := []string{"list-panes", "-F", format}

	if opts.AllSessions {
		args = append(args, "-a")
	} else if opts.Target != "" {
		args = append(args, "-t", opts.Target, "-s")
	} else {
		args = append(args, "-s")
	}

	cmd := exec.CommandContext(ctx, "tmux", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var panes []Pane
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		varMap := make(map[string]string)
		for i, v := range vars {
			if i < len(parts) {
				varMap[v] = parts[i]
			}
		}

		panes = append(panes, Pane{Vars: varMap})
	}

	return panes, nil
}

// Session represents a tmux session with variable values.
type Session struct {
	Vars map[string]string
}

// ListSessions returns tmux sessions with variable values.
func ListSessions(ctx context.Context, format string, vars []string) ([]Session, error) {
	cmd := exec.CommandContext(ctx, "tmux", "list-sessions", "-F", format)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var sessions []Session
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		varMap := make(map[string]string)
		for i, v := range vars {
			if i < len(parts) {
				varMap[v] = parts[i]
			}
		}

		sessions = append(sessions, Session{Vars: varMap})
	}

	return sessions, nil
}
