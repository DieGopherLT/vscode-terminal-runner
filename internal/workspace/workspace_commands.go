package workspace

import (
	"fmt"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/vscode"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/spf13/cobra"
)

// RunCmd runs a workspace by name
var RunCmd = &cobra.Command{
	Use:               "run <name>",
	Short:             "Run a workspace",
	Long:              `Execute all tasks defined in a workspace`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeWorkspaceNames,
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]

		runner, err := vscode.NewRunner()
		if err != nil {
			styles.PrintError(fmt.Sprintf("Failed to connect to secure VSCode: %v", err))
			return
		}

		if err := runner.RunWorkspace(workspaceName); err != nil {
			styles.PrintError(fmt.Sprintf("Error running workspace: %v", err))
			return
		}
	},
}

// listWorkspacesCmd lists all saved workspaces
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  `Display a list of all configured workspaces`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement workspace listing
		styles.PrintInfo("Workspace listing not yet implemented")
	},
}

// createWorkspaceCmd creates a new workspace
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new workspace",
	Long:  `Create a new workspace with selected tasks`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := CreateWorkspaceCommand(); err != nil {
			styles.PrintError(fmt.Sprintf("Failed to create workspace: %v", err))
		}
	},
}

// EditCmd opens the TUI form to edit an existing workspace.
var EditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit an existing workspace",
	Long:  `Edit an existing workspace with the specified name`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]

		if err := EditWorkspaceCommand(workspaceName); err != nil {
			styles.PrintError(fmt.Sprintf("Failed to edit workspace '%s': %v", workspaceName, err))
		}
	},
}

// DeleteCmd deletes a workspace specified by name.
var DeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a workspace",
	Long:  `Delete a workspace with the specified name`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]

		if _, err := repository.FindWorkspaceByName(workspaceName); err != nil {
			styles.PrintError(fmt.Sprintf("Workspace '%s' not found", workspaceName))
			return
		}

		if err := repository.DeleteWorkspace(workspaceName); err != nil {
			styles.PrintError(fmt.Sprintf("Failed to delete workspace '%s': %v", workspaceName, err))
			return
		}

		styles.PrintSuccess(fmt.Sprintf("Workspace '%s' deleted successfully!", workspaceName))
	},
}
