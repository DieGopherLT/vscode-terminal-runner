/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/task"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// taskCmd represents the project command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Interactive task manager",
	Long: `Interactive TUI to manage all task operations including create, list, edit, delete, and run tasks`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(task.NewMenuModel())
		if _, err := p.Run(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	
	taskCmd.AddCommand(task.CreateCmd)
	taskCmd.AddCommand(task.ListCmd)
	taskCmd.AddCommand(task.DeleteCmd)
	taskCmd.AddCommand(task.EditCmd)
	taskCmd.AddCommand(task.RunCmd)
}
