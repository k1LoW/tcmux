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
			name: "Running with time first format",
			content: `Some output
✢ Reticulating… (1m 52s · ↓ 11.5k tokens · thought for 7s)`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "1m 52s",
		},
		{
			name: "Running with esc to interrupt at end of status line",
			content: `Some output
✶ Proofing… (thinking)
───────────────────────────────────────
❯
───────────────────────────────────────
  4 files +20 -0 · esc to interrupt`,
			wantState: StateRunning,
			wantMode:  "",
			wantDesc:  "",
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
			name: "Not Running when indented status line (quoted text)",
			content: `⏺ 現在の内容は：
  ✻ Galloping… (esc to interrupt · 1m 19s · ↓ 5.9k tokens · thinking)

✻ Cooked for 1m 29s

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────`,
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
			name: "Idle after task completion with file changes",
			content: `⏺ Window 10 is now Idle (accept edits), Window 11 shows Running (4m 48s) correctly.

✻ Sautéed for 2m 55s

! make install
  ⎿  go install completed

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  4 files +73 -3`,
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
			content: ` /home/user/projects/myapp

 Claude Code may read, write, or execute files contained in this directory.

 Execution allowed by:

   • .claude/settings.local.json

 Learn more

 ❯ 1. Yes, proceed
   2. No, exit

 Enter to confirm · Esc to cancel

╭─── Claude Code v2.1.15 ───╮
│  Welcome back user!       │
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
			name: "Idle with file changes status line",
			content: `⏺ Some output about "Do you want to proceed?"

✻ Churned for 3m 5s

! make install
  ⎿  go install completed

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  4 files +42 -0`,
			wantState: StateIdle,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Waiting - Bash command confirmation dialog",
			content: `⏺ Bash(gh issue view 123 --repo owner/repo 2>/dev/null || echo "Issue #123 not found or closed")
  ⎿  title:     Fix bug in parser
     state:     CLOSED
     author:    contributor
     … +22 lines (ctrl+o to expand)

⏺ Bash(grep --help 2>/dev/null | head -10)
  ⎿  Running…

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 Bash command

   grep --help 2>/dev/null | head -10
   Check grep help

 Do you want to proceed?
 ❯ 1. Yes
   2. Yes, and don't ask again for grep commands in /home/user/projects/myapp
   3. No

 Esc to cancel · Tab to amend · ctrl+e to explain`,
			wantState: StateWaiting,
			wantMode:  "",
			wantDesc:  "",
		},
		{
			name: "Running with action text containing spaces",
			content: `      219 + export function createUserHandler(): UserHandler<UserArgs> {
      220 +   return {
      221 +     kind: "createUser",
      222 +     __args: {} as UserArgs,
      223 +   };
      224 + }

✶ Adding handler types and functions to handlers.ts… (ctrl+c to interrupt · ctrl+t to hide todos · 3m 27s · ↑ 11.0k tokens)
  ⎿  ☐ Add handler types and functions to handlers.ts
     ☐ Update Handler type in index.ts
     ☐ Add Zod schemas to schema.ts
     ☐ Export types from types.ts
     ☐ Add conversion logic to cli.ts
     ☐ Run typecheck, test, and build

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  ⏵⏵ accept edits on (shift+tab to cycle)`,
			wantState: StateRunning,
			wantMode:  ModeAcceptEdits,
			wantDesc:  "3m 27s",
		},
		{
			name: "Running with Spinning and plan mode (ctrl+c to interrupt without time)",
			content: `⏺ Understanding the feature request. First, let me check the documentation and current implementation.

  Explore(Explore handler implementation)
  ⎿  Found 5 files
     Found 18 files
     Read(packages/sdk/src/cli/apply/services/handler.ts)
     +27 more tool uses (ctrl+o to expand)

⏺ Fetch(https://example.com/docs/guides/handlers)
  ⎿  Received 51.6KB (200 OK)
     ctrl+b ctrl+b (twice) to run in background

✻ Spinning… (ctrl+c to interrupt)
  ⎿  Tip: Type 'ultrathink' in your message to enable thinking for just that turn

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  ⏸ plan mode on (shift+tab to cycle)`,
			wantState: StateRunning,
			wantMode:  ModePlan,
			wantDesc:  "",
		},
		{
			name: "Running with Crystallizing and accept edits mode",
			content: `✶ Crystallizing… (esc to interrupt · ctrl+t to hide tasks · 52s · ↓ 574 tokens)
  ⎿  ✔ Add dependency to go.mod
     ◻ Add configuration setting to config.go
     ◻ Update library usage in parser.go
     ◻ Pass format option to handler
     ◻ Add and update tests
     ◻ Run go test to verify

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  ⏵⏵ accept edits on (shift+Tab to cycle)`,
			wantState: StateRunning,
			wantMode:  ModeAcceptEdits,
			wantDesc:  "52s",
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
