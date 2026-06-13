package models

// Task represents an individual task that can be executed in a VSCode terminal.
type Task struct {
	Name      string   `json:"name"`      // Task name
	Path      string   `json:"path"`      // Associated project path
	Cmds      []string `json:"cmds"`      // Commands to execute
	Icon      string   `json:"icon"`      // VSCode terminal icon
	IconColor string   `json:"iconColor"` // Icon color in the terminal
}

// GetName returns the name that uniquely identifies this task.
func (t Task) GetName() string {
	return t.Name
}

// Workspace represents a workspace containing multiple tasks.
type Workspace struct {
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}

// GetName returns the name that uniquely identifies this workspace.
func (w Workspace) GetName() string {
	return w.Name
}

// Config represents the configuration for the terminal runner.
type Config struct {
	IsSetupComplete bool `json:"is_setup_complete"`
}
