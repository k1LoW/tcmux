package agent

import "testing"

func Test_descendantCommands(t *testing.T) {
	tests := []struct {
		name    string
		pid     string
		entries []processEntry
		want    []string
	}{
		{
			name: "direct child",
			pid:  "100",
			entries: []processEntry{
				{pid: 100, ppid: 1, comm: "zsh"},
				{pid: 101, ppid: 100, comm: "claude"},
			},
			want: []string{"claude"},
		},
		{
			name: "nested grandchild",
			pid:  "100",
			entries: []processEntry{
				{pid: 100, ppid: 1, comm: "zsh"},
				{pid: 101, ppid: 100, comm: "bash"},
				{pid: 102, ppid: 101, comm: "node"},
			},
			want: []string{"bash", "node"},
		},
		{
			name: "no children",
			pid:  "100",
			entries: []processEntry{
				{pid: 100, ppid: 1, comm: "zsh"},
				{pid: 200, ppid: 99, comm: "other"},
			},
			want: nil,
		},
		{
			name: "invalid pid",
			pid:  "notanumber",
			entries: []processEntry{
				{pid: 100, ppid: 1, comm: "zsh"},
			},
			want: nil,
		},
		{
			name: "multiple children",
			pid:  "100",
			entries: []processEntry{
				{pid: 100, ppid: 1, comm: "zsh"},
				{pid: 101, ppid: 100, comm: "sh"},
				{pid: 102, ppid: 100, comm: "copilot"},
			},
			want: []string{"sh", "copilot"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := descendantCommands(tt.pid, tt.entries)
			if len(got) != len(tt.want) {
				t.Errorf("descendantCommands(%q) = %v, want %v", tt.pid, got, tt.want)
				return
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("descendantCommands(%q)[%d] = %q, want %q", tt.pid, i, got[i], v)
				}
			}
		})
	}
}
