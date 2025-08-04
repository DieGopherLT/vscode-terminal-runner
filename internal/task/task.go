package task

import (
	"fmt"

	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type Task struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Cmds      []string `json:"cmds"`
	Icon      string   `json:"icon"`
	IconColor string   `json:"iconColor"`
}

type TaskModel struct {
	nav        *tui.FormNavigator
	inputs     []textinput.Model
}

func NewModel() tea.Model {
	numberOfFields := 5

	model := &TaskModel{
		inputs: make([]textinput.Model, numberOfFields),
		nav:    tui.NewNavigator(numberOfFields),
	}

	for i := range model.inputs {
		t := textinput.New()
		t.Cursor.Style = cursorStyle
		t.Width = 50  // Agregar width fijo

		switch i {
		case nameField:
			t.Placeholder = "Task name"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case pathField:
			t.Placeholder = "Path"
		case cmdsField:
			t.Placeholder = "Commands"
		case iconField:
			t.Placeholder = "Icon"
		case iconColorField:
			t.Placeholder = "Icon color"
		}
		model.inputs[i] = t
	}

	return model
}