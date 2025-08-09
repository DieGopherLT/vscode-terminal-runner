package cfg

import (
	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
)

type Config struct {
	Mode       string             `json:"mode"`
	Tasks      []models.Task      `json:"tasks"`
	Workspaces []models.Workspace `json:"workspaces"`
}
