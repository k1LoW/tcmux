package claude

import (
	"regexp"
	"strings"
)

// Status represents the status of a Claude Code instance.
type Status struct {
	State       string // Idle, Running, Completed, Unknown
	Mode        string // plan mode, accept edits, or empty
	Description string // Additional description (e.g., time elapsed)
}

// Status state constants
const (
	StateIdle    = "Idle"
	StateRunning = "Running"
	StateWaiting = "Waiting" // Agent is waiting for user input/selection
	StateUnknown = "Unknown"
)

// Mode constants
const (
	ModePlan        = "plan mode"
	ModeAcceptEdits = "accept edits"
)

var (
	// Running indicators: various symbols followed by action text and time
	// Symbols include: ✢, ✽, ✶, ✻, · (middle dot), etc.
	// Action text can be English (Clauding, Thinking, etc.) or Japanese
	// Time format: "30s", "1m 45s", "2m 10s", etc.
	runningPattern = regexp.MustCompile(`[✢✽✶✻·]\s+\S+…?\s+\([^)]*·\s*((?:\d+[smh]\s*)+)`)

	// Waiting patterns: Agent is asking for user input/selection
	// From agent-deck: permission prompts, confirmation dialogs, etc.
	waitingPatterns = []string{
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
	selectionMenuPattern = regexp.MustCompile(`❯\s+\d+\.`)

	// File changes status line pattern: "4 files +42 -0", "1 file +10 -5"
	fileChangesPattern = regexp.MustCompile(`^\s*\d+\s+files?\s+[+-]`)

	// Plan mode pattern
	planModePattern = regexp.MustCompile(`⏸\s+plan\s+mode\s+on`)

	// Accept edits pattern
	acceptEditsPattern = regexp.MustCompile(`⏵⏵\s+accept\s+edits\s+on`)

	// Idle pattern: prompt line (with or without completion suggestions)
	// Note: Claude Code uses NBSP (U+00A0) after the prompt
	// Allow optional leading whitespace for nested prompts
	idlePattern = regexp.MustCompile(`(?m)^\s*❯`)

	// Interview mode pattern: Claude Code asking user to select from options
	interviewPattern = regexp.MustCompile(`Enter to select.*↑/↓ to navigate.*Esc to cancel`)

	// Running fallback pattern (from agent-deck)
	// Matches Claude Code status line: ✻ Verb… (esc to interrupt) or (ctrl+c to interrupt)
	// Must start with status symbol to avoid matching text that mentions these phrases
	runningFallbackPattern = regexp.MustCompile(`[✢✽✶✻·]\s+\S+…?\s+\((esc|ctrl\+c) to interrupt`)
)

// ParseStatus parses the pane content and determines the status.
func ParseStatus(content string) Status {
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
	if planModePattern.MatchString(combined) {
		status.Mode = ModePlan
	} else if acceptEditsPattern.MatchString(combined) {
		status.Mode = ModeAcceptEdits
	}

	// Check for running state (primary pattern with time extraction)
	if matches := runningPattern.FindStringSubmatch(combined); len(matches) > 0 {
		status.State = StateRunning
		status.Description = strings.TrimSpace(matches[1]) // Time elapsed
		return status
	}

	// Check for running state (fallback pattern without time)
	// This ensures "esc to interrupt" is detected as Running even without time info
	if runningFallbackPattern.MatchString(combined) {
		status.State = StateRunning
		return status
	}

	// Check for idle state first: if the last meaningful line is a prompt,
	// then the agent is idle (even if there are dialogs/overlays above)
	if isPromptLine(lines) {
		status.State = StateIdle
		return status
	}

	// Check for waiting state (agent asking for user input/selection)
	// This includes permission prompts, confirmation dialogs, interview mode
	for _, pattern := range waitingPatterns {
		if strings.Contains(combined, pattern) {
			status.State = StateWaiting
			return status
		}
	}
	if interviewPattern.MatchString(combined) {
		status.State = StateWaiting
		return status
	}
	// Selection menu with numbered options (❯ 1. Yes, ❯ 2. No, etc.)
	if selectionMenuPattern.MatchString(combined) {
		status.State = StateWaiting
		return status
	}

	// Fallback: check for idle state using pattern match
	if idlePattern.MatchString(combined) {
		status.State = StateIdle
		return status
	}

	return status
}

// lastNonEmptyLines returns the last n non-empty lines.
// It skips lines that are only separators (─, ═, etc.).
func lastNonEmptyLines(lines []string, n int) []string {
	var result []string
	for i := len(lines) - 1; i >= 0 && len(result) < n; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		// Skip separator lines (lines consisting only of box-drawing characters)
		if isSeparatorLine(line) {
			continue
		}
		result = append([]string{lines[i]}, result...)
	}
	return result
}

// isPromptLine checks if the last non-empty, non-separator line is a prompt.
// This helps distinguish between "at prompt" vs "prompt visible but not at end".
func isPromptLine(lines []string) bool {
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
			fileChangesPattern.MatchString(line) {
			continue
		}
		// Check if this line is a prompt
		// Claude Code prompt: "❯" or "❯ " (possibly with suggestion)
		// But NOT selection menu markers like "❯ 1. Yes"
		if strings.HasPrefix(line, "❯") {
			// Check if it's a selection menu marker (❯ followed by number and dot)
			if selectionMenuPattern.MatchString(line) {
				return false
			}
			return true
		}
		// Not a prompt line
		return false
	}
	return false
}

// isSeparatorLine checks if a line consists only of box-drawing characters.
func isSeparatorLine(line string) bool {
	for _, r := range line {
		// Box-drawing characters range: U+2500 to U+257F
		if r < 0x2500 || r > 0x257F {
			return false
		}
	}
	return true
}
