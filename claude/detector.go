package claude

import (
	"strings"
	"unicode"
)

// Claude Code title prefixes
const (
	claudePrefixIdle = "✳" // Idle state
)

// ClaudePane represents a Claude Code instance running in a tmux pane.
type ClaudePane struct {
	Session    string
	Window     string
	WindowName string
	Summary    string
	Status     Status
}

// MayBeTitle checks if the pane title may indicate a Claude Code instance.
// Use together with MayBeProcess(currentCommand) to confirm.
func MayBeTitle(title string) bool {
	if strings.HasPrefix(title, claudePrefixIdle) {
		return true
	}
	// Check for Braille pattern dots (U+2800-U+28FF) used as spinner
	if len(title) > 0 {
		r := []rune(title)
		if isBraillePattern(r[0]) {
			return true
		}
	}
	return false
}

// isBraillePattern checks if a rune is a Braille pattern character (U+2800-U+28FF).
func isBraillePattern(r rune) bool {
	return unicode.In(r, unicode.Braille)
}

// MayBeProcess checks if the current command may be a Claude Code process.
// Claude Code runs as "node" (npm install) or "claude" (brew install --cask claude-code).
// Use together with MayBeTitle(title) to confirm.
func MayBeProcess(currentCommand string) bool {
	return currentCommand == "node" || currentCommand == "claude"
}

// ExtractSummary extracts the task summary from the pane title.
func ExtractSummary(title string) string {
	// Remove the "✳ " prefix
	if strings.HasPrefix(title, claudePrefixIdle) {
		summary := strings.TrimPrefix(title, claudePrefixIdle)
		return strings.TrimSpace(summary)
	}
	// Remove Braille pattern prefix (spinner)
	if len(title) > 0 {
		r := []rune(title)
		if isBraillePattern(r[0]) {
			summary := string(r[1:])
			return strings.TrimSpace(summary)
		}
	}
	return strings.TrimSpace(title)
}
