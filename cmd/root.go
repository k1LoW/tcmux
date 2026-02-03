package cmd

import (
	"os"

	"github.com/k1LoW/tcmux/output"
	"github.com/spf13/cobra"
)

var colorMode string

var rootCmd = &cobra.Command{
	Use:   "tcmux",
	Short: "terminal and coding agent mux viewer",
	Long:  `tcmux is a terminal and coding agent mux viewer (supports Claude Code, Copilot CLI).`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return output.SetColorMode(colorMode)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&colorMode, "color", "auto", "When to use colors: always, never, or auto")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
