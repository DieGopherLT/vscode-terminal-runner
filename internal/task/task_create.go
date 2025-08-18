package task

import (
	"os"
	"path"
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
		Cmds:      strings.Split(t.inputs[cmdsField].Value(), ","),
		Icon:      t.inputs[iconField].Value(),
		IconColor: t.inputs[iconColorField].Value(),
	}
}

// saveTask saves a task to the local configuration file.
func (t TaskModel) saveTask(task models.Task) error {
	if t.isEditMode {
		return repository.UpdateTask(t.originalTaskName, task)
	}
	return repository.SaveTask(task)
}

func (t *TaskModel) isValidTask(task models.Task) bool {

	if strings.TrimSpace(task.Name) == "" {
		t.messages.AddError("Name is required")
	}

	p := strings.TrimSpace(task.Path)
	if p != "" {
		// Expand ~ to home directory for validation
		expandedPath := t.expandPathForValidation(p)
		
		if strings.HasSuffix(p, ".") {
			relativePath := path.Join(os.Getenv("PWD"), expandedPath)
			if _, err := os.Stat(relativePath); os.IsNotExist(err) {
				t.messages.AddError("Path does not exist")
			}
		} else {
			if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
				t.messages.AddError("Path does not exist")
			}
		}
	}

	if len(task.Cmds) == 0 || (len(task.Cmds) == 1 && strings.TrimSpace(task.Cmds[0]) == "") {
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
