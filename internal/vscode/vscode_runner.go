package vscode

import (
	"fmt"
	"strings"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/charmbracelet/lipgloss"
)

// Runner orchestrates the execution of tasks in VSCode terminals
type Runner struct {
	launcher *TerminalLauncher
}

// NewRunner creates a new runner instance
func NewRunner() (*Runner, error) {
	// First, try to detect parent VSCode instance
	instance, err := DetectParentVSCode()
	if err != nil {
		// If not running in VSCode terminal, list available instances
		return nil, handleNoVSCodeParent()
	}

	styles.PrintInfo(fmt.Sprintf("Detected VSCode instance: %s (PID: %d)", instance.Name, instance.PID))

	launcher := NewTerminalLauncher(instance)
	return &Runner{launcher: launcher}, nil
}

// RunTask executes a single task in a new VSCode terminal
func (r *Runner) RunTask(taskName string) error {
	// Find the task
	t, err := repository.FindTaskByName(taskName)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	styles.PrintProgress(fmt.Sprintf("Launching terminal for task '%s'...", t.Name))

	// Display task info
	r.displayTaskInfo(t)

	// Launch the terminal
	styles.PrintInfo("Launching terminal...")
	if err := r.launcher.LaunchTerminal(*t); err != nil {
		return fmt.Errorf("failed to launch terminal: %w", err)
	}

	styles.PrintSuccess(fmt.Sprintf("✓ Terminal '%s' launched successfully", t.Name))
	return nil
}

// RunWorkspace executes all tasks in a workspace
func (r *Runner) RunWorkspace(workspaceName string, options RunOptions) error {
	// TODO: Implement workspace loading (this would come from a workspace package)
	// For now, we'll simulate with a slice of tasks
	tasks := []models.Task{} // This should be loaded from workspace configuration

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found in workspace '%s'", workspaceName)
	}

	// Display workspace info
	r.displayWorkspaceInfo(workspaceName, tasks)

	// Launch terminals
	if options.Parallel {
		return r.runParallel(tasks)
	}
	return r.runSequential(tasks, options.Delay)
}

// RunOptions contains options for running tasks
type RunOptions struct {
	Delay    time.Duration
	Parallel bool
}

// runSequential launches tasks one by one with a delay
func (r *Runner) runSequential(tasks []models.Task, delay time.Duration) error {
	styles.PrintInfo(fmt.Sprintf("Launching %d terminals sequentially...", len(tasks)))

	for i, t := range tasks {
		styles.PrintProgress(fmt.Sprintf("[%d/%d] Launching '%s'...", i+1, len(tasks), t.Name))

		if err := r.launcher.LaunchTerminal(t); err != nil {
			styles.PrintError(fmt.Sprintf("✗ Failed to launch '%s': %v", t.Name, err))
			continue
		}

		styles.PrintSuccess(fmt.Sprintf("✓ '%s' launched", t.Name))

		if i < len(tasks)-1 && delay > 0 {
			time.Sleep(delay)
		}
	}

	return nil
}

// runParallel launches all tasks at once
func (r *Runner) runParallel(tasks []models.Task) error {
	styles.PrintInfo(fmt.Sprintf("Launching %d terminals in parallel...", len(tasks)))

	errChan := make(chan error, len(tasks))
	for _, t := range tasks {
		go func(task models.Task) {
			if err := r.launcher.LaunchTerminal(task); err != nil {
				errChan <- fmt.Errorf("'%s': %w", task.Name, err)
			} else {
				errChan <- nil
			}
		}(t)
	}

	// Collect results
	var errors []error
	for i := 0; i < len(tasks); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
			styles.PrintError(fmt.Sprintf("✗ %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to launch %d terminals", len(errors))
	}

	styles.PrintSuccess(fmt.Sprintf("✓ All %d terminals launched successfully", len(tasks)))
	return nil
}

// displayTaskInfo shows task details before launching
func (r *Runner) displayTaskInfo(t *models.Task) {
	fmt.Println(styles.RunnerHeaderStyle.Render("TASK DETAILS"))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Name: %s %s", t.Icon, t.Name)))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Path: %s", t.Path)))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Commands: %s", strings.Join(t.Cmds, " && "))))
	fmt.Println()
}

// displayWorkspaceInfo shows workspace details before launching
func (r *Runner) displayWorkspaceInfo(name string, tasks []models.Task) {
	fmt.Println(styles.RunnerHeaderStyle.Render("WORKSPACE: " + name))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Tasks to launch: %d", len(tasks))))
	fmt.Println()

	for _, t := range tasks {
		fmt.Printf("  %s %s\n", t.Icon, styles.RunnerTaskNameStyle.Render(t.Name))
	}
	fmt.Println()
}

// handleNoVSCodeParent handles the case when not running in VSCode
func handleNoVSCodeParent() error {
	styles.PrintWarning("Not running inside a VSCode terminal")

	// List available VSCode instances
	instances, err := ListRunningVSCodeInstances()
	if err != nil {
		return fmt.Errorf("failed to list VSCode instances: %w", err)
	}

	if len(instances) == 0 {
		styles.PrintError("No running VSCode instances found")
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(styles.LightGray).Render(
			"VSCT Runner requires VSCode to be running.\n" +
				"Please open VSCode and run this command from its integrated terminal."))
		return fmt.Errorf("VSCode not running")
	}

	// Show available instances
	fmt.Println(styles.RunnerHeaderStyle.Render("Available VSCode Instances:"))
	for i, instance := range instances {
		workspace := instance.GetWorkspacePath()
		if workspace == "" {
			workspace = "(no workspace)"
		}
		fmt.Printf("%d. PID: %d - %s\n", i+1, instance.PID, workspace)
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(styles.LightGray).Render(
		"Please run this command from within a VSCode integrated terminal."))

	return fmt.Errorf("not in VSCode terminal")
}
