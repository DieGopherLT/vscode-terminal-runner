package workspace

import "github.com/DieGopherLT/vscode-terminal-runner/internal/task"

type Workspace struct {
	Name  string      `json:"name"`
	Tasks []task.Task `json:"tasks"`
}
