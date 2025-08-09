package vscode

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
)

// TerminalLauncher handles terminal creation in VSCode
type TerminalLauncher struct {
	vsCodeInstance *VSCodeInstance
}

// NewTerminalLauncher creates a new terminal launcher for the given VSCode instance
func NewTerminalLauncher(instance *VSCodeInstance) *TerminalLauncher {
	return &TerminalLauncher{
		vsCodeInstance: instance,
	}
}

// LaunchTerminal creates a new terminal in VSCode with the given task configuration
func (l *TerminalLauncher) LaunchTerminal(task models.Task) error {
	// VSCode CLI command structure
	codeCommand := l.getVSCodeCommand()

	// Build the command to create a new terminal
	args := []string{
		"--new-window=false", // Use existing window
		"--goto", task.Path,  // Navigate to the task path
	}

	// Execute VSCode command
	cmd := exec.Command(codeCommand, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch VSCode terminal: %w", err)
	}

	// Give VSCode time to process the command
	time.Sleep(100 * time.Millisecond)

	// Send terminal configuration via temporary file
	if err := l.configureTerminal(task); err != nil {
		return fmt.Errorf("failed to configure terminal: %w", err)
	}

	return nil
}

// configureTerminal sends terminal configuration to VSCode
func (l *TerminalLauncher) configureTerminal(task models.Task) error {
	// Create terminal profile configuration
	config := map[string]interface{}{
		"name":      task.Name,
		"cwd":       task.Path,
		"icon":      task.Icon,           // VSCode icon ID (e.g., "rocket", "terminal", "cloud")
		"color":     task.IconColor,      // VSCode color ID (e.g., "terminal.ansiRed")
		"env":       map[string]string{}, // Environment variables if needed
		"shellArgs": []string{},
	}

	// If there are commands to execute, add them as initial command
	if len(task.Cmds) > 0 {
		// Join commands with && to execute them sequentially
		initialCommand := strings.Join(task.Cmds, " && ")
		config["initialCommand"] = initialCommand
	}

	// Write config to temporary file
	tmpFile, err := os.CreateTemp("", "vscode-terminal-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if err := json.NewEncoder(tmpFile).Encode(config); err != nil {
		return err
	}
	tmpFile.Close()

	// Use VSCode's command line to create terminal with config
	return l.executeVSCodeCommand("terminal.create", tmpFile.Name())
}

// executeVSCodeCommand runs a VSCode command via CLI
func (l *TerminalLauncher) executeVSCodeCommand(command, configPath string) error {
	codeCommand := l.getVSCodeCommand()

	// Build command arguments
	args := []string{
		"--command", fmt.Sprintf("%s:%s", command, configPath),
	}

	cmd := exec.Command(codeCommand, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("VSCode command failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// getVSCodeCommand returns the appropriate VSCode command based on the instance
func (l *TerminalLauncher) getVSCodeCommand() string {
	if l.vsCodeInstance.IsInsiders {
		return "code-insiders"
	}
	return "code"
}

// LaunchMultipleTerminals launches multiple terminals with a delay between each
func (l *TerminalLauncher) LaunchMultipleTerminals(tasks []models.Task, delay time.Duration) error {
	for i, t := range tasks {
		if err := l.LaunchTerminal(t); err != nil {
			return fmt.Errorf("failed to launch terminal %s: %w", t.Name, err)
		}

		// Add delay between terminals (except for the last one)
		if i < len(tasks)-1 && delay > 0 {
			time.Sleep(delay)
		}
	}
	return nil
}
