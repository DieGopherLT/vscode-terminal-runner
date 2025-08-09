package models

// Task represents an individual task that can be executed in a VSCode terminal.
type Task struct {
	Name      string   `json:"name"`      // Task name
	Path      string   `json:"path"`      // Associated project path
	Cmds      []string `json:"cmds"`      // Commands to execute
	Icon      string   `json:"icon"`      // VSCode terminal icon
	IconColor string   `json:"iconColor"` // Icon color in the terminal
}

// Workspace represents a workspace containing multiple tasks.
type Workspace struct {
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}
