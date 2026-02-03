package agent

import "strings"

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
