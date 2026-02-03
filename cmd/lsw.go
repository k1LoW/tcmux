package cmd

import (
	"fmt"
	"strings"

	"github.com/k1LoW/tcmux/agent"
	"github.com/k1LoW/tcmux/output"
	"github.com/k1LoW/tcmux/tmux"
	"github.com/spf13/cobra"
)

const defaultWindowFormat = "#{window_index}: #{window_name} (#{window_panes} panes) #{agent_status}"

var (
	allWindows  bool
	allSessions bool
	target      string
	lswFormat   string
)

var lswCmd = &cobra.Command{
	Use:     "list-windows",
	Aliases: []string{"lsw"},
	Short:   "List coding agent instances running in tmux windows",
	Long:    `List coding agent instances (Claude Code, Copilot CLI) running in tmux windows with their status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use format string if specified, otherwise use default
		format := lswFormat
		if format == "" {
			format = defaultWindowFormat
		}

		// Extract tmux variables from format
		userVars := output.ExtractTmuxVars(format)

		// Build combined variable list (user vars + internal vars)
		allVars := mergeVars(userVars, tmux.InternalPaneVars)

		// Build tmux format string
		tmuxFormat := buildTmuxFormat(allVars)

		opts := tmux.ListPanesOptions{
			AllSessions: allSessions,
			Target:      target,
		}

		ctx := cmd.Context()
		panes, err := tmux.ListPanes(ctx, tmuxFormat, allVars, opts)
		if err != nil {
			return fmt.Errorf("failed to list tmux panes: %w", err)
		}

		// Group panes by window
		type windowData struct {
			tmuxVars       map[string]string
			agentInstances []output.AgentInfo
		}
		var windowOrder []string
		windows := make(map[string]*windowData)

		for _, pane := range panes {
			// Create window key
			windowKey := fmt.Sprintf("%s:%s", pane.Vars["session_name"], pane.Vars["window_index"])

			// Initialize window data if not seen
			if _, ok := windows[windowKey]; !ok {
				windows[windowKey] = &windowData{
					tmuxVars: pane.Vars,
				}
				windowOrder = append(windowOrder, windowKey)
			}

			// Check if this is a coding agent pane
			title := pane.Vars["pane_title"]
			currentCommand := pane.Vars["pane_current_command"]

			if detectedAgent := agent.Detect(title, currentCommand); detectedAgent != nil {
				// Get coding agent status
				paneID := pane.Vars["pane_id"]
				content, err := tmux.CapturePane(ctx, paneID)
				if err == nil {
					status := detectedAgent.ParseStatus(content)
					if status.State != agent.StateUnknown {
						summary := detectedAgent.ExtractSummary(title)
						windows[windowKey].agentInstances = append(windows[windowKey].agentInstances, output.AgentInfo{
							AgentType: detectedAgent.Type(),
							Icon:      detectedAgent.Icon(),
							Summary:   summary,
							Status:    status,
						})
					}
				}
			}
		}

		// Build results
		var results []string
		for _, windowKey := range windowOrder {
			win := windows[windowKey]

			// Skip non-agent windows unless -A is specified
			if !allWindows && len(win.agentInstances) == 0 {
				continue
			}

			// Expand format
			ctx := &output.FormatContext{
				TmuxVars:       win.tmuxVars,
				AgentInstances: win.agentInstances,
			}

			line := output.ExpandFormat(format, ctx)
			// Trim trailing whitespace (in case agent_status is empty)
			line = strings.TrimRight(line, " ")
			results = append(results, line)
		}

		if len(results) == 0 {
			if allWindows {
				fmt.Println("No tmux windows found.")
			} else {
				fmt.Println("No coding agent instances found.")
			}
			return nil
		}

		for _, line := range results {
			fmt.Println(line)
		}

		return nil
	},
}

func init() {
	lswCmd.Flags().BoolVarP(&allWindows, "all-windows", "A", false, "Show all windows, not just coding agents")
	lswCmd.Flags().BoolVarP(&allSessions, "all-sessions", "a", false, "List windows from all sessions")
	lswCmd.Flags().StringVarP(&target, "target-session", "t", "", "Specify target session")
	lswCmd.Flags().StringVarP(&lswFormat, "format", "F", "", "Specify output format (tmux-compatible with tcmux extensions)")
	rootCmd.AddCommand(lswCmd)
}

// mergeVars merges two variable lists, removing duplicates.
func mergeVars(vars1, vars2 []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, v := range vars1 {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	for _, v := range vars2 {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

// buildTmuxFormat builds a tmux -F format string from variable names.
func buildTmuxFormat(vars []string) string {
	var parts []string
	for _, v := range vars {
		parts = append(parts, fmt.Sprintf("#{%s}", v))
	}
	return strings.Join(parts, "\t")
}
