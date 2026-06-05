// pkg/testutils/helpers.go
package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
)

// ContainsString checks if a string contains a substring
func ContainsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

// CreateTestJSONFile creates a temporary JSON file with the given data and permissions
func CreateTestJSONFile(data interface{}, permissions os.FileMode) (string, error) {
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "test-file.json")

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(tempFile, jsonData, permissions); err != nil {
		return "", err
	}

	return tempFile, nil
}

// CreateTempFileWithPermissions creates a temporary file with specific permissions
func CreateTempFileWithPermissions(permissions os.FileMode) (string, error) {
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "test-file")

	if err := os.WriteFile(tempFile, []byte("test"), permissions); err != nil {
		return "", err
	}

	return tempFile, nil
}

// CreateTempDirWithPermissions creates a temporary directory with specific permissions
func CreateTempDirWithPermissions(permissions os.FileMode) (string, error) {
	tempDir := filepath.Join(os.TempDir(), "test-dir")

	if err := os.MkdirAll(tempDir, permissions); err != nil {
		return "", err
	}

	return tempDir, nil
}

// TaskBuilder builds models.Task values with sensible defaults and per-field overrides.
// Call newTask() to start, chain With* methods, then Build().
type TaskBuilder struct {
	task models.Task
}

// NewTask returns a TaskBuilder whose defaults produce a valid, runnable Task.
func NewTask() *TaskBuilder {
	return &TaskBuilder{
		task: models.Task{
			Name:      "default-task",
			Path:      "/workspace/project",
			Cmds:      []string{"echo hello"},
			Icon:      "terminal",
			IconColor: "terminal.ansiGreen",
		},
	}
}

// WithName overrides the task name.
func (b *TaskBuilder) WithName(name string) *TaskBuilder {
	b.task.Name = name
	return b
}

// WithPath overrides the working directory path.
func (b *TaskBuilder) WithPath(path string) *TaskBuilder {
	b.task.Path = path
	return b
}

// WithCmds overrides the command list.
func (b *TaskBuilder) WithCmds(cmds ...string) *TaskBuilder {
	b.task.Cmds = cmds
	return b
}

// WithIcon overrides the terminal icon.
func (b *TaskBuilder) WithIcon(icon string) *TaskBuilder {
	b.task.Icon = icon
	return b
}

// WithIconColor overrides the icon color.
func (b *TaskBuilder) WithIconColor(color string) *TaskBuilder {
	b.task.IconColor = color
	return b
}

// Build returns the fully constructed Task.
func (b *TaskBuilder) Build() models.Task {
	return b.task
}

// WorkspaceBuilder builds models.Workspace values with sensible defaults and per-field overrides.
// Call NewWorkspace() to start, chain With* methods, then Build().
type WorkspaceBuilder struct {
	workspace models.Workspace
}

// NewWorkspace returns a WorkspaceBuilder whose defaults produce a valid Workspace
// containing one default Task.
func NewWorkspace() *WorkspaceBuilder {
	return &WorkspaceBuilder{
		workspace: models.Workspace{
			Name:  "default-workspace",
			Tasks: []models.Task{NewTask().Build()},
		},
	}
}

// WithName overrides the workspace name.
func (b *WorkspaceBuilder) WithName(name string) *WorkspaceBuilder {
	b.workspace.Name = name
	return b
}

// WithTasks replaces the task list entirely.
func (b *WorkspaceBuilder) WithTasks(tasks ...models.Task) *WorkspaceBuilder {
	b.workspace.Tasks = tasks
	return b
}

// WithNoTasks sets an empty task list (useful for testing the empty-workspace error path).
func (b *WorkspaceBuilder) WithNoTasks() *WorkspaceBuilder {
	b.workspace.Tasks = nil
	return b
}

// Build returns the fully constructed Workspace.
func (b *WorkspaceBuilder) Build() models.Workspace {
	return b.workspace
}
