package output

import (
	"fmt"
	"os"

	"github.com/muesli/termenv"
)

var (
	output *termenv.Output

	// Status colors (Claude Code style)
	idleColor    termenv.Color
	runningColor termenv.Color
	waitingColor termenv.Color
	unknownColor termenv.Color

	// Mode color
	modeColor termenv.Color

	// Claude Code theme color (for branding/separators)
	claudeThemeColor termenv.Color

	// Copilot CLI theme color
	copilotThemeColor termenv.Color
)

func init() {
	initOutput(termenv.NewOutput(os.Stdout, termenv.WithColorCache(true)))
}

func initOutput(o *termenv.Output) {
	output = o
	idleColor = output.Color("#00B359")         // Green
	runningColor = output.Color("#E5A000")      // Orange/Yellow
	waitingColor = output.Color("#5CC8FF")      // Cyan/Light blue - awaiting input
	unknownColor = output.Color("#666666")      // Dark gray
	modeColor = output.Color("#B366FF")         // Purple/Magenta
	claudeThemeColor = output.Color("#E5A000")  // Claude Code orange
	copilotThemeColor = output.Color("#8534F3") // Copilot purple (official brand color)
}

// SetColorMode sets the color output mode: always, never, or auto.
func SetColorMode(mode string) error {
	switch mode {
	case "always":
		initOutput(termenv.NewOutput(os.Stdout, termenv.WithProfile(termenv.TrueColor), termenv.WithColorCache(true)))
	case "never":
		initOutput(termenv.NewOutput(os.Stdout, termenv.WithProfile(termenv.Ascii), termenv.WithColorCache(true)))
	case "auto":
		// Default behavior: detect TTY
		initOutput(termenv.NewOutput(os.Stdout, termenv.WithColorCache(true)))
	default:
		return fmt.Errorf("invalid color mode: %s (must be always, never, or auto)", mode)
	}
	return nil
}
