package agent

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
	for _, d := range detectors {
		if d.MayBeTitle(title) && d.MayBeProcess(currentCommand) {
			return d
		}
	}
	return nil
}
