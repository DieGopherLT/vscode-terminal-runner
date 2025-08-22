package workspace

import (
	"fmt"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/vscode"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/spf13/cobra"
)

// RunCmd runs a workspace by name
var RunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run a workspace",
	Long:  `Execute all tasks defined in a workspace`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]

		runner, err := vscode.NewSecureRunner()
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
		fmt.Println("Workspace listing not yet implemented")
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
