package task

import (
	"fmt"
	"os"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/vscode"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// CreateCmd starts the TUI form to create a new task.
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long:  `Create a new task with the specified configuration`,
	Run: func(cmd *cobra.Command, args []string) {

		batchPath, _ := cmd.Flags().GetString("file")
		if batchPath != "" {
			err := repository.SaveFromFile(batchPath) 
			if err != nil {
				styles.PrintError(fmt.Sprintf("Failed to create tasks from file: %v", err))
				os.Exit(1)
			}
			styles.PrintSuccess("Tasks created successfully from file!")

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
	Use:   "delete <name>",
	Short: "Delete a task",
	Long:  `Delete a task with the specified name`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]
		fmt.Println("Deleting task:", taskName)
	},
}

// EditCmd starts the TUI form to edit an existing task.
var EditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit an existing task",
	Long:  `Edit an existing task with the specified name`,
	Args:  cobra.ExactArgs(1),
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
	Use:   "run <name>",
	Short: "Run a task",
	Long:  `Run a task with the specified name`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]

		runner, err := vscode.NewSecureRunner()
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
	
	fileHelpText := "Creates tasks from a JSON file\n\n" +
		"Example JSON format:\n" +
		"[\n" +
		"  {\n" +
		"    \"name\": \"Build Project\",\n" +
		"    \"icon\": \"tools\",\n" +
		"    \"iconColor\": \"terminal.ansiBlue\",\n" +
		"    \"cmds\": [\"npm run build\", \"echo Build completed\"]\n" +
		"  },\n" +
		"  {\n" +
		"    \"name\": \"Run Tests\",\n" +
		"    \"icon\": \"beaker\",\n" +
		"    \"iconColor\": \"terminal.ansiGreen\",\n" +
		"    \"cmds\": [\"npm test\"]\n" +
		"  }\n" +
		"]"
	
	CreateCmd.Flags().StringP("file", "f", "", fileHelpText)
}