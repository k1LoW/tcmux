package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/k1LoW/tcmux/agent"
	"github.com/k1LoW/tcmux/output"
	"github.com/k1LoW/tcmux/tmux"
	"github.com/spf13/cobra"
)

const defaultSessionFormat = "#{session_name}: #{session_windows} windows#{?session_attached, (attached),} #{agent_status}"

var lsFormat string

var lsCmd = &cobra.Command{
	Use:     "list-sessions",
	Aliases: []string{"ls"},
	Short:   "List tmux sessions with coding agent status",
	Long:    `List tmux sessions with coding agent status (Claude Code, Copilot CLI).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use format string if specified, otherwise use default
		format := lsFormat
		if format == "" {
			format = defaultSessionFormat
		}

		// Extract tmux variables from format
		userVars := output.ExtractTmuxVars(format)

		// Build combined variable list
		allVars := mergeVars(userVars, tmux.InternalSessionVars)

		// Build tmux format string
		tmuxFormat := buildTmuxFormat(allVars)

		ctx := cmd.Context()
		sessions, err := tmux.ListSessions(ctx, tmuxFormat, allVars)
		if err != nil {
			return fmt.Errorf("failed to list tmux sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No tmux sessions found.")
			return nil
		}

		// Get all panes to count coding agent instances per session
		paneFormat := buildTmuxFormat(tmux.InternalPaneVars)
		panes, err := tmux.ListPanes(ctx, paneFormat, tmux.InternalPaneVars, tmux.ListPanesOptions{AllSessions: true})
		if err != nil {
			return fmt.Errorf("failed to list tmux panes: %w", err)
		}

		// Build session stats
		sessionStats := make(map[string]*output.SessionFormatContext)
		for _, session := range sessions {
			sessionName := session.Vars["session_name"]
			sessionStats[sessionName] = &output.SessionFormatContext{
				TmuxVars: session.Vars,
			}
		}

		// Count coding agent instances per session
		for _, pane := range panes {
			stats, ok := sessionStats[pane.Vars["session_name"]]
			if !ok {
				continue
			}

			detectedAgent := agent.Detect(pane.Vars["pane_title"], pane.Vars["pane_current_command"])
			if detectedAgent == nil {
				continue
			}

			content, err := tmux.CapturePane(ctx, pane.Vars["pane_id"])
			if err != nil {
				continue
			}

			status := detectedAgent.ParseStatus(content)
			if status.State == agent.StateUnknown {
				continue
			}

			switch status.State {
			case agent.StateIdle:
				stats.IdleCount++
			case agent.StateRunning:
				stats.RunningCount++
			case agent.StateWaiting:
				stats.WaitingCount++
			}
		}

		// Output formatted sessions
		for _, session := range sessions {
			sessionName := session.Vars["session_name"]
			stats := sessionStats[sessionName]

			line := output.ExpandSessionFormat(format, stats)
			// Handle conditional: #{?session_attached, (attached),}
			line = expandConditional(line, session.Vars)
			// Trim trailing whitespace
			line = strings.TrimRight(line, " ")
			fmt.Println(line)
		}

		return nil
	},
}

func init() {
	lsCmd.Flags().StringVarP(&lsFormat, "format", "F", "", "Specify output format (tmux-compatible with tcmux extensions)")
	rootCmd.AddCommand(lsCmd)
}

// expandConditional expands simple tmux conditionals like #{?var,true,false}
func expandConditional(format string, vars map[string]string) string {
	// Simple conditional pattern: #{?var,true_value,false_value}
	re := regexp.MustCompile(`#\{\?([^,]+),([^,]*),([^}]*)\}`)

	return re.ReplaceAllStringFunc(format, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}

		varName := parts[1]
		trueVal := parts[2]
		falseVal := parts[3]

		// Check if variable is truthy
		val, ok := vars[varName]
		if ok && val != "" && val != "0" {
			return trueVal
		}
		return falseVal
	})
}
