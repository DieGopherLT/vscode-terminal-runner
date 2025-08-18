/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/task"
	"github.com/spf13/cobra"
)

// taskCmd represents the project command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "TUI to manage projects",
	Long: `Interactive TUI to manage all tasks operations`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Interactive TUI to manage all tasks operations")
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
