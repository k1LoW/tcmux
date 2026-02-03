package agent

import "strings"

// CopilotAgent detects and parses GitHub Copilot CLI instances.
type CopilotAgent struct{}

func (a *CopilotAgent) Type() Type {
	return TypeCopilot
}

func (a *CopilotAgent) Icon() string {
	return "⬢"
}

// MayBeTitle checks if the pane title may indicate a Copilot CLI instance.
// For Copilot CLI, we primarily rely on process name detection.
// Title check is permissive - any title is accepted when process is "copilot".
func (a *CopilotAgent) MayBeTitle(title string) bool {
	// Copilot CLI can have various titles, so we accept any non-empty title
	// The actual detection is done by MayBeProcess
	return true
}

// MayBeProcess checks if the current command may be a Copilot CLI process.
// Copilot CLI runs as "copilot".
func (a *CopilotAgent) MayBeProcess(currentCommand string) bool {
	return currentCommand == "copilot"
}

// ExtractSummary extracts the task summary from the pane title.
func (a *CopilotAgent) ExtractSummary(title string) string {
	// Remove common emoji prefixes
	title = strings.TrimSpace(title)
	if len(title) > 0 {
		r := []rune(title)
		// Skip leading emoji (if any)
		if len(r) > 0 && r[0] > 0x1F000 {
			title = strings.TrimSpace(string(r[1:]))
		}
	}
	// "GitHub Copilot" is the default title, return empty
	if title == "GitHub Copilot" {
		return ""
	}
	return title
}

// ParseStatus parses the pane content and determines the Copilot CLI status.
func (a *CopilotAgent) ParseStatus(content string) Status {
	return parseCopilotStatus(content)
}
