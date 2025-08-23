/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/cfg"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vstr",
	Short: "VSCode Terminal Runner - Automate your development workflow",
	Long: `VSCode Terminal Runner is a powerful CLI tool that eliminates the pain of 
manually setting up your development environment. With configurable tasks and 
workspaces, you can launch all your project terminals and commands through VSCode.

Perfect for developers working with microservices, full-stack applications, 
or any multi-project setup.

Examples:
	vstr task create              # Create a new task interactively
	vstr task run my-backend      # Run a specific task
	vstr workspace create         # Create a new workspace
	vstr workspace run my-project # Launch all workspace tasks in VSCode`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior - show help when no subcommand is provided
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(cfg.SetupCMD)
}
