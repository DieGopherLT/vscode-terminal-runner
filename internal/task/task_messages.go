package task

import "github.com/DieGopherLT/vscode-terminal-runner/internal/models"

// taskSubmitMsg is sent when the user submits the task form.
type taskSubmitMsg struct {
	task models.Task
}

// taskSaveResultMsg is sent after async path validation and save operations.
type taskSaveResultMsg struct {
	err error
}
