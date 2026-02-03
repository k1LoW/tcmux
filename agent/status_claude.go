package agent

import (
	"regexp"
	"strings"
)

var (
	// Running indicators: various symbols followed by action text and time
	// Symbols include: ✢, ✽, ✶, ✻, · (middle dot), etc.
	// Action text can be English (Clauding, Thinking, etc.) or Japanese
	// Action text may contain spaces (e.g., "Adding types to file.ts")
	// Time format: "30s", "1m 45s", "2m 10s", etc.
	// Pattern must start at beginning of line to avoid matching quoted text
	// Format 1: (esc to interrupt · 1m 45s · ...) - time after middle dot
	claudeRunningPattern = regexp.MustCompile(`(?m)^[✢✽✶✻·]\s+.+?…?\s*\([^)]*·\s*((?:\d+[smh]\s*)+)`)

	// Format 2: (1m 52s · ...) - time at the beginning of parentheses
	claudeRunningPatternTimeFirst = regexp.MustCompile(`(?m)^[✢✽✶✻·]\s+.+?…?\s*\(((?:\d+[smh]\s*)+)\s*·`)

	// "esc to interrupt" at the end of status line indicates Running
	claudeEscToInterruptEndPattern = regexp.MustCompile(`(?m)·\s*esc to interrupt\s*$`)

	// Waiting patterns: Agent is asking for user input/selection
	// From agent-deck: permission prompts, confirmation dialogs, etc.
	claudeWaitingPatterns = []string{
		"Yes, allow once",
		"Yes, allow always",
		"Allow once",
		"Allow always",
		"❯ Yes",
		"❯ No",
		"Do you trust",
		"Run this command?",
		"Allow this MCP server",
		"Continue?",
		"Proceed?",
		"Do you want to proceed?",
		"(Y/n)",
		"(y/N)",
		"[Y/n]",
		"[y/N]",
	}

	// Selection menu pattern: numbered options with ❯ marker
	// e.g., "❯ 1. Yes", "❯ 2. No"
	claudeSelectionMenuPattern = regexp.MustCompile(`❯\s+\d+\.`)

	// File changes status line pattern: "4 files +42 -0", "1 file +10 -5"
	claudeFileChangesPattern = regexp.MustCompile(`^\s*\d+\s+files?\s+[+-]`)

	// Plan mode pattern
	claudePlanModePattern = regexp.MustCompile(`⏸\s+plan\s+mode\s+on`)

	// Accept edits pattern
	claudeAcceptEditsPattern = regexp.MustCompile(`⏵⏵\s+accept\s+edits\s+on`)

	// Idle pattern: prompt line (with or without completion suggestions)
	// Note: Claude Code uses NBSP (U+00A0) after the prompt
	// Allow optional leading whitespace for nested prompts
	claudeIdlePattern = regexp.MustCompile(`(?m)^\s*❯`)

	// Interview mode pattern: Claude Code asking user to select from options
	claudeInterviewPattern = regexp.MustCompile(`Enter to select.*↑/↓ to navigate.*Esc to cancel`)

	// Running fallback pattern (from agent-deck)
	// Matches Claude Code status line: ✻ Verb… (esc to interrupt) or (ctrl+c to interrupt)
	// Must start at beginning of line to avoid matching quoted text
	// Action text may contain spaces
	claudeRunningFallbackPattern = regexp.MustCompile(`(?m)^[✢✽✶✻·]\s+.+?…?\s*\((esc|ctrl\+c) to interrupt`)
)

// parseClaudeStatus parses the pane content and determines the Claude Code status.
func parseClaudeStatus(content string) Status {
	// Get the last non-empty lines for analysis
	lines := strings.Split(content, "\n")

	// Check the last few lines (status might not be on the very last line)
	// Use 30 lines to ensure we capture status indicators even when output is verbose
	lastLines := lastNonEmptyLines(lines, 30)
	combined := strings.Join(lastLines, "\n")

	status := Status{
		State:       StateUnknown,
		Mode:        "",
		Description: "",
	}

	// Check for modes first
	if claudePlanModePattern.MatchString(combined) {
		status.Mode = ModePlan
	} else if claudeAcceptEditsPattern.MatchString(combined) {
		status.Mode = ModeAcceptEdits
	}

	// Check for running state (primary pattern with time extraction)
	// Format 1: (esc to interrupt · 1m 45s · ...) - time after middle dot
	if matches := claudeRunningPattern.FindStringSubmatch(combined); len(matches) > 0 {
		status.State = StateRunning
		status.Description = strings.TrimSpace(matches[1]) // Time elapsed
		return status
	}

	// Format 2: (1m 52s · ...) - time at the beginning of parentheses
	if matches := claudeRunningPatternTimeFirst.FindStringSubmatch(combined); len(matches) > 0 {
		status.State = StateRunning
		status.Description = strings.TrimSpace(matches[1]) // Time elapsed
		return status
	}

	// Check for running state (fallback pattern without time)
	// This ensures "esc to interrupt" is detected as Running even without time info
	if claudeRunningFallbackPattern.MatchString(combined) {
		status.State = StateRunning
		return status
	}

	// Check for "esc to interrupt" at end of status line (e.g., "4 files +20 -0 · esc to interrupt")
	if claudeEscToInterruptEndPattern.MatchString(combined) {
		status.State = StateRunning
		return status
	}

	// Check for idle state first: if the last meaningful line is a prompt,
	// then the agent is idle (even if there are dialogs/overlays above)
	if isClaudePromptLine(lines) {
		status.State = StateIdle
		return status
	}

	// Check for waiting state (agent asking for user input/selection)
	// This includes permission prompts, confirmation dialogs, interview mode
	for _, pattern := range claudeWaitingPatterns {
		if strings.Contains(combined, pattern) {
			status.State = StateWaiting
			return status
		}
	}
	if claudeInterviewPattern.MatchString(combined) {
		status.State = StateWaiting
		return status
	}
	// Selection menu with numbered options (❯ 1. Yes, ❯ 2. No, etc.)
	if claudeSelectionMenuPattern.MatchString(combined) {
		status.State = StateWaiting
		return status
	}

	// Fallback: check for idle state using pattern match
	if claudeIdlePattern.MatchString(combined) {
		status.State = StateIdle
		return status
	}

	return status
}

// isClaudePromptLine checks if the last non-empty, non-separator line is a prompt.
// This helps distinguish between "at prompt" vs "prompt visible but not at end".
func isClaudePromptLine(lines []string) bool {
	// Find the last meaningful line (not empty, not separator, not help text)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if isSeparatorLine(line) {
			continue
		}
		// Skip common footer lines
		if strings.Contains(line, "? for shortcuts") ||
			strings.Contains(line, "ctrl+") ||
			strings.Contains(line, "shift+") ||
			claudeFileChangesPattern.MatchString(line) {
			continue
		}
		// Check if this line is a prompt
		// Claude Code prompt: "❯" or "❯ " (possibly with suggestion)
		// But NOT selection menu markers like "❯ 1. Yes"
		if strings.HasPrefix(line, "❯") {
			// Check if it's a selection menu marker (❯ followed by number and dot)
			if claudeSelectionMenuPattern.MatchString(line) {
				return false
			}
			return true
		}
		// Not a prompt line
		return false
	}
	return false
}
