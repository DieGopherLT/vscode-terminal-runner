package vscode

import (
	"context"
	"fmt"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/client"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
)

// Runner orchestrates execution of tasks in VSCode terminals via the authenticated bridge
type Runner struct {
	client *client.Client
}

// NewRunner creates a new runner instance connected to the VSCode bridge
func NewRunner() (*Runner, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Discover bridge
	bridgeInfo, err := DiscoverBridge()
	if err != nil {
		return nil, fmt.Errorf("VSCode bridge not found. Please ensure:\n1. VSCode is running\n2. VSCR Bridge extension is installed and updated\n3. The extension is active and in secure mode\n\nError: %w", err)
	}

	styles.PrintInfo(fmt.Sprintf("Found bridge on port %d", bridgeInfo.Port))
	styles.PrintInfo(fmt.Sprintf("Workspace: %s", bridgeInfo.WorkspaceName))

	// 2. Create client
	bridgeClient := client.NewClient(bridgeInfo.Port)

	// 3. Load authentication from the token validated during discovery
	if err := bridgeClient.LoadAuthFromToken(bridgeInfo.AuthToken); err != nil {
		return nil, fmt.Errorf("failed to load authentication: %w", err)
	}

	// 4. Test connection and authentication
	if err := bridgeClient.TestConnection(ctx); err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	styles.PrintSuccess("Connected to bridge")

	return &Runner{client: bridgeClient}, nil
}

// RunTask executes a single task in a new VSCode terminal
func (r *Runner) RunTask(taskName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Find the task
	task, err := repository.FindTaskByName(taskName)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	styles.PrintProgress(fmt.Sprintf("Launching terminal for task '%s'...", task.Name))

	// Display task info
	r.displayTaskInfo(task)

	// Send to bridge
	if err := r.client.ExecuteTask(ctx, *task); err != nil {
		return handleBridgeError(err)
	}

	styles.PrintSuccess(fmt.Sprintf("Terminal '%s' launched successfully", task.Name))
	return nil
}

// RunWorkspace executes all tasks in a workspace
func (r *Runner) RunWorkspace(workspaceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

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
	if err := r.client.ExecuteWorkspace(ctx, *workspace); err != nil {
		return handleBridgeError(err)
	}

	styles.PrintSuccess("All terminals launched successfully")
	return nil
}

// displayTaskInfo shows task details before launching
func (r *Runner) displayTaskInfo(task *models.Task) {
	fmt.Println(styles.RunnerHeaderStyle.Render("TASK DETAILS"))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Name: %s %s", task.Icon, task.Name)))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Path: %s", task.Path)))

	if len(task.Cmds) > 0 {
		fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Commands: %s", task.Cmds[0])))
		for i := 1; i < len(task.Cmds); i++ {
			fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("          %s", task.Cmds[i])))
		}
	}
	fmt.Println()
}

// displayWorkspaceInfo shows workspace details before launching
func (r *Runner) displayWorkspaceInfo(name string, tasks []models.Task) {
	fmt.Println(styles.RunnerHeaderStyle.Render("WORKSPACE: " + name))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Tasks to launch: %d", len(tasks))))
	fmt.Println()

	for _, task := range tasks {
		fmt.Printf("  %s %s\n", task.Icon, styles.RunnerTaskNameStyle.Render(task.Name))
	}
	fmt.Println()
}
