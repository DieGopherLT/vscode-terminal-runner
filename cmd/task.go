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
	taskCmd.AddCommand(task.RunCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
