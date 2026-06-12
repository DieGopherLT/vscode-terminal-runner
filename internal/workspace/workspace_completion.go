package workspace

import (
	"fmt"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// completeWorkspaceNames provides shell completion for commands that take a
// single workspace name (currently only run). It reads workspace names straight
// from the local repository — never triggering bridge discovery or any network
// call — and returns only the names that start with what the user has typed.
//
// A workspace has no command or path of its own, so the tab-separated
// description lists the tasks it groups. The prefix match is always against the
// workspace name, not the description.
func completeWorkspaceNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// run accepts exactly one name, so once the first positional argument is
	// present there is nothing left to complete.
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	workspaces, err := repository.ReadWorkspaces()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := make([]string, 0, len(workspaces))
	for _, workspace := range workspaces {
		if !strings.HasPrefix(workspace.Name, toComplete) {
			continue
		}
		taskNames := lo.Map(workspace.Tasks, func(task models.Task, _ int) string {
			return task.Name
		})
		description := strings.Join(taskNames, ", ")
		completions = append(completions, fmt.Sprintf("%s\t%s", workspace.Name, description))
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
