package cfg

import "github.com/spf13/cobra"

var SetupCMD = &cobra.Command{
	Use:   "setup",
	Short: "Setup the CLI tool and install required VSCode extension",
	Long:  `Setup the CLI tool and install required VSCode extension`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Setup()
	},
}
