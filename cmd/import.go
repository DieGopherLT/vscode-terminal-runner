package cmd

import (
	"fmt"
	"os"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/spf13/cobra"
)

type unifiedImportPayload struct {
	Tasks      []models.Task      `json:"tasks"`
	Workspaces []models.Workspace `json:"workspaces"`
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import tasks and workspaces from a JSON file",
	Long:  `Import tasks and workspaces in bulk from a JSON file or stdin (-)`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("from-json")

		payload, err := repository.ReadJSONSource[unifiedImportPayload](source)
		if err != nil {
			styles.PrintError(fmt.Sprintf("Failed to read source: %v", err))
			os.Exit(1)
		}

		// Validate both arrays fully before writing either (cross-file atomicity).
		taskErr := repository.ValidateTaskBatch(payload.Tasks)
		wsErr := repository.ValidateWorkspaceBatch(payload.Workspaces)
		if taskErr != nil || wsErr != nil {
			if taskErr != nil {
				styles.PrintError(fmt.Sprintf("Task validation errors:\n%v", taskErr))
			}
			if wsErr != nil {
				styles.PrintError(fmt.Sprintf("Workspace validation errors:\n%v", wsErr))
			}
			os.Exit(1)
		}

		if err := repository.ImportTasks(payload.Tasks); err != nil {
			styles.PrintError(fmt.Sprintf("Failed to import tasks: %v", err))
			os.Exit(1)
		}

		if err := repository.ImportWorkspaces(payload.Workspaces); err != nil {
			// Tasks were already written; report partial state explicitly.
			styles.PrintError(fmt.Sprintf("Tasks imported successfully but workspace import failed (partial state): %v", err))
			os.Exit(1)
		}

		styles.PrintSuccess(fmt.Sprintf(
			"Imported %d task(s) and %d workspace(s) successfully!",
			len(payload.Tasks), len(payload.Workspaces),
		))
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringP("from-json", "j", "", "Import from a JSON file or stdin (-)")
	importCmd.MarkFlagRequired("from-json")
}
