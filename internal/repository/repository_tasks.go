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

// NewTaskRepository builds the task repository: tasks are keyed under "tasks"
// and rejected on a case-sensitive name collision, since the task name is the
// handle every read path (find, run, completion, workspace selector) keys on.
func NewTaskRepository(getSaveFile func() string) *JSONRepository[models.Task] {
	return &JSONRepository[models.Task]{
		getSaveFile: getSaveFile,
		jsonKey:     "tasks",
		entityLabel: "task",
		onAppend: func(existing []models.Task, task models.Task) ([]models.Task, error) {
			_, found := lo.Find(existing, func(candidate models.Task) bool {
				return candidate.Name == task.Name
			})
			if found {
				return nil, fmt.Errorf("task '%s' already exists", task.Name)
			}
			return append(existing, task), nil
		},
	}
}

// ReadTasks loads all tasks from the persistence file.
func ReadTasks() ([]models.Task, error) {
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
