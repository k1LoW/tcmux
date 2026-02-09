package agent

import (
	"regexp"
	"strings"
)

// Copilot CLI status indicators
var (
	// Running pattern: "(Esc to cancel" at end of a status line (not quoted text)
	// Format: ∙ Action... (Esc to cancel · 492 B)
	copilotRunningPattern = regexp.MustCompile(`\(Esc to cancel`)

	// Waiting patterns: Copilot is asking for user input
	copilotWaitingPatterns = []string{
		"Asking user:",
		"Use ↑↓ or number keys to select",
		"Do you want to run this command?",
		"Confirm with number keys",
		"Cancel with Esc",
	}

	// Plan mode pattern: "plan mode · shift+tab cycle mode" in footer
	copilotPlanModePattern = regexp.MustCompile(`plan mode\s*·`)
)

// parseCopilotStatus parses the pane content and determines the Copilot CLI status.
func parseCopilotStatus(content string) Status {
	lines := strings.Split(content, "\n")

	// Check the last few lines for status indicators
	lastLines := lastNonEmptyLines(lines, 30)
	combined := strings.Join(lastLines, "\n")

	status := Status{
		State:       StateUnknown,
		Mode:        "",
		Description: "",
	}

	// Check for plan mode first
	if copilotPlanModePattern.MatchString(combined) {
		status.Mode = ModePlan
	}

	// Check for running state: "(Esc to cancel" indicates Copilot is processing
	// Must be in parentheses to avoid matching quoted text in documentation
	if copilotRunningPattern.MatchString(combined) {
		status.State = StateRunning
		return status
	}

	// Check for idle state first: if the last meaningful line is a prompt,
	// then the agent is idle (even if there are dialogs/overlays or quoted text above)
	if isCopilotPromptLine(lines) {
		status.State = StateIdle
		return status
	}

	// Check for waiting state: Copilot is asking for user input
	for _, pattern := range copilotWaitingPatterns {
		if strings.Contains(combined, pattern) {
			status.State = StateWaiting
			return status
		}
	}

	return status
}

// isCopilotPromptLine checks if the last non-empty, non-separator line is a prompt.
func isCopilotPromptLine(lines []string) bool {
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if isSeparatorLine(line) {
			continue
		}
		// Skip common footer lines (same patterns as Claude Code)
		if strings.Contains(line, "? for shortcuts") ||
			strings.Contains(line, "ctrl+") ||
			strings.Contains(line, "shift+") ||
			strings.Contains(line, "Remaining requests:") {
			continue
		}
		// Copilot CLI prompt: "❯" (same as Claude Code)
		if strings.HasPrefix(line, "❯") {
			return true
		}
		// Not a prompt line
		return false
	}
	return false
}
