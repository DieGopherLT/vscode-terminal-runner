package cfg

import (
	"github.com/DieGopherLT/vscode-terminal-runner/internal/task"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/workspace"
)

type Config struct {
	Mode       string                `json:"mode"`
	Tasks      []task.Task           `json:"tasks"`
	Workspaces []workspace.Workspace `json:"workspaces"`
}
