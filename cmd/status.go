package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/cfg"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/client"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/vscode"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/spf13/cobra"
)

type statusCheckResult struct {
	extensionInstalled bool
	bridgePort         int
	bridgeWorkspace    string
	bridgeErr          error
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check VSTR-Bridge extension and bridge connection",
	Long:  `Reports whether the VSTR-Bridge VSCode extension is installed and whether the CLI can reach the active bridge.`,
	Run: func(cmd *cobra.Command, args []string) {
		result := gatherStatus()
		allOK := renderStatus(result)
		if !allOK {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func gatherStatus() statusCheckResult {
	result := statusCheckResult{
		extensionInstalled: cfg.IsExtensionInstalled(),
	}

	bridge, err := vscode.DiscoverBridge()
	if err != nil {
		result.bridgeErr = err
		return result
	}

	bridgeClient := client.NewClient(bridge.Port)
	if err := bridgeClient.LoadAuthFromToken(bridge.AuthToken); err != nil {
		result.bridgeErr = fmt.Errorf("failed to load auth token: %w", err)
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := bridgeClient.TestConnection(ctx); err != nil {
		result.bridgeErr = err
		return result
	}

	result.bridgePort = bridge.Port
	result.bridgeWorkspace = bridge.WorkspaceName
	return result
}

func renderStatus(result statusCheckResult) bool {
	styles.PrintInfo("VSTR status")

	allChecksPass := true

	if result.extensionInstalled {
		styles.PrintSuccess("Extension 'diegopherlt.vstr-bridge' installed")
	} else {
		styles.PrintError("Extension 'diegopherlt.vstr-bridge' not installed (run 'vstr setup')")
		allChecksPass = false
	}

	if result.bridgeErr == nil {
		styles.PrintSuccess(fmt.Sprintf("Bridge connection (port %d, workspace %q)", result.bridgePort, result.bridgeWorkspace))
	} else {
		styles.PrintError(fmt.Sprintf("Bridge connection: %s", result.bridgeErr))
		allChecksPass = false
	}

	return allChecksPass
}
