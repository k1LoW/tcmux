package claude

import "testing"

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantState   string
		wantMode    string
		wantDesc    string
	}{
		{
			name: "Idle with prompt only",
			content: `Some output
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: StateIdle,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Idle with completion suggestion",
			content: `Some output
───────────────────────────────────────
❯ Try "edit file.go to..."
───────────────────────────────────────`,
			wantState: StateIdle,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Running with Clauding",
			content: `Some output
✢ Clauding… (esc to interrupt · 1m 45s · ↓ 1.2k tokens)`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "1m 45s",
		},
		{
			name: "Running with Moseying",
			content: `Some output
✽ Moseying… (esc to interrupt · 30s · ↓ 500 tokens)`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "30s",
		},
		{
			name: "Running with Thinking",
			content: `Some output
✶ Thinking… (esc to interrupt · 2m 10s · ↓ 3k tokens)`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "2m 10s",
		},
		{
			name: "Running with Jitterbugging (middle dot)",
			content: `Some output
· Jitterbugging… (esc to interrupt · 1m 8s · ↓ 3.6k tokens · thinking)`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "1m 8s",
		},
		{
			name: "Running fallback (ctrl+c to interrupt)",
			content: `Some output
✻ Thinking… (ctrl+c to interrupt)`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Running fallback (esc to interrupt)",
			content: `Some output
✻ Processing… (esc to interrupt)`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Not Running when text mentions esc to interrupt",
			content: `Some output about "esc to interrupt" in quotes
❯ `,
			wantState: StateIdle,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Waiting with permission prompt (Yes allow once)",
			content: `Some output
Yes, allow once
Yes, allow always`,
			wantState: StateWaiting,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Waiting with confirmation prompt",
			content: `Some output
Continue? (Y/n)`,
			wantState: StateWaiting,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Idle after task completion",
			content: `Some output
✻ Cooked for 43s
───────────────────────────────────────
❯ `,
			wantState: StateIdle,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Idle with plan mode",
			content: `Some output
⏸ plan mode on
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: StateIdle,
			wantMode:  ModePlan,
			wantDesc:  "",
		},
		{
			name: "Idle with accept edits",
			content: `Some output
⏵⏵ accept edits on
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: StateIdle,
			wantMode:  ModeAcceptEdits,
			wantDesc:  "",
		},
		{
			name: "Waiting - Interview mode (asking user to select)",
			content: `  3. ドキュメントのレビュー
  4. Issue の修正
  5. Type something.
  Chat about this
  Skip interview and plan immediately
Enter to select · ↑/↓ to navigate · Esc to cancel`,
			wantState: StateWaiting,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Running with Japanese status and todo list",
			content: `· importパスを更新中… (esc to interrupt · ctrl+t to hide todos · 1m 32s · ↑ 3.4k tokens · thinking)
  ⎿  ☒ go.mod のモジュール名を変更
     ☐ 全ファイルのimportパスを更新
     ☐ README.md を更新
     ☐ その他の参照を更新
     ☐ ビルドとテストの確認

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "1m 32s",
		},
		{
			name: "Idle with trust dialog overlay",
			content: ` /Users/k1low/src/github.com/k1LoW/tcmux

 Claude Code may read, write, or execute files contained in this directory.

 Execution allowed by:

   • .claude/settings.local.json

 Learn more

 ❯ 1. Yes, proceed
   2. No, exit

 Enter to confirm · Esc to cancel

╭─── Claude Code v2.1.15 ───╮
│  Welcome back k1low!      │
╰───────────────────────────╯

───────────────────────────────────────
❯ Try "fix typecheck errors"
───────────────────────────────────────
  ? for shortcuts`,
			wantState: StateIdle,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Unknown state",
			content: `Some random output
without any recognizable pattern`,
			wantState: StateUnknown,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Empty content",
			content: "",
			wantState: StateUnknown,
			wantMode:  "",
			wantDesc:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStatus(tt.content)
			if got.State != tt.wantState {
				t.Errorf("ParseStatus().State = %q, want %q", got.State, tt.wantState)			}
			if got.Mode != tt.wantMode {
				t.Errorf("ParseStatus().Mode = %q, want %q", got.Mode, tt.wantMode)			}
			if got.Description != tt.wantDesc {
				t.Errorf("ParseStatus().Description = %q, want %q", got.Description, tt.wantDesc)			}
		})
	}
}

func TestIsSeparatorLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"Box drawing line", "─────────────────", true},
		{"Double line", "═════════════════", true},
		{"Mixed box drawing", "─═─═─═─═─", true},
		{"Text content", "Some text", false},
		{"Mixed with text", "───text───", false},
		{"Empty line", "", true}, // Empty is technically all box-drawing
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSeparatorLine(tt.line)
			if got != tt.want {
				t.Errorf("isSeparatorLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}
