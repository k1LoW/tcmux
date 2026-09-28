package agent

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type processEntry struct {
	pid  int
	ppid int
	comm string
}

// DescendantCommands returns command names of all descendant processes of the given pid.
func DescendantCommands(ctx context.Context, pid string) ([]string, error) {
	entries, err := getProcessList(ctx)
	if err != nil {
		return nil, err
	}
	return descendantCommands(pid, entries), nil
}

func getProcessList(ctx context.Context) ([]processEntry, error) {
	out, err := exec.CommandContext(ctx, "ps", "-ax", "-o", "pid=,ppid=,comm=").Output()
	if err != nil {
		return nil, err
	}

	var entries []processEntry
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		comm := filepath.Base(fields[2])
		entries = append(entries, processEntry{pid: pid, ppid: ppid, comm: comm})
	}
	return entries, nil
}

// descendantCommands traverses the process tree via BFS from the given pid
// and returns the command names of all descendant processes.
func descendantCommands(pidStr string, entries []processEntry) []string {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return nil
	}

	// Build parent -> children map
	children := make(map[int][]int)
	commByPID := make(map[int]string)
	for _, e := range entries {
		children[e.ppid] = append(children[e.ppid], e.pid)
		commByPID[e.pid] = e.comm
	}

	var result []string
	queue := children[pid]
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if comm, ok := commByPID[curr]; ok {
			result = append(result, comm)
		}
		queue = append(queue, children[curr]...)
	}
	return result
}
