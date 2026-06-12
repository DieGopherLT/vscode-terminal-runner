package task

import (
	"fmt"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/spf13/cobra"
)

// completeTaskNames provides shell completion for commands that take a single
// task name (run, delete, edit). It reads task names straight from the local
// repository — never triggering bridge discovery or any network call — and
// returns only the names that start with what the user has typed so far.
//
// Each candidate carries a tab-separated description (the task's commands) so
// shells like fish and zsh render it alongside the name. The name alone is the
// completion value; the prefix match is always against the name, not the
// description.
func completeTaskNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// These commands accept exactly one name, so once the first positional
	// argument is present there is nothing left to complete.
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	tasks, err := repository.ReadTasks()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if !strings.HasPrefix(task.Name, toComplete) {
			continue
		}
		description := strings.Join(task.Cmds, " && ")
		completions = append(completions, fmt.Sprintf("%s\t%s", task.Name, description))
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
