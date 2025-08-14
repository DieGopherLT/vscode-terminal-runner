package vscode

import (
	"fmt"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
)

// Runner orchestrates the execution of tasks in VSCode terminals via bridge
type Runner struct {
	client *BridgeClient
}

// NewRunner creates a new runner instance connected to VSCode bridge
func NewRunner() (*Runner, error) {
	// Discover the bridge
	bridgeInfo, err := DiscoverBridge()
	if err != nil {
		return nil, fmt.Errorf("VSCode bridge not found. Please ensure:\n1. VSCode is running\n2. VSCR Bridge extension is installed\n3. The extension is active\n\nError: %w", err)
	}
	
	styles.PrintInfo(fmt.Sprintf("Connected to VSCode bridge on port %d", bridgeInfo.Port))
	styles.PrintInfo(fmt.Sprintf("Workspace: %s", bridgeInfo.WorkspaceName))
	
	// Create client
	client := NewBridgeClient(bridgeInfo.Port)
	
	// Verify connection
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to bridge: %w", err)
	}
	
	return &Runner{client: client}, nil
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
	
	// Send to bridge
	if err := r.client.ExecuteTask(*t); err != nil {
		return fmt.Errorf("failed to launch terminal: %w", err)
	}
	
	styles.PrintSuccess(fmt.Sprintf("✓ Terminal '%s' launched successfully", t.Name))
	return nil
}

// RunWorkspace executes all tasks in a workspace
func (r *Runner) RunWorkspace(workspaceName string) error {
	// Load workspace from repository
	workspace, err := repository.FindWorkspaceByName(workspaceName)
	if err != nil {
		return fmt.Errorf("workspace not found: %w", err)
	}
	
	if len(workspace.Tasks) == 0 {
		return fmt.Errorf("no tasks found in workspace '%s'", workspaceName)
	}
	
	// Display workspace info
	r.displayWorkspaceInfo(workspace.Name, workspace.Tasks)
	
	styles.PrintProgress(fmt.Sprintf("Launching %d terminals...", len(workspace.Tasks)))
	
	// Send to bridge
	if err := r.client.ExecuteWorkspace(*workspace); err != nil {
		return fmt.Errorf("failed to execute workspace: %w", err)
	}
	
	styles.PrintSuccess("✓ All terminals launched successfully")
	return nil
}

// displayTaskInfo shows task details before launching
func (r *Runner) displayTaskInfo(t *models.Task) {
	fmt.Println(styles.RunnerHeaderStyle.Render("TASK DETAILS"))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Name: %s %s", t.Icon, t.Name)))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Path: %s", t.Path)))
	
	if len(t.Cmds) > 0 {
		fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Commands: %s", t.Cmds[0])))
		for i := 1; i < len(t.Cmds); i++ {
			fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("          %s", t.Cmds[i])))
		}
	}
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