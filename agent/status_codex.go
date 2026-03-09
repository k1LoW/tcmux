package agent

import (
	"regexp"
	"strings"
)

// Codex CLI status indicators.
var (
	// Running patterns.
	codexRunningParenPattern = regexp.MustCompile(`(?i)\([^)]*(Esc to cancel|esc to interrupt|ctrl\+c to interrupt)`)

	// Waiting patterns.
	codexWaitingPatterns = []string{
		"Do you want to run this command?",
		"Do you want to allow this command?",
		"Confirm with number keys",
		"Cancel with Esc",
		"approval required",
		"Would you like to run the following command?",
		"Yes, proceed (y)",
		"No, and tell Codex what to do differently",
		"Press enter to confirm or esc to cancel",
	}

	// Mode patterns.
	codexPlanModePattern       = regexp.MustCompile(`(?i)(plan mode\s*·|collaboration mode:\s*plan|for plan mode|you are now in plan mode|#\s*plan mode\b)`)
	codexDefaultModePattern    = regexp.MustCompile(`(?i)(for default mode|you are now in default mode|collaboration mode:\s*default)`)
	codexAcceptModePattern     = regexp.MustCompile(`(?i)accept edits`)
	codexProgressFooterPattern = regexp.MustCompile(`\b\d{1,3}% left\b`)
)

// parseCodexStatus parses the pane content and determines the Codex CLI status.
func parseCodexStatus(content string) Status {
	lines := strings.Split(content, "\n")
	lastLines := lastNonEmptyLines(lines, 30)
	combined := strings.Join(lastLines, "\n")

	status := Status{
		State:       StateUnknown,
		Mode:        "",
		Description: "",
	}

	status.Mode = detectCodexMode(lastLines)

	if codexRunningParenPattern.MatchString(combined) {
		status.State = StateRunning
		return status
	}

	if isCodexPromptLine(lines) {
		status.State = StateIdle
		return status
	}

	for _, pattern := range codexWaitingPatterns {
		if strings.Contains(combined, pattern) {
			status.State = StateWaiting
			return status
		}
	}

	return status
}

func detectCodexMode(lines []string) string {
	mode := ""
	for _, line := range lines {
		switch {
		case codexDefaultModePattern.MatchString(line):
			mode = ""
		case codexPlanModePattern.MatchString(line):
			mode = ModePlan
		case codexAcceptModePattern.MatchString(line):
			mode = ModeAcceptEdits
		}
	}
	return mode
}

// isCodexPromptLine checks if the last non-empty, non-separator line is a prompt.
func isCodexPromptLine(lines []string) bool {
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if isSeparatorLine(line) {
			continue
		}
		if strings.Contains(line, "? for shortcuts") ||
			strings.Contains(line, "ctrl+") ||
			strings.Contains(line, "shift+") ||
			strings.Contains(line, "Remaining requests:") ||
			(strings.Contains(line, "·") && codexProgressFooterPattern.MatchString(line)) {
			continue
		}
		if strings.HasPrefix(line, "❯") ||
			strings.HasPrefix(line, "›") ||
			line == ">" {
			return true
		}
		return false
	}
	return false
}
