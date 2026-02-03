package agent

import "testing"

func TestDetect(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		currentCommand string
		wantType       Type
		wantNil        bool
	}{
		{
			name:           "Claude Code with node",
			title:          "✳ Task summary",
			currentCommand: "node",
			wantType:       TypeClaude,
		},
		{
			name:           "Claude Code with claude binary",
			title:          "✳ Task summary",
			currentCommand: "claude",
			wantType:       TypeClaude,
		},
		{
			name:           "Claude Code with Braille spinner",
			title:          "⠂ Task summary",
			currentCommand: "node",
			wantType:       TypeClaude,
		},
		{
			name:           "Copilot CLI with default title",
			title:          "GitHub Copilot",
			currentCommand: "copilot",
			wantType:       TypeCopilot,
		},
		{
			name:           "Copilot CLI with custom title",
			title:          "🤖 Detailed code review",
			currentCommand: "copilot",
			wantType:       TypeCopilot,
		},
		{
			name:           "Normal shell",
			title:          "zsh",
			currentCommand: "zsh",
			wantNil:        true,
		},
		{
			name:           "Claude title but wrong process",
			title:          "✳ Task summary",
			currentCommand: "zsh",
			wantNil:        true,
		},
		{
			name:           "Copilot title but wrong process",
			title:          "GitHub Copilot",
			currentCommand: "zsh",
			wantNil:        true,
		},
		{
			name:           "Node process with non-Claude title is not detected",
			title:          "some random title",
			currentCommand: "node",
			wantNil:        true,
		},
		{
			name:           "Empty title",
			title:          "",
			currentCommand: "node",
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Detect(tt.title, tt.currentCommand)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Detect(%q, %q) = %v, want nil", tt.title, tt.currentCommand, got.Type())
				}
				return
			}
			if got == nil {
				t.Errorf("Detect(%q, %q) = nil, want %v", tt.title, tt.currentCommand, tt.wantType)
				return
			}
			if got.Type() != tt.wantType {
				t.Errorf("Detect(%q, %q).Type() = %v, want %v", tt.title, tt.currentCommand, got.Type(), tt.wantType)
			}
		})
	}
}
