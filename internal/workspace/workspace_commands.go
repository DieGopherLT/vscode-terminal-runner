package workspace

import (
	"fmt"
	"os"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
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

// ListCmd lists all saved workspaces with their tasks, working directories and commands.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  `Display all configured workspaces with their tasks, working directories and commands`,
	Run: func(cmd *cobra.Command, args []string) {
		onlyNames, _ := cmd.Flags().GetBool("only-names")

		if onlyNames {
			if err := listAllWorkspaceNames(); err != nil {
				fmt.Println("Error listing workspace names:", err)
			}
			return
		}

		if err := listAllWorkspaces(); err != nil {
			fmt.Println("Error listing workspaces:", err)
		}
	},
}

// createWorkspaceCmd creates a new workspace, or imports workspaces from JSON when --from-json is set.
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new workspace",
	Long:  `Create a new workspace with selected tasks`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("from-json")
		if source != "" {
			workspaces, err := repository.ReadJSONSource[[]models.Workspace](source)
			if err != nil {
				styles.PrintError(fmt.Sprintf("Failed to read source: %v", err))
				os.Exit(1)
			}

			if err := repository.ImportWorkspaces(workspaces); err != nil {
				styles.PrintError(fmt.Sprintf("Import failed:\n%v", err))
				os.Exit(1)
			}

			styles.PrintSuccess(fmt.Sprintf("%d workspace(s) imported successfully!", len(workspaces)))
			os.Exit(0)
		}

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

func init() {
	ListCmd.Flags().BoolP("only-names", "n", false, "List only workspace names")
	CreateCmd.Flags().StringP("from-json", "j", "", "Import workspaces from a JSON file or stdin (-)")
}
