package agent

import "testing"

func TestCopilotAgent_MayBeTitle(t *testing.T) {
	agent := &CopilotAgent{}
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"Copilot CLI default title", "GitHub Copilot", true},
		{"Copilot CLI custom title", "🤖 Detailed code review", true},
		{"Any title is accepted", "zsh", true}, // Title check is permissive
		{"Empty title", "", true},              // Detection relies on process name
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.MayBeTitle(tt.title)
			if got != tt.want {
				t.Errorf("CopilotAgent.MayBeTitle(%q) = %v, want %v", tt.title, got, tt.want)
			}
		})
	}
}

func TestCopilotAgent_MayBeProcess(t *testing.T) {
	agent := &CopilotAgent{}
	tests := []struct {
		name           string
		currentCommand string
		want           bool
	}{
		{"Copilot binary", "copilot", true},
		{"Node process", "node", false}, // node alone is not enough, could be Claude Code
		{"Claude binary", "claude", false},
		{"Zsh shell", "zsh", false},
		{"Bash shell", "bash", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.MayBeProcess(tt.currentCommand)
			if got != tt.want {
				t.Errorf("CopilotAgent.MayBeProcess(%q) = %v, want %v", tt.currentCommand, got, tt.want)
			}
		})
	}
}

func TestCopilotAgent_ExtractSummary(t *testing.T) {
	agent := &CopilotAgent{}
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"Default title", "GitHub Copilot", ""},
		{"Custom title", "Some other title", "Some other title"},
		{"Title with emoji prefix", "🤖 Detailed code review", "Detailed code review"},
		{"Title with robot emoji", "🤖 Review PR", "Review PR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.ExtractSummary(tt.title)
			if got != tt.want {
				t.Errorf("CopilotAgent.ExtractSummary(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestCopilotAgent_ParseStatus(t *testing.T) {
	agent := &CopilotAgent{}
	tests := []struct {
		name      string
		content   string
		wantState string
		wantMode  string
	}{
		{
			name: "Running with Esc to cancel in parentheses",
			content: `Thinking...
∙ Processing (Esc to cancel · 100 B)`,
			wantState: StateRunning,
		},
		{
			name: "Running with middle dot spinner and Esc to cancel",
			content: `● Check git status
  $ git --no-pager status
  └ 21 lines...

● Show git diff
  $ git --no-pager diff
  └ 8 lines...

∙ Reviewing git diff (Esc to cancel · 492 B)

 ~/src/github.com/k1LoW/tcmux/.worktrees/copilot[⎇ copilot*]                                                                                                        claude-sonnet-4.5 (1x)
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯  Type @ to mention files or / for commands
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 shift+tab cycle mode · ctrl+d enqueue`,
			wantState: StateRunning,
		},
		{
			name: "Waiting with Asking user",
			content: `What would you like to do?
Asking user: Please select an option`,
			wantState: StateWaiting,
		},
		{
			name: "Waiting with selection prompt",
			content: `Select an action:
Use ↑↓ or number keys to select`,
			wantState: StateWaiting,
		},
		{
			name: "Waiting with command confirmation dialog",
			content: `● Read agent/status_copilot.go
  └ 86 lines read

○ Check if README or CLAUDE.md were updated
  $ git --no-pager diff --unified=0 README.md CLAUDE.md 2>/dev/null || echo "No changes"

╭─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ Check if README or CLAUDE.md were updated                                                                                                                                               │
│                                                                                                                                                                                         │
│ Do you want to run this command?                                                                                                                                                        │
│                                                                                                                                                                                         │
│ ❯ 1. Yes                                                                                                                                                                                │
│   2. No, and tell Copilot what to do differently (Esc to stop)                                                                                                                          │
│                                                                                                                                                                                         │
│ Confirm with number keys or ↑↓ keys and Enter, Cancel with Esc                                                                                                                          │
╰─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯`,
			wantState: StateWaiting,
		},
		{
			name: "Idle with prompt",
			content: `Previous output
❯ `,
			wantState: StateIdle,
		},
		{
			name: "Idle - quoted Esc to cancel in documentation should not match",
			content: `  - **Running**: Contains "Esc to cancel"
  - **Waiting**: Contains prompt dialogs
  - **Idle**: Prompt line starting with ❯

 ~/src/github.com/k1LoW/tcmux
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯  Type @ to mention files or / for commands
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 shift+tab cycle mode`,
			wantState: StateIdle,
		},
		{
			name: "Unknown state",
			content: `Some random output
without any recognizable pattern`,
			wantState: StateUnknown,
		},
		{
			name:      "Empty content",
			content:   "",
			wantState: StateUnknown,
		},
		{
			name: "Plan mode with Running state",
			content: `● Read proto/private/controlplane/idp/v1/resource.proto lines 1-50
  └ 55 lines read

● Read service/idp/dataplane/op/google_oauth.go
  └ 309 lines read

◎ Creating review plan (Esc to cancel · 60 B)

 ~/src/github.com/tailor-inc/platform-core-services/.worktrees/idp-google-oauth[⎇ idp-google-oauth*]                                                                claude-sonnet-4.5 (1x)
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯  Type @ to mention files or / for commands
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 plan mode · shift+tab cycle mode                                                                                                                                Remaining requests: 81.4%`,
			wantState: StateRunning,
			wantMode:  ModePlan,
		},
		{
			name: "Plan mode with Idle state",
			content: `Some previous output

 ~/src/github.com/project
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯  Type @ to mention files or / for commands
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 plan mode · shift+tab cycle mode`,
			wantState: StateIdle,
			wantMode:  ModePlan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.ParseStatus(tt.content)
			if got.State != tt.wantState {
				t.Errorf("CopilotAgent.ParseStatus().State = %q, want %q", got.State, tt.wantState)
			}
			if got.Mode != tt.wantMode {
				t.Errorf("CopilotAgent.ParseStatus().Mode = %q, want %q", got.Mode, tt.wantMode)
			}
		})
	}
}
