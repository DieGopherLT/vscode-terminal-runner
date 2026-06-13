package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the installed vstr version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", rootCmd.Name(), rootCmd.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
