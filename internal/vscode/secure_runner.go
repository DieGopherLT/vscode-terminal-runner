package vscode

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/client"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
)

// SecureRunner orchestrates secure execution of tasks in VSCode terminals via authenticated bridge
type SecureRunner struct {
	client *client.SecureClient
}

// NewSecureRunner creates a new secure runner instance connected to VSCode bridge
func NewSecureRunner() (*SecureRunner, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 1. Discover secure bridge
	bridgeInfo, err := DiscoverSecureBridge()
	if err != nil {
		return nil, fmt.Errorf("VSCode secure bridge not found. Please ensure:\n1. VSCode is running\n2. VSCR Bridge extension is installed and updated\n3. The extension is active and in secure mode\n\nError: %w", err)
	}
	
	styles.PrintInfo(fmt.Sprintf("Found secure bridge on port %d", bridgeInfo.Port))
	styles.PrintInfo(fmt.Sprintf("Workspace: %s", bridgeInfo.WorkspaceName))
	
	// 2. Create secure client
	secureClient := client.NewSecureClient(bridgeInfo.Port)
	
	// 3. Load authentication
	bridgeFilePath := filepath.Join(getBridgeDirectory(), fmt.Sprintf("bridge-%d.json", bridgeInfo.Port))
	if err := secureClient.LoadAuth(bridgeFilePath); err != nil {
		return nil, fmt.Errorf("failed to load authentication: %w", err)
	}
	
	// 4. Test connection and authentication
	if err := secureClient.TestConnection(ctx); err != nil {
		return nil, fmt.Errorf("secure connection test failed: %w", err)
	}
	
	styles.PrintSuccess("✓ Successfully connected to secure bridge")
	
	return &SecureRunner{client: secureClient}, nil
}

// RunTask executes a single task in a new VSCode terminal securely
func (sr *SecureRunner) RunTask(taskName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	// Find the task
	task, err := repository.FindTaskByName(taskName)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	
	styles.PrintProgress(fmt.Sprintf("Launching secure terminal for task '%s'...", task.Name))
	
	// Display task info
	sr.displayTaskInfo(task)
	
	// Send to secure bridge
	if err := sr.client.ExecuteTask(ctx, *task); err != nil {
		return handleSecureError(err)
	}
	
	styles.PrintSuccess(fmt.Sprintf("✓ Secure terminal '%s' launched successfully", task.Name))
	return nil
}

// RunWorkspace executes all tasks in a workspace securely
func (sr *SecureRunner) RunWorkspace(workspaceName string) error {
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
	sr.displayWorkspaceInfo(workspace.Name, workspace.Tasks)
	
	styles.PrintProgress(fmt.Sprintf("Launching %d secure terminals...", len(workspace.Tasks)))
	
	// Send to secure bridge
	if err := sr.client.ExecuteWorkspace(ctx, *workspace); err != nil {
		return handleSecureError(err)
	}
	
	styles.PrintSuccess("✓ All secure terminals launched successfully")
	return nil
}

// displayTaskInfo shows task details before launching
func (sr *SecureRunner) displayTaskInfo(task *models.Task) {
	fmt.Println(styles.RunnerHeaderStyle.Render("SECURE TASK DETAILS"))
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
func (sr *SecureRunner) displayWorkspaceInfo(name string, tasks []models.Task) {
	fmt.Println(styles.RunnerHeaderStyle.Render("SECURE WORKSPACE: " + name))
	fmt.Println(styles.RunnerInfoStyle.Render(fmt.Sprintf("Tasks to launch: %d", len(tasks))))
	fmt.Println()
	
	for _, task := range tasks {
		fmt.Printf("  %s %s\n", task.Icon, styles.RunnerTaskNameStyle.Render(task.Name))
	}
	fmt.Println()
}

