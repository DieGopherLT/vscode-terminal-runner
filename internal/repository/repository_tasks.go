package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/samber/lo"
)

var (
	// tasksSaveFile holds the absolute path to the tasks.json file in the user's config directory.
	TasksSaveFile string
)

type TasksBatchModel []models.Task

func init() {
	cfgFolder, err := os.UserConfigDir()
	if err != nil {
		panic("could not determine user config directory: " + err.Error())
	}
	TasksSaveFile = filepath.Join(cfgFolder, "vscode-terminal-runner", "tasks.json")

	if _, err := os.Stat(TasksSaveFile); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(TasksSaveFile), 0755); err != nil {
			return
		}
		if _, err := os.Create(TasksSaveFile); err != nil {
			return
		}
	}
}

// TaskSaveFileContent represents the structure of the task persistence file.
type TaskSaveFileContent struct {
	Tasks []models.Task `json:"tasks"`
}

// ReadTasks loads all tasks from the persistence file.
func ReadTasks() ([]models.Task, error) {
	file, err := os.OpenFile(TasksSaveFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	jsonContent, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var content TaskSaveFileContent
	if len(jsonContent) > 0 {
		if err = json.Unmarshal(jsonContent, &content); err != nil {
			return nil, err
		}
	}

	return content.Tasks, nil
}

// FindTaskByName retrieves a task by its name from the saved tasks.
func FindTaskByName(name string) (*models.Task, error) {
	tasks, err := ReadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	task, found := lo.Find(tasks, func(task models.Task) bool {
		return strings.EqualFold(task.Name, name)
	})

	if !found {
		return nil, fmt.Errorf("task '%s' not found", name)
	}

	return &task, nil
}

// SaveTask saves a task to the local configuration file.
func SaveTask(task models.Task) error {
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


// SaveFromFile saves tasks from a given JSON file specified by a flag
func SaveFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.New("failed to open file: " + err.Error())
	}
	defer file.Close()

	var newTasks TasksBatchModel
	err = json.NewDecoder(file).Decode(&newTasks)
	if err != nil {
		return errors.New("Incorrect file format: " + err.Error())
	}

	saveFile, err := os.Open(TasksSaveFile)
	if err != nil {
		return errors.New("Error when creating tasks:" + err.Error())
	}

	jsonBytes, err := io.ReadAll(saveFile)
	if err != nil {
		return errors.New("Error when creating tasks" + err.Error())
	}

	if len(jsonBytes) == 0 {
		return errors.New("Provided file is empty")
	}

	var content TaskSaveFileContent
	if err = json.Unmarshal(jsonBytes, &content); err != nil {
		return errors.New("Error when creating tasks:" + err.Error())
	}

	content.Tasks = append(content.Tasks, newTasks...)
	newJsonContent, err := json.Marshal(content)
	if err != nil {
		return errors.New("Error when saving tasks:" + err.Error())
	}
	
	err = os.WriteFile(TasksSaveFile, newJsonContent, 0666)
	if err != nil {
		return errors.New("Error when saving tasks:" + err.Error())
	}

	return nil
}

// UpdateTask modifies an existing task in the local configuration file.
func UpdateTask(originalName string, updatedTask models.Task) error {
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

	// Find and replace the task
	taskIndex := -1
	for i, task := range content.Tasks {
		if task.Name == originalName {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("task '%s' not found", originalName)
	}

	content.Tasks[taskIndex] = updatedTask

	newJsonContent, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(TasksSaveFile, newJsonContent, 0666)
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

	content.Tasks = lo.Filter(content.Tasks, func(task models.Task, index int) bool {
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

// GetAllTasks retrieves all saved tasks.
func GetAllTasks() ([]models.Task, error) {
	return ReadTasks()
}
