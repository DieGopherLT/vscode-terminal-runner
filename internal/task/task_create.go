package task

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/vscode"
	"github.com/samber/lo"
)

var TasksSaveFile = path.Join(os.Getenv("HOME"), ".config/vsct-runner/tasks.json")

// TaskSaveFileContent represents the structure of the task persistence file.
type TaskSaveFileContent struct {
	Tasks []Task `json:"tasks"`
}

// handleTaskCreation builds a Task instance from the form values.
func (t TaskModel) handleTaskCreation() Task {
	return Task{
		Name:      t.inputs[nameField].Value(),
		Path:      t.inputs[pathField].Value(),
		Cmds:      strings.Split(t.inputs[cmdsField].Value(), ","),
		Icon:      t.inputs[iconField].Value(),
		IconColor: t.inputs[iconColorField].Value(),
	}
}

// saveTask saves a task to the local configuration file.
func (t TaskModel) saveTask(task Task) error {
	if err := os.MkdirAll(path.Dir(TasksSaveFile), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(TasksSaveFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonContent, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var content TaskSaveFileContent
	if len(jsonContent) > 0 {
		if err = json.Unmarshal(jsonContent, &content); err != nil {
			return err
		}
	}

	content.Tasks = append(content.Tasks, task)

	newJsonContent, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(TasksSaveFile, newJsonContent, 0666)
}

func (t *TaskModel) isValidTask(task Task) bool {

	if strings.TrimSpace(task.Name) == "" {
		t.messages.AddError("Name is required")
	}

	p := strings.TrimSpace(task.Path)
	if strings.HasSuffix(p, ".") {
		relativePath := path.Join(os.Getenv("PWD"), p)
		if _, err := os.Stat(relativePath); os.IsNotExist(err) {
			t.messages.AddError("Path does not exist")
		}
	} else {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.messages.AddError("Path does not exist")
		}
	}

	if len(task.Cmds) == 0 || (len(task.Cmds) == 1 && strings.TrimSpace(task.Cmds[0]) == "") {
		t.messages.AddError("At least one command is required")
	}

	_, taskIconExists := lo.Find(vscode.Icons, func(i vscode.Icon) bool {
		return i.Name == task.Icon
	})
	if task.Icon == "" || !taskIconExists {
		t.messages.AddError("Invalid Icon")
	}

	_, taskColorExists := lo.Find(vscode.ANSIColors, func(c vscode.ANSIColor) bool {
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

	if err := os.MkdirAll(path.Dir(TasksSaveFile), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(TasksSaveFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonContent, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var content TaskSaveFileContent
	if len(jsonContent) > 0 {
		if err = json.Unmarshal(jsonContent, &content); err != nil {
			return err
		}
	}

	content.Tasks = lo.Filter(content.Tasks, func(task Task, index int) bool {
		return task.Name != name
	})

	encoded, err := json.Marshal(content)
	if err != nil {
		return err
	}

	file.Truncate(0)
	file.Seek(0, 0)
	_, err = file.Write(encoded)
	return err
}
