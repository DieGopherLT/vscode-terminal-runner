package cmd

import (
	"github.com/DieGopherLT/vscode-terminal-runner/internal/workspace"
	"github.com/spf13/cobra"
)

// workspaceCmd represents the workspace command
var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage and run workspaces",
	Long:  `Workspaces allow you to run multiple tasks together in VSCode terminals`,
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
	
	workspaceCmd.AddCommand(workspace.CreateCmd)
	workspaceCmd.AddCommand(workspace.ListCmd)
	workspaceCmd.AddCommand(workspace.RunCmd)
}