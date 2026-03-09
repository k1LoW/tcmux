package agent

import "testing"

func TestCodexAgent_MayBeTitle(t *testing.T) {
	agent := &CodexAgent{}
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"Codex default title", "Codex", true},
		{"Custom title", "Implement feature", true},
		{"Any title is accepted", "zsh", true},
		{"Empty title", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.MayBeTitle(tt.title)
			if got != tt.want {
				t.Errorf("CodexAgent.MayBeTitle(%q) = %v, want %v", tt.title, got, tt.want)
			}
		})
	}
}

func TestCodexAgent_MayBeProcess(t *testing.T) {
	agent := &CodexAgent{}
	tests := []struct {
		name           string
		currentCommand string
		want           bool
	}{
		{"Codex binary", "codex", true},
		{"Codex wrapped binary name", "codex-aarch64-a", true},
		{"Copilot binary", "copilot", false},
		{"Claude binary", "claude", false},
		{"Zsh shell", "zsh", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.MayBeProcess(tt.currentCommand)
			if got != tt.want {
				t.Errorf("CodexAgent.MayBeProcess(%q) = %v, want %v", tt.currentCommand, got, tt.want)
			}
		})
	}
}

func TestCodexAgent_ExtractSummary(t *testing.T) {
	agent := &CodexAgent{}
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"Default title", "Codex", ""},
		{"Custom title", "Implement feature", "Implement feature"},
		{"Trim spaces", "  Fix bug  ", "Fix bug"},
		{"Title with emoji prefix", "🤖 Refactor parser", "Refactor parser"},
		{"Local host title", "tailor.local", ""},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.ExtractSummary(tt.title)
			if got != tt.want {
				t.Errorf("CodexAgent.ExtractSummary(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestCodexAgent_ParseStatus(t *testing.T) {
	agent := &CodexAgent{}
	tests := []struct {
		name      string
		content   string
		wantState string
		wantMode  string
	}{
		{
			name: "Running with Esc to cancel",
			content: `Planning update...
● Updating files (Esc to cancel · 120 B)`,
			wantState: StateRunning,
		},
		{
			name:      "Running with elapsed and esc to interrupt",
			content:   `• Working (29s • esc to interrupt)`,
			wantState: StateRunning,
		},
		{
			name: "Waiting with command approval",
			content: `Do you want to allow this command?
Confirm with number keys or ↑↓ keys and Enter, Cancel with Esc`,
			wantState: StateWaiting,
		},
		{
			name: "Waiting with codex approval prompt",
			content: `Would you like to run the following command?
› 1. Yes, proceed (y)
  2. Yes, and don't ask again
  3. No, and tell Codex what to do differently (esc)
Press enter to confirm or esc to cancel`,
			wantState: StateWaiting,
		},
		{
			name: "Idle with prompt",
			content: `Previous output
❯ `,
			wantState: StateIdle,
		},
		{
			name: "Idle with model progress footer",
			content: `Previous output
❯ Type a message
gpt-5.3-codex high · 67% left · ~/src/github.com/k1LoW/tcmux`,
			wantState: StateIdle,
		},
		{
			name: "Plan mode with Idle",
			content: `Some output
• Model changed to gpt-5.3-codex medium for Plan mode.
❯ Type a message`,
			wantState: StateIdle,
			wantMode:  ModePlan,
		},
		{
			name: "Default mode overrides previous plan mode",
			content: `• Model changed to gpt-5.3-codex medium for Plan mode.
• Model changed to gpt-5.3-codex high for Default mode.
❯`,
			wantState: StateIdle,
			wantMode:  "",
		},
		{
			name: "Unknown state",
			content: `Some random output
without any recognizable pattern`,
			wantState: StateUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.ParseStatus(tt.content)
			if got.State != tt.wantState {
				t.Errorf("CodexAgent.ParseStatus().State = %q, want %q", got.State, tt.wantState)
			}
			if got.Mode != tt.wantMode {
				t.Errorf("CodexAgent.ParseStatus().Mode = %q, want %q", got.Mode, tt.wantMode)
			}
		})
	}
}
