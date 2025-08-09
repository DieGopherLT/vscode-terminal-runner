package vscode

import (
	"fmt"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// VSCodeInstance represents a running VSCode process
type VSCodeInstance struct {
	PID         int32
	Name        string
	CommandLine string
	IsInsiders  bool
}

// DetectParentVSCode finds the VSCode instance that spawned the current terminal
func DetectParentVSCode() (*VSCodeInstance, error) {
	// Get parent process ID directly from OS
	ppid := int32(os.Getppid())

	// Walk up the process tree to find VSCode
	currentPID := ppid
	for range 10 { // Limit iterations to prevent infinite loop
		proc, err := process.NewProcess(currentPID)
		if err != nil {
			return nil, fmt.Errorf("failed to get process %d: %w", currentPID, err)
		}

		name, err := proc.Name()
		if err != nil {
			return nil, fmt.Errorf("failed to get process name: %w", err)
		}

		// Check if this is a VSCode process
		if isVSCodeProcess(name) {
			cmdline, _ := proc.Cmdline()
			return &VSCodeInstance{
				PID:         currentPID,
				Name:        name,
				CommandLine: cmdline,
				IsInsiders:  strings.Contains(name, "insiders"),
			}, nil
		}

		// Get parent process
		parent, err := proc.Parent()
		if err != nil || parent == nil {
			break
		}
		currentPID = parent.Pid
	}

	return nil, fmt.Errorf("VSCode parent process not found")
}

// ListRunningVSCodeInstances returns all running VSCode instances
func ListRunningVSCodeInstances() ([]*VSCodeInstance, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	var instances []*VSCodeInstance
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}

		if isVSCodeProcess(name) {
			cmdline, _ := proc.Cmdline()
			instances = append(instances, &VSCodeInstance{
				PID:         proc.Pid,
				Name:        name,
				CommandLine: cmdline,
				IsInsiders:  strings.Contains(name, "insiders"),
			})
		}
	}

	return instances, nil
}

// isVSCodeProcess checks if a process name matches VSCode
func isVSCodeProcess(name string) bool {
	lowerName := strings.ToLower(name)
	return strings.Contains(lowerName, "code") || strings.Contains(lowerName, "code-insiders")
}

// GetWorkspacePath extracts the workspace path from VSCode instance
func (v *VSCodeInstance) GetWorkspacePath() string {
	// Parse command line to find workspace path
	parts := strings.Split(v.CommandLine, " ")
	for i, part := range parts {
		if part == "--folder-uri" && i+1 < len(parts) {
			uri := parts[i+1]
			// Remove file:// prefix if present
			return strings.TrimPrefix(uri, "file://")
		}
	}
	return ""
}
