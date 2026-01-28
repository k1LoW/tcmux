package claude

import "testing"

func TestMayBeTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"Claude Code title", "✳ Task summary", true},
		{"Claude Code title with Japanese", "✳ 日本語タスク", true},
		{"Empty after prefix", "✳", true},
		{"Braille spinner prefix", "⠂ Task summary", true},
		{"Braille spinner with Japanese", "⠋ 日本語タスク", true},
		{"Braille spinner only", "⠙", true},
		{"Normal shell", "zsh", false},
		{"Empty title", "", false},
		{"Similar but not Claude", "* Task", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MayBeTitle(tt.title)
			if got != tt.want {
				t.Errorf("MayBeTitle(%q) = %v, want %v", tt.title, got, tt.want)			}
		})
	}
}

func TestMayBeProcess(t *testing.T) {
	tests := []struct {
		name           string
		currentCommand string
		want           bool
	}{
		{"Node process", "node", true},
		{"Claude binary", "claude", true},
		{"Zsh shell", "zsh", false},
		{"Bash shell", "bash", false},
		{"Emacs", "emacs-30.1", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MayBeProcess(tt.currentCommand)
			if got != tt.want {
				t.Errorf("MayBeProcess(%q) = %v, want %v", tt.currentCommand, got, tt.want)			}
		})
	}
}

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"Normal title", "✳ Task summary", "Task summary"},
		{"Japanese title", "✳ 日本語タスク", "日本語タスク"},
		{"With extra spaces", "✳  Multiple  spaces", "Multiple  spaces"},
		{"Only prefix", "✳", ""},
		{"Only prefix with space", "✳ ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSummary(tt.title)
			if got != tt.want {
				t.Errorf("ExtractSummary(%q) = %q, want %q", tt.title, got, tt.want)			}
		})
	}
}
