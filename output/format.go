package output

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/k1LoW/tcmux/agent"
	"github.com/muesli/termenv"
)

// tcmux custom format variables
const (
	VarAgentStatus = "agent_status" // Coding agent status (context-dependent output)
)

var (
	// Pattern to match format variables: #{variable_name}
	formatVarPattern = regexp.MustCompile(`#\{([^}]+)\}`)

	// tcmux custom variables
	tcmuxVars = map[string]bool{
		VarAgentStatus: true,
	}
)

// AgentInfo holds info for a single coding agent instance.
type AgentInfo struct {
	AgentType agent.Type
	Icon      string
	Summary   string
	Status    agent.Status
}

// FormatContext holds data for format expansion.
type FormatContext struct {
	// tmux variables (from tmux -F output)
	TmuxVars map[string]string

	// Coding agent instances in the window
	AgentInstances []AgentInfo
}

// SessionFormatContext holds data for session format expansion.
type SessionFormatContext struct {
	// tmux variables
	TmuxVars map[string]string

	// Coding agent stats
	IdleCount    int
	RunningCount int
	WaitingCount int
}

// ExtractTmuxVars extracts tmux variable names from a format string.
// Returns only tmux variables (excludes tcmux custom variables).
func ExtractTmuxVars(format string) []string {
	matches := formatVarPattern.FindAllStringSubmatch(format, -1)
	seen := make(map[string]bool)
	var vars []string

	for _, m := range matches {
		varName := m[1]
		// Skip tcmux custom variables
		if tcmuxVars[varName] {
			continue
		}
		// Skip already seen
		if seen[varName] {
			continue
		}
		seen[varName] = true
		vars = append(vars, varName)
	}

	return vars
}

// ExpandFormat expands a format string with the given context.
func ExpandFormat(format string, ctx *FormatContext) string {
	result := format

	// Expand tcmux custom variables
	result = formatVarPattern.ReplaceAllStringFunc(result, func(match string) string {
		// Extract variable name from #{var_name}
		varName := match[2 : len(match)-1]

		switch varName {
		case VarAgentStatus:
			return formatAgentStatus(ctx.AgentInstances)
		default:
			// tmux variable - use value from TmuxVars
			if val, ok := ctx.TmuxVars[varName]; ok {
				return val
			}
			return match // Keep original if not found
		}
	})

	return result
}

// ExpandSessionFormat expands a format string for sessions.
func ExpandSessionFormat(format string, ctx *SessionFormatContext) string {
	result := format

	result = formatVarPattern.ReplaceAllStringFunc(result, func(match string) string {
		varName := match[2 : len(match)-1]

		switch varName {
		case VarAgentStatus:
			return formatAgentStats(ctx.IdleCount, ctx.RunningCount, ctx.WaitingCount)
		default:
			// tmux variable
			if val, ok := ctx.TmuxVars[varName]; ok {
				return val
			}
			return match
		}
	})

	return result
}

// formatAgentStatus formats the full coding agent status string for multiple instances.
// Format: "✻ summary [Status (description, mode)], ⬢ summary2 [Status2]"
// Returns empty string if no coding agent instances.
func formatAgentStatus(instances []AgentInfo) string {
	if len(instances) == 0 {
		return ""
	}

	var instanceParts []string
	for _, inst := range instances {
		if inst.Status.State == "" || inst.Status.State == agent.StateUnknown {
			continue
		}

		// Get the color for the state and agent theme
		var stateColor termenv.Color
		var themeColor termenv.Color
		switch inst.Status.State {
		case agent.StateIdle:
			stateColor = idleColor
		case agent.StateRunning:
			stateColor = runningColor
		case agent.StateWaiting:
			stateColor = waitingColor
		default:
			stateColor = unknownColor
		}

		switch inst.AgentType {
		case agent.TypeClaude:
			themeColor = claudeThemeColor
		case agent.TypeCopilot:
			themeColor = copilotThemeColor
		default:
			themeColor = claudeThemeColor
		}

		// Build the status string with colors
		coloredState := output.String(inst.Status.State).Foreground(stateColor).String()

		var extras []string
		if inst.Status.Description != "" {
			extras = append(extras, inst.Status.Description)
		}
		if inst.Status.Mode != "" {
			coloredMode := output.String(inst.Status.Mode).Foreground(modeColor).String()
			extras = append(extras, coloredMode)
		}

		var statusPart string
		if len(extras) > 0 {
			statusPart = fmt.Sprintf("%s (%s)", coloredState, strings.Join(extras, ", "))
		} else {
			statusPart = coloredState
		}

		// Build the info string for this instance
		var parts []string
		if inst.Summary != "" {
			parts = append(parts, inst.Summary)
		}
		parts = append(parts, fmt.Sprintf("[%s]", statusPart))

		separator := output.String(inst.Icon).Foreground(themeColor).String()
		instanceParts = append(instanceParts, separator+" "+strings.Join(parts, " "))
	}

	if len(instanceParts) == 0 {
		return ""
	}

	return strings.Join(instanceParts, ", ")
}

// formatAgentStats formats coding agent statistics for a session.
func formatAgentStats(idle, running, waiting int) string {
	total := idle + running + waiting
	if total == 0 {
		return ""
	}

	var parts []string
	if idle > 0 {
		colored := output.String(fmt.Sprintf("%d Idle", idle)).Foreground(idleColor).String()
		parts = append(parts, colored)
	}
	if running > 0 {
		colored := output.String(fmt.Sprintf("%d Running", running)).Foreground(runningColor).String()
		parts = append(parts, colored)
	}
	if waiting > 0 {
		colored := output.String(fmt.Sprintf("%d Waiting", waiting)).Foreground(waitingColor).String()
		parts = append(parts, colored)
	}

	return strings.Join(parts, ", ")
}

// BuildTmuxFormat builds a tmux -F format string that includes all required variables.
// It combines user-requested tmux variables with internally required variables.
func BuildTmuxFormat(userVars []string, internalVars []string) string {
	seen := make(map[string]bool)
	var allVars []string

	// Add user variables first
	for _, v := range userVars {
		if !seen[v] {
			seen[v] = true
			allVars = append(allVars, v)
		}
	}

	// Add internal variables
	for _, v := range internalVars {
		if !seen[v] {
			seen[v] = true
			allVars = append(allVars, v)
		}
	}

	// Build format string with tab separator
	var parts []string
	for _, v := range allVars {
		parts = append(parts, fmt.Sprintf("#{%s}", v))
	}

	return strings.Join(parts, "\t")
}

// ParseTmux parses tmux output into a map of variable name to value.
func ParseTmux(line string, vars []string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(line, "\t")

	for i, v := range vars {
		if i < len(parts) {
			result[v] = parts[i]
		}
	}

	return result
}
