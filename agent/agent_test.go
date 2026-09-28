package agent

import (
	"testing"
)

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
			name:           "Claude Code with version string (Native Install)",
			title:          "✳ Task summary",
			currentCommand: "2.1.34",
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
			name:           "Codex CLI",
			title:          "Codex",
			currentCommand: "codex",
			wantType:       TypeCodex,
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

func TestDetectWithChildCmds(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		currentCommand string
		childCmds      []string
		wantType       Type
		wantNil        bool
	}{
		{
			name:           "Claude title, zsh command, claude in children",
			title:          "✳ Task summary",
			currentCommand: "zsh",
			childCmds:      []string{"sh", "claude"},
			wantType:       TypeClaude,
		},
		{
			name:           "Claude title, zsh command, node in children",
			title:          "✳ Task summary",
			currentCommand: "zsh",
			childCmds:      []string{"node"},
			wantType:       TypeClaude,
		},
		{
			name:           "Claude title, zsh command, no agent child",
			title:          "✳ Task summary",
			currentCommand: "zsh",
			childCmds:      []string{"bash", "sh"},
			wantNil:        true,
		},
		{
			name:           "Copilot title, zsh command, copilot in children",
			title:          "GitHub Copilot",
			currentCommand: "zsh",
			childCmds:      []string{"copilot"},
			wantType:       TypeCopilot,
		},
		{
			name:           "Normal shell title with agent child is detected via child process",
			title:          "some random title",
			currentCommand: "zsh",
			childCmds:      []string{"node"},
			wantType:       TypeClaude,
		},
		{
			name:           "Claude in neovim terminal (nvim title, nvim command, claude in children)",
			title:          "nvim",
			currentCommand: "nvim",
			childCmds:      []string{"node"},
			wantType:       TypeClaude,
		},
		{
			name:           "Claude in neovim terminal with claude binary child",
			title:          "nvim",
			currentCommand: "nvim",
			childCmds:      []string{"sh", "claude"},
			wantType:       TypeClaude,
		},
		{
			name:           "Direct detection still works with child cmds",
			title:          "✳ Task summary",
			currentCommand: "claude",
			childCmds:      []string{"sh"},
			wantType:       TypeClaude,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectWithChildCmds(tt.title, tt.currentCommand, tt.childCmds)
			if tt.wantNil {
				if got != nil {
					t.Errorf("detectWithChildCmds(%q, %q, %v) = %v, want nil", tt.title, tt.currentCommand, tt.childCmds, got.Type())
				}
				return
			}
			if got == nil {
				t.Errorf("detectWithChildCmds(%q, %q, %v) = nil, want %v", tt.title, tt.currentCommand, tt.childCmds, tt.wantType)
				return
			}
			if got.Type() != tt.wantType {
				t.Errorf("detectWithChildCmds(%q, %q, %v).Type() = %v, want %v", tt.title, tt.currentCommand, tt.childCmds, got.Type(), tt.wantType)
			}
		})
	}
}

func TestDetectFromTree_integration(t *testing.T) {
	// When Claude title is present and current command is not an agent,
	// DetectFromTree should try child processes (real ps call)
	// We just verify it doesn't panic and returns consistently
	got := DetectFromTree(t.Context(), "✳ Task summary", "zsh", "1")
	// PID 1 (launchd/systemd) won't have claude as child in test env, so result should be nil
	if got != nil {
		t.Logf("Unexpected agent detected from PID 1: %v (may be OK in some environments)", got.Type())
	}
}
