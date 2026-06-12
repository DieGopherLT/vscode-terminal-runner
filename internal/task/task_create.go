package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/samber/lo"
)

// handleTaskCreation builds a Task instance from the form values.
func (t TaskModel) handleTaskCreation() models.Task {
	return models.Task{
		Name:      t.inputs[nameField].Value(),
		Path:      t.inputs[pathField].Value(),
		Cmds:      parseCommands(t.inputs[cmdsField].Value()),
		Icon:      t.inputs[iconField].Value(),
		IconColor: t.inputs[iconColorField].Value(),
	}
}

// parseCommands splits a comma-separated command string, trimming whitespace
// around each command and dropping empty entries so leading spaces never reach
// the terminal (e.g. "yarn install, yarn dev" -> ["yarn install", "yarn dev"]).
func parseCommands(raw string) []string {
	trimmed := lo.Map(strings.Split(raw, ","), func(cmd string, _ int) string {
		return strings.TrimSpace(cmd)
	})
	return lo.Filter(trimmed, func(cmd string, _ int) bool {
		return cmd != ""
	})
}

// saveTask saves a task to the local configuration file.
func (t TaskModel) saveTask(task models.Task) error {
	if t.isEditMode {
		return repository.UpdateTask(t.originalTaskName, task)
	}
	return repository.SaveTask(task)
}

// validateTaskPath checks whether the given path exists on the filesystem.
// It handles tilde expansion and relative paths ending with ".".
// Returns a non-nil error if the path does not exist.
func validateTaskPath(taskPath string, expandFn func(string) string) error {
	p := strings.TrimSpace(taskPath)
	if p == "" {
		return nil
	}

	expandedPath := expandFn(p)

	if strings.HasSuffix(p, ".") {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}
		relativePath := filepath.Join(cwd, expandedPath)
		if _, err := os.Stat(relativePath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist")
		}
		return nil
	}

	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist")
	}
	return nil
}

func (t *TaskModel) isValidTask(task models.Task) bool {
	t.messages.Clear()

	if strings.TrimSpace(task.Name) == "" {
		t.messages.AddError("Name is required")
	}

	if len(task.Cmds) == 0 {
		t.messages.AddError("At least one command is required")
	}

	_, taskIconExists := lo.Find(styles.VSCodeIcons, func(i styles.VSCodeIcon) bool {
		return i.Name == task.Icon
	})
	if task.Icon == "" || !taskIconExists {
		t.messages.AddError("Invalid Icon")
	}

	_, taskColorExists := lo.Find(styles.VSCodeANSIColors, func(c styles.VSCodeANSIColor) bool {
		return c.Name == task.IconColor
	})
	if task.IconColor == "" || !taskColorExists {
		t.messages.AddError("Invalid Icon Color")
	}

	if t.messages.HasErrors() {
		return false
	}

	return true
}

// DeleteTask removes a task from the local configuration file by name.
func DeleteTask(name string) error {
	return repository.DeleteTask(name)
}

// expandPathForValidation expands ~ to home directory for path validation
func (t *TaskModel) expandPathForValidation(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // Return original if can't get home
		}
		if path == "~" {
			return home
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
