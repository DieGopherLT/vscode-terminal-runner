package task

import (
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	noStyle = lipgloss.NewStyle()
)

type Task struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Cmds      []string `json:"cmds"`
	Icon      string   `json:"icon"`
	IconColor string   `json:"iconColor"`
}

type TaskModel struct {
	nav              *tui.FormNavigator
	inputs           []textinput.Model
	suggestionIndex  int  // Índice de la sugerencia seleccionada
}

func NewModel() tea.Model {
	numberOfFields := 5

	model := &TaskModel{
		inputs: make([]textinput.Model, numberOfFields),
		nav:    tui.NewNavigator(numberOfFields),
	}

	for i := range model.inputs {
		t := textinput.New()
		t.Cursor.Style = styles.FocusedInputStyle
		t.Width = 60
		t.Prompt = ""  // Sin prompt, usaremos labels externos
		t.PlaceholderStyle = styles.PlaceholderStyle

		switch i {
		case nameField:
			t.Placeholder = "Enter task name..."
			t.Focus()
			t.PromptStyle = styles.FocusedInputStyle
			t.TextStyle = styles.FocusedInputStyle
		case pathField:
			t.Placeholder = "e.g., /home/user/project"
		case cmdsField:
			t.Placeholder = "npm start"
		case iconField:
			t.Placeholder = "e.g., terminal-bash"
			t.ShowSuggestions = true
		case iconColorField:
			t.Placeholder = "terminal.<color> (e.g., terminal.ansiGreen)"
			t.ShowSuggestions = true
		}
		model.inputs[i] = t
	}

	return model
}