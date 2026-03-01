package cmd

import (
	"fmt"

	"github.com/k1LoW/tcmux/agent"
	"github.com/k1LoW/tcmux/output"
	"github.com/k1LoW/tcmux/tmux"
	"github.com/spf13/cobra"
)

const defaultStatsFormat = "I:#{total_idle} R:#{total_running} W:#{total_waiting}"

var statsFormat string

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show total coding agent stats across all sessions",
	Long:  `Show aggregated coding agent statistics (Claude Code, Copilot CLI) across all tmux sessions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Get all panes
		paneFormat := buildTmuxFormat(tmux.InternalPaneVars)
		panes, err := tmux.ListPanes(ctx, paneFormat, tmux.InternalPaneVars, tmux.ListPanesOptions{AllSessions: true})
		if err != nil {
			return fmt.Errorf("failed to list tmux panes: %w", err)
		}

		// Count agent states
		var totalStats output.TotalStatsContext
		for _, pane := range panes {
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
				totalStats.IdleCount++
			case agent.StateRunning:
				totalStats.RunningCount++
			case agent.StateWaiting:
				totalStats.WaitingCount++
			}
		}

		// Output
		format := statsFormat
		if format == "" {
			format = defaultStatsFormat
		}

		line := output.ExpandStatsFormat(format, &totalStats)
		fmt.Println(line)

		return nil
	},
}

func init() {
	statsCmd.Flags().StringVarP(&statsFormat, "format", "F", "", "Specify output format (use #{total_idle}, #{total_running}, #{total_waiting}, #{total_agents}, #{agent_status})")
	rootCmd.AddCommand(statsCmd)
}
