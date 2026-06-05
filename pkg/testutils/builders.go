// pkg/testutils/builders.go
package testutils

import "github.com/DieGopherLT/vscode-terminal-runner/internal/models"

// TaskBuilder constructs models.Task values with sensible defaults and per-test overrides.
// Use NewTask() to start, chain With* methods, and call Build() to produce the value.
//
// Example:
//
//	task := testutils.NewTask().WithName("build").WithCmds("make", "go build").Build()
type TaskBuilder struct {
	task models.Task
}

// NewTask returns a TaskBuilder pre-populated with stable defaults so tests
// only need to declare the field that is relevant to the scenario under test.
func NewTask() *TaskBuilder {
	return &TaskBuilder{
		task: models.Task{
			Name:      "default-task",
			Path:      "/default/path",
			Cmds:      []string{"echo hello"},
			Icon:      "",
			IconColor: "",
		},
	}
}

// WithName overrides the task name.
func (b *TaskBuilder) WithName(name string) *TaskBuilder {
	b.task.Name = name
	return b
}

// WithPath overrides the task path.
func (b *TaskBuilder) WithPath(path string) *TaskBuilder {
	b.task.Path = path
	return b
}

// WithCmds overrides the task commands.
func (b *TaskBuilder) WithCmds(cmds ...string) *TaskBuilder {
	b.task.Cmds = cmds
	return b
}

// WithIcon overrides the task icon.
func (b *TaskBuilder) WithIcon(icon string) *TaskBuilder {
	b.task.Icon = icon
	return b
}

// WithIconColor overrides the task icon color.
func (b *TaskBuilder) WithIconColor(color string) *TaskBuilder {
	b.task.IconColor = color
	return b
}

// Build returns the constructed models.Task.
func (b *TaskBuilder) Build() models.Task {
	return b.task
}

// WorkspaceBuilder constructs models.Workspace values with sensible defaults and per-test overrides.
// Use NewWorkspace() to start, chain With* methods, and call Build() to produce the value.
//
// Example:
//
//	ws := testutils.NewWorkspace().WithName("my-project").WithTasks(task1, task2).Build()
type WorkspaceBuilder struct {
	workspace models.Workspace
}

// NewWorkspace returns a WorkspaceBuilder pre-populated with stable defaults so tests
// only need to declare the field that is relevant to the scenario under test.
func NewWorkspace() *WorkspaceBuilder {
	return &WorkspaceBuilder{
		workspace: models.Workspace{
			Name:  "default-workspace",
			Tasks: []models.Task{},
		},
	}
}

// WithName overrides the workspace name.
func (b *WorkspaceBuilder) WithName(name string) *WorkspaceBuilder {
	b.workspace.Name = name
	return b
}

// WithTasks overrides the workspace task list.
func (b *WorkspaceBuilder) WithTasks(tasks ...models.Task) *WorkspaceBuilder {
	b.workspace.Tasks = tasks
	return b
}

// Build returns the constructed models.Workspace.
func (b *WorkspaceBuilder) Build() models.Workspace {
	return b.workspace
}
