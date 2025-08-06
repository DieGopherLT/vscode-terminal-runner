package task

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// CreateCmd starts the TUI form to create a new task.
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long:  `Create a new task with the specified configuration`,
	Run: func(cmd *cobra.Command, args []string) {
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
		// Lógica para listar tareas
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
