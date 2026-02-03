package output

import (
	"testing"

	"github.com/k1LoW/tcmux/agent"
)

func TestExtractTmuxVars(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   []string
	}{
		{
			name:   "Simple tmux variable",
			format: "#{window_index}",
			want:   []string{"window_index"},
		},
		{
			name:   "Multiple tmux variables",
			format: "#{window_index}: #{window_name}",
			want:   []string{"window_index", "window_name"},
		},
		{
			name:   "Exclude tcmux variables",
			format: "#{window_index} #{agent_status}",
			want:   []string{"window_index"},
		},
		{
			name:   "Only tcmux variables",
			format: "#{agent_status} #{agent_status}",
			want:   nil,
		},
		{
			name:   "Duplicate variables",
			format: "#{window_index} #{window_index}",
			want:   []string{"window_index"},
		},
		{
			name:   "No variables",
			format: "plain text",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTmuxVars(tt.format)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractTmuxVars() = %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("ExtractTmuxVars()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestExpandFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		ctx    *FormatContext
		want   string
	}{
		{
			name:   "Expand tmux variable",
			format: "#{window_index}: #{window_name}",
			ctx: &FormatContext{
				TmuxVars: map[string]string{
					"window_index": "0",
					"window_name":  "editor",
				},
			},
			want: "0: editor",
		},
		{
			name:   "Expand agent_status with status only (Claude)",
			format: "#{agent_status}",
			ctx: &FormatContext{
				TmuxVars: map[string]string{},
				AgentInstances: []AgentInfo{
					{AgentType: agent.TypeClaude, Icon: "✻", Summary: "", Status: agent.Status{State: agent.StateIdle}},
				},
			},
			want: "✻ [Idle]",
		},
		{
			name:   "Expand agent_status with summary and status (Claude)",
			format: "#{agent_status}",
			ctx: &FormatContext{
				TmuxVars: map[string]string{},
				AgentInstances: []AgentInfo{
					{AgentType: agent.TypeClaude, Icon: "✻", Summary: "Fix login bug", Status: agent.Status{State: agent.StateIdle}},
				},
			},
			want: "✻ Fix login bug [Idle]",
		},
		{
			name:   "Expand agent_status with full status (Claude)",
			format: "#{agent_status}",
			ctx: &FormatContext{
				TmuxVars: map[string]string{},
				AgentInstances: []AgentInfo{
					{
						AgentType: agent.TypeClaude,
						Icon:      "✻",
						Summary:   "Fix login bug",
						Status: agent.Status{
							State:       agent.StateRunning,
							Description: "1m 30s",
							Mode:        "plan mode",
						},
					},
				},
			},
			want: "✻ Fix login bug [Running (1m 30s, plan mode)]",
		},
		{
			name:   "Empty agent_status when no instances",
			format: "test #{agent_status}",
			ctx: &FormatContext{
				TmuxVars:       map[string]string{},
				AgentInstances: []AgentInfo{},
			},
			want: "test ",
		},
		{
			name:   "Combined format (Claude)",
			format: "#{window_index}: #{window_name} #{agent_status}",
			ctx: &FormatContext{
				TmuxVars: map[string]string{
					"window_index": "0",
					"window_name":  "editor",
				},
				AgentInstances: []AgentInfo{
					{AgentType: agent.TypeClaude, Icon: "✻", Summary: "Fix login bug", Status: agent.Status{State: agent.StateIdle}},
				},
			},
			want: "0: editor ✻ Fix login bug [Idle]",
		},
		{
			name:   "Multiple Claude Code instances",
			format: "#{agent_status}",
			ctx: &FormatContext{
				TmuxVars: map[string]string{},
				AgentInstances: []AgentInfo{
					{AgentType: agent.TypeClaude, Icon: "✻", Summary: "Fix login bug", Status: agent.Status{State: agent.StateIdle}},
					{AgentType: agent.TypeClaude, Icon: "✻", Summary: "Add API endpoint", Status: agent.Status{State: agent.StateRunning}},
				},
			},
			want: "✻ Fix login bug [Idle], ✻ Add API endpoint [Running]",
		},
		{
			name:   "Copilot CLI instance",
			format: "#{agent_status}",
			ctx: &FormatContext{
				TmuxVars: map[string]string{},
				AgentInstances: []AgentInfo{
					{AgentType: agent.TypeCopilot, Icon: "⬢", Summary: "", Status: agent.Status{State: agent.StateRunning}},
				},
			},
			want: "⬢ [Running]",
		},
		{
			name:   "Mixed agents",
			format: "#{agent_status}",
			ctx: &FormatContext{
				TmuxVars: map[string]string{},
				AgentInstances: []AgentInfo{
					{AgentType: agent.TypeClaude, Icon: "✻", Summary: "Fix login bug", Status: agent.Status{State: agent.StateIdle}},
					{AgentType: agent.TypeCopilot, Icon: "⬢", Summary: "", Status: agent.Status{State: agent.StateRunning}},
				},
			},
			want: "✻ Fix login bug [Idle], ⬢ [Running]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandFormat(tt.format, tt.ctx)
			if got != tt.want {
				t.Errorf("ExpandFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpandSessionFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		ctx    *SessionFormatContext
		want   string
	}{
		{
			name:   "Expand agent_status with all states",
			format: "#{agent_status}",
			ctx: &SessionFormatContext{
				TmuxVars:     map[string]string{},
				IdleCount:    2,
				RunningCount: 1,
				WaitingCount: 1,
			},
			want: "2 Idle, 1 Running, 1 Waiting",
		},
		{
			name:   "Expand agent_status with only idle",
			format: "#{agent_status}",
			ctx: &SessionFormatContext{
				TmuxVars:  map[string]string{},
				IdleCount: 3,
			},
			want: "3 Idle",
		},
		{
			name:   "Empty agent_status when no agents",
			format: "#{agent_status}",
			ctx: &SessionFormatContext{
				TmuxVars: map[string]string{},
			},
			want: "",
		},
		{
			name:   "Combined session format",
			format: "#{session_name}: #{session_windows} windows #{agent_status}",
			ctx: &SessionFormatContext{
				TmuxVars: map[string]string{
					"session_name":    "dev",
					"session_windows": "3",
				},
				IdleCount:    2,
				RunningCount: 1,
			},
			want: "dev: 3 windows 2 Idle, 1 Running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandSessionFormat(tt.format, tt.ctx)
			if got != tt.want {
				t.Errorf("ExpandSessionFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}
