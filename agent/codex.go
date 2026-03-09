package agent

import "strings"

// CodexAgent detects and parses Codex CLI instances.
type CodexAgent struct{}

func (a *CodexAgent) Type() Type {
	return TypeCodex
}

func (a *CodexAgent) Icon() string {
	return "❂"
}

// MayBeTitle checks if the pane title may indicate a Codex CLI instance.
// Title check is permissive - process name is the primary signal.
func (a *CodexAgent) MayBeTitle(title string) bool {
	return true
}

// MayBeProcess checks if the current command may be a Codex CLI process.
func (a *CodexAgent) MayBeProcess(currentCommand string) bool {
	currentCommand = strings.ToLower(strings.TrimSpace(currentCommand))
	return currentCommand == "codex" || strings.HasPrefix(currentCommand, "codex-")
}

// ExtractSummary extracts the task summary from the pane title.
func (a *CodexAgent) ExtractSummary(title string) string {
	// Remove common emoji prefixes.
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	if len(title) > 0 {
		r := []rune(title)
		if len(r) > 0 && r[0] > 0x1F000 {
			title = strings.TrimSpace(string(r[1:]))
		}
	}

	// Treat default Codex title / host-like local title as empty summary.
	if title == "Codex" || strings.HasSuffix(strings.ToLower(title), ".local") {
		return ""
	}

	return title
}

// ParseStatus parses the pane content and determines the Codex CLI status.
func (a *CodexAgent) ParseStatus(content string) Status {
	return parseCodexStatus(content)
}
