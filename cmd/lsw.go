package cmd

import (
	"fmt"
	"strings"

	"github.com/k1LoW/tcmux/claude"
	"github.com/k1LoW/tcmux/output"
	"github.com/k1LoW/tcmux/tmux"
	"github.com/spf13/cobra"
)

const defaultWindowFormat = "#{window_index}: #{window_name} (#{window_panes} panes) #{c_status}"

var (
	allWindows  bool
	allSessions bool
	target      string
	lswFormat   string
)

var lswCmd = &cobra.Command{
	Use:     "list-windows",
	Aliases: []string{"lsw"},
	Short:   "List Claude Code instances running in tmux windows",
	Long:    `List Claude Code instances running in tmux windows with their status.`,
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
			tmuxVars        map[string]string
			claudeInstances []output.ClaudeInfo
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

			// Check if this is a Claude Code pane
			title := pane.Vars["pane_title"]
			currentCommand := pane.Vars["pane_current_command"]

			if claude.MayBeTitle(title) && claude.MayBeProcess(currentCommand) {
				// Get Claude Code status
				paneID := pane.Vars["pane_id"]
				content, err := tmux.CapturePane(ctx, paneID)
				if err == nil {
					status := claude.ParseStatus(content)
					if status.State != claude.StateUnknown {
						summary := claude.ExtractSummary(title)
						windows[windowKey].claudeInstances = append(windows[windowKey].claudeInstances, output.ClaudeInfo{
							Summary: summary,
							Status:  status,
						})
					}
				}
			}
		}

		// Build results
		var results []string
		for _, windowKey := range windowOrder {
			win := windows[windowKey]

			// Skip non-Claude Code windows unless -A is specified
			if !allWindows && len(win.claudeInstances) == 0 {
				continue
			}

			// Expand format
			ctx := &output.FormatContext{
				TmuxVars:        win.tmuxVars,
				ClaudeInstances: win.claudeInstances,
			}

			line := output.ExpandFormat(format, ctx)
			// Trim trailing whitespace (in case c_status is empty)
			line = strings.TrimRight(line, " ")
			results = append(results, line)
		}

		if len(results) == 0 {
			if allWindows {
				fmt.Println("No tmux windows found.")
			} else {
				fmt.Println("No Claude Code instances found.")
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
	lswCmd.Flags().BoolVarP(&allWindows, "all-windows", "A", false, "Show all windows, not just Claude Code")
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
