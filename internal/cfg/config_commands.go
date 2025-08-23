package cfg

import (
	"errors"

	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/spf13/cobra"
)

var SetupCMD = &cobra.Command{
	Use:   "setup",
	Short: "Setup the CLI tool and install required VSCode extension",
	Long:  `Setup the CLI tool and install required VSCode extension`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Setup()
		if errors.Is(err, ErrSetupCompleted) {
			styles.PrintInfo("Setup has already been completed.")
			return nil
		}
		if errors.Is(err, ErrSetupFailed) {
			styles.PrintError("Setup failed. Please try again.")
			return nil
		}
		return err
	},
}
