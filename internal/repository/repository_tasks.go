package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/samber/lo"
)

var (
	// TasksSaveFile holds the absolute path to the tasks.json file in the user's config directory.
	TasksSaveFile string

	// taskRepo is the singleton generic repository backing every task adapter below.
	taskRepo *JSONRepository[models.Task]
)

func init() {
	cfgFolder, err := os.UserConfigDir()
	if err != nil {
		panic("could not determine user config directory: " + err.Error())
	}
	TasksSaveFile = filepath.Join(cfgFolder, "vscode-terminal-runner", "tasks.json")
	taskRepo = NewTaskRepository(func() string { return TasksSaveFile })
}

// ReadTasks loads all tasks from the persistence file.
func ReadTasks() ([]models.Task, error) {
	return taskRepo.ReadAll()
}

// GetAllTasks retrieves all saved tasks.
func GetAllTasks() ([]models.Task, error) {
	return taskRepo.ReadAll()
}

// FindTaskByName retrieves a task by its name from the saved tasks.
func FindTaskByName(name string) (*models.Task, error) {
	return taskRepo.FindByName(name)
}

// SaveTask saves a task to the local configuration file.
func SaveTask(task models.Task) error {
	return taskRepo.Save(task)
}

// UpdateTask modifies an existing task in the local configuration file.
func UpdateTask(originalName string, updatedTask models.Task) error {
	return taskRepo.Update(originalName, updatedTask)
}

// DeleteTask removes a task from the local configuration file by name.
func DeleteTask(name string) error {
	return taskRepo.Delete(name)
}

// appendTask returns a new slice with task appended. Pure: no I/O.
func appendTask(tasks []models.Task, task models.Task) []models.Task {
	return append(tasks, task)
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

// filterOutTaskByName returns a copy of tasks without the entry whose Name equals name.
// Pure: no I/O.
func filterOutTaskByName(tasks []models.Task, name string) []models.Task {
	return lo.Filter(tasks, func(task models.Task, _ int) bool {
		return task.Name != name
	})
}
