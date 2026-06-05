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
}

// ensureTasksSaveFile creates the directory and the tasks file if they do not
// exist yet. It is called lazily by every public function so that importing
// this package no longer touches the filesystem at init time.
func ensureTasksSaveFile() {
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
	ensureTasksSaveFile()
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
	ensureTasksSaveFile()
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

	content.Tasks = appendTask(content.Tasks, task)

	newJsonContent, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(TasksSaveFile, newJsonContent, 0666)
}

// appendTask returns a new slice with task appended. Pure: no I/O.
func appendTask(tasks []models.Task, task models.Task) []models.Task {
	return append(tasks, task)
}

// SaveFromFile saves tasks from a given JSON file specified by a flag
func SaveFromFile(path string) error {
	ensureTasksSaveFile()
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
		return errors.New("failed to open existing tasks file: " + err.Error())
	}

	jsonBytes, err := io.ReadAll(saveFile)
	saveFile.Close()
	if err != nil {
		return errors.New("failed to read existing tasks file: " + err.Error())
	}

	// An empty destination means zero existing tasks, not an error: a fresh
	// install always starts with an empty tasks.json, and the batch must still
	// import. Only decode when there is content to decode.
	var content TaskSaveFileContent
	if len(jsonBytes) > 0 {
		if err = json.Unmarshal(jsonBytes, &content); err != nil {
			return errors.New("failed to parse existing tasks file: " + err.Error())
		}
	}

	content.Tasks = appendTaskBatch(content.Tasks, newTasks)
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

// appendTaskBatch returns a new slice with all items from batch appended. Pure: no I/O.
func appendTaskBatch(tasks []models.Task, batch []models.Task) []models.Task {
	return append(tasks, batch...)
}

// UpdateTask modifies an existing task in the local configuration file.
func UpdateTask(originalName string, updatedTask models.Task) error {
	ensureTasksSaveFile()
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

	updated, err := replaceTaskByName(content.Tasks, originalName, updatedTask)
	if err != nil {
		return err
	}
	content.Tasks = updated

	newJsonContent, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(TasksSaveFile, newJsonContent, 0666)
}

// replaceTaskByName returns a copy of tasks with the entry matching originalName
// replaced by updatedTask. Returns an error when the name is not found. Pure: no I/O.
func replaceTaskByName(tasks []models.Task, originalName string, updatedTask models.Task) ([]models.Task, error) {
	taskIndex := -1
	for i, task := range tasks {
		if task.Name == originalName {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return nil, fmt.Errorf("task '%s' not found", originalName)
	}

	result := make([]models.Task, len(tasks))
	copy(result, tasks)
	result[taskIndex] = updatedTask
	return result, nil
}

// DeleteTask removes a task from the local configuration file by name.
func DeleteTask(name string) error {
	ensureTasksSaveFile()
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

	content.Tasks = filterOutTaskByName(content.Tasks, name)

	encoded, err := json.Marshal(content)
	if err != nil {
		return err
	}

	file.Truncate(0)
	file.Seek(0, 0)
	_, err = file.Write(encoded)
	return err
}

// filterOutTaskByName returns a copy of tasks without the entry whose Name equals name.
// Pure: no I/O.
func filterOutTaskByName(tasks []models.Task, name string) []models.Task {
	return lo.Filter(tasks, func(task models.Task, _ int) bool {
		return task.Name != name
	})
}

// GetAllTasks retrieves all saved tasks.
func GetAllTasks() ([]models.Task, error) {
	return ReadTasks()
}
