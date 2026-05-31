package agent

import "context"

// Type identifies the type of coding agent.
type Type string

const (
	TypeClaude  Type = "claude"
	TypeCopilot Type = "copilot"
)

// Status represents the status of a coding agent instance.
type Status struct {
	State       string // Idle, Running, Waiting, Unknown
	Mode        string // plan mode, accept edits, or empty
	Description string // Additional description (e.g., time elapsed)
}

// Status state constants
const (
	StateIdle    = "Idle"
	StateRunning = "Running"
	StateWaiting = "Waiting" // Agent is waiting for user input/selection
	StateUnknown = "Unknown"
)

// Mode constants
const (
	ModePlan        = "plan mode"
	ModeAcceptEdits = "accept edits"
)

// Detector defines the interface for detecting and parsing coding agents.
type Detector interface {
	Type() Type
	Icon() string
	MayBeTitle(title string) bool
	MayBeProcess(currentCommand string) bool
	ExtractSummary(title string) string
	ParseStatus(content string) Status
}

// All registered detectors
var detectors = []Detector{
	&ClaudeAgent{},
	&CopilotAgent{},
}

// Detect checks if a pane might be running a coding agent.
// Returns the detected agent or nil if no agent is detected.
func Detect(title, currentCommand string) Detector {
	return detectWithChildCmds(title, currentCommand, nil)
}

// DetectFromTree checks if a pane might be running a coding agent,
// traversing the process tree rooted at pid to find descendant processes.
func DetectFromTree(ctx context.Context, title, currentCommand, pid string) Detector {
	if d := Detect(title, currentCommand); d != nil {
		return d
	}
	childCmds, err := DescendantCommands(ctx, pid)
	if err != nil {
		return nil
	}
	return detectWithChildCmds(title, currentCommand, childCmds)
}

// detectWithChildCmds is the internal implementation that also checks child process names.
func detectWithChildCmds(title, currentCommand string, childCmds []string) Detector {
	for _, d := range detectors {
		// Direct detection requires title match to avoid false positives
		if d.MayBeTitle(title) && d.MayBeProcess(currentCommand) {
			return d
		}
		// Child process detection works without title match to support nested terminals
		// (e.g., Claude running inside a neovim terminal). ParseStatus → StateUnknown
		// acts as a safety net for false positives.
		for _, cmd := range childCmds {
			if d.MayBeProcess(cmd) {
				return d
			}
		}
	}
	return nil
}
