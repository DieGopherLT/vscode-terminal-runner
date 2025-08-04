package task

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"

	"github.com/samber/lo"
)

var TasksSaveFile = path.Join(os.Getenv("HOME"), ".config/vsct-runner/tasks.json")

type TaskSaveFileContent struct {
	Tasks []Task `json:"tasks"`
}

func (t TaskModel) HandleTaskCreation() Task {
	return Task{
		Name:      t.inputs[nameField].Value(),
		Path:      t.inputs[pathField].Value(),
		Cmds:      strings.Split(t.inputs[cmdsField].Value(), "\n"),
		Icon:      t.inputs[iconField].Value(),
		IconColor: t.inputs[iconColorField].Value(),
	}
}

func SaveTask(task Task) error {
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
