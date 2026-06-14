package workspace

import (
	"fmt"
	"os"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/charmbracelet/lipgloss"
)

// workspaceNameStyle renders the list header and each workspace name as the
// primary hierarchy level.
var workspaceNameStyle = lipgloss.NewStyle().Foreground(styles.VSCodeBlue).Bold(true)

// lightGrayItalicStyle renders the dimmed "(no tasks)" / "(no commands)"
// placeholders that mark an incomplete workspace or task.
var lightGrayItalicStyle = styles.LightGrayStyle.Italic(true)

// listAllWorkspaces prints every saved workspace as a grouped section: the
// workspace name, then each task it groups with that task's working directory
// and the commands the task runs. The three-level shape (workspace -> task ->
// commands) is rendered with indentation rather than a flat table, which a
// tabwriter cannot express without collapsing the hierarchy.
func listAllWorkspaces() error {
	workspaces, err := repository.ReadWorkspaces()
	if err != nil {
		return err
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	// $HOME cannot change mid-render, so read it once rather than per task.
	home := os.Getenv("HOME")

	var b strings.Builder
	b.WriteString(workspaceNameStyle.Render(fmt.Sprintf("Workspaces (%d)", len(workspaces))))
	b.WriteString("\n\n")

	for _, ws := range workspaces {
		b.WriteString("  " + workspaceNameStyle.Render(ws.Name) + "\n")
		b.WriteString("  " + styles.LightGrayStyle.Render(strings.Repeat("─", 40)) + "\n")

		if len(ws.Tasks) == 0 {
			b.WriteString("    " + lightGrayItalicStyle.Render("(no tasks)") + "\n\n")
			continue
		}

		nameWidth := longestTaskName(ws.Tasks)
		for _, task := range ws.Tasks {
			paddedName := fmt.Sprintf("%-*s", nameWidth, task.Name)
			b.WriteString("    " + styles.FieldLabelStyle.Render(paddedName))
			b.WriteString("  " + styles.LightGrayStyle.Render(abbreviateHome(home, task.Path)) + "\n")

			if len(task.Cmds) == 0 {
				b.WriteString("      " + lightGrayItalicStyle.Render("(no commands)") + "\n")
				continue
			}
			for _, cmd := range task.Cmds {
				b.WriteString("      " + styles.LightGrayStyle.Render("$ ") + cmd + "\n")
			}
		}
		b.WriteString("\n")
	}

	fmt.Print(b.String())
	return nil
}

// listAllWorkspaceNames prints just the workspace names as a bullet list, the
// compact view used by `--only-names` for scripting and quick scanning.
func listAllWorkspaceNames() error {
	workspaces, err := repository.ReadWorkspaces()
	if err != nil {
		return err
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	var b strings.Builder
	b.WriteString("Workspace Names:\n")
	for _, ws := range workspaces {
		b.WriteString("- " + ws.Name + "\n")
	}
	fmt.Print(b.String())
	return nil
}

// longestTaskName returns the byte length of the longest task name in tasks,
// used to align each task's working directory into a single column.
func longestTaskName(tasks []models.Task) int {
	longest := 0
	for _, task := range tasks {
		if len(task.Name) > longest {
			longest = len(task.Name)
		}
	}
	return longest
}

// abbreviateHome replaces a leading $HOME path segment in path with "~",
// matching the path display used by `vstr task list`. The match requires a
// path-component boundary so that, with HOME=/home/user, /home/user/proj
// abbreviates to ~/proj while /home/userdata is left untouched. home is read
// once by the caller.
func abbreviateHome(home, path string) string {
	if home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(os.PathSeparator)) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
