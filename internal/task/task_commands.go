package task

import (
	"fmt"
	"os"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/vscode"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// CreateCmd starts the TUI form to create a new task, or imports tasks from JSON when --from-json is set.
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long:  `Create a new task with the specified configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("from-json")
		if source != "" {
			tasks, err := repository.ReadJSONSource[[]models.Task](source)
			if err != nil {
				styles.PrintError(fmt.Sprintf("Failed to read source: %v", err))
				os.Exit(1)
			}

			if err := repository.ImportTasks(tasks); err != nil {
				styles.PrintError(fmt.Sprintf("Import failed:\n%v", err))
				os.Exit(1)
			}

			styles.PrintSuccess(fmt.Sprintf("%d task(s) imported successfully!", len(tasks)))
			os.Exit(0)
		}

		p := tea.NewProgram(NewModel())
		if _, err := p.Run(); err != nil {
			os.Exit(1)
		}
	},
}

// ListCmd displays the list of configured tasks.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  `Display a list of all configured tasks`,
	Run: func(cmd *cobra.Command, args []string) {
		onlyNames, _ := cmd.Flags().GetBool("only-names")

		if onlyNames {
			err := listAllTaskNames()
			if err != nil {
				fmt.Println("Error listing task names:", err)
			}
			return
		}

		if err := listAllTasks(); err != nil {
			fmt.Println("Error listing tasks:", err)
		}
	},
}

// DeleteCmd deletes a task specified by name.
var DeleteCmd = &cobra.Command{
	Use:               "delete <name>",
	Short:             "Delete a task",
	Long:              `Delete a task with the specified name`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskNames,
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]

		if _, err := repository.FindTaskByName(taskName); err != nil {
			styles.PrintError(fmt.Sprintf("Task '%s' not found", taskName))
			return
		}

		if err := DeleteTask(taskName); err != nil {
			styles.PrintError(fmt.Sprintf("Failed to delete task '%s': %v", taskName, err))
			return
		}

		styles.PrintSuccess(fmt.Sprintf("Task '%s' deleted successfully!", taskName))
	},
}

// EditCmd starts the TUI form to edit an existing task.
var EditCmd = &cobra.Command{
	Use:               "edit <name>",
	Short:             "Edit an existing task",
	Long:              `Edit an existing task with the specified name`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskNames,
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]

		// Find the existing task
		task, err := repository.FindTaskByName(taskName)
		if err != nil {
			styles.PrintError(fmt.Sprintf("Task '%s' not found: %v", taskName, err))
			return
		}

		// Start the edit form with the existing task
		p := tea.NewProgram(NewEditModel(task))
		if _, err := p.Run(); err != nil {
			os.Exit(1)
		}
	},
}

var RunCmd = &cobra.Command{
	Use:               "run <name>",
	Short:             "Run a task",
	Long:              `Run a task with the specified name`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskNames,
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]

		runner, err := vscode.NewRunner()
		if err != nil {
			styles.PrintError(fmt.Sprintf("Failed to create secure runner: %v", err))
			return
		}

		styles.PrintProgress(fmt.Sprintf("Detected secure VSCode instance, proceeding to run task '%s'...", taskName))

		if err := runner.RunTask(taskName); err != nil {
			styles.PrintError(fmt.Sprintf("Error running task: %v", err))
			return
		}
	},
}

func init() {
	ListCmd.Flags().BoolP("only-names", "n", false, "List only task names")
	CreateCmd.Flags().StringP("from-json", "j", "", "Import tasks from a JSON file or stdin (-)")
}
