package task

import (
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/messages"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui/suggestions"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

var (
	noStyle = lipgloss.NewStyle()
)


// TaskModel manages the state and logic of the TUI form for creating/editing tasks.
type TaskModel struct {
	nav                *tui.FormNavigator
	inputs             []textinput.Model
	iconSuggestions    *suggestions.Manager
	colorSuggestions   *suggestions.Manager
	pathSuggestions    *suggestions.Manager
	messages           *messages.MessageManager
	lastPathDirectory  string
}

// NewModel initializes and returns the TUI model for the task creation form.
func NewModel() tea.Model {
	numberOfFields := 5

	// Create suggestion managers
	iconNames := lo.Map(styles.VSCodeIcons, func(i styles.VSCodeIcon, _ int) string { return i.Name })
	colorNames := lo.Map(styles.VSCodeANSIColors, func(c styles.VSCodeANSIColor, _ int) string { return c.Name })

	model := &TaskModel{
		inputs:           make([]textinput.Model, numberOfFields),
		nav:              tui.NewNavigator(numberOfFields),
		iconSuggestions:  suggestions.NewManager(iconNames, 3, suggestions.ContainsFilter),
		colorSuggestions: suggestions.NewManager(colorNames, 3, suggestions.ContainsFilter),
		pathSuggestions:  suggestions.NewManagerWithOptions([]string{}, 5, suggestions.StartsWithFilter, false),
		messages:         messages.NewManager(),
	}

	for i := range model.inputs {
		t := textinput.New()
		t.Cursor.Style = styles.FocusedInputStyle
		t.Width = 70
		t.Prompt = ""  // No prompt, we'll use external labels
		t.PlaceholderStyle = styles.PlaceholderStyle

		switch i {
		case nameField:
			t.Placeholder = "Enter task name..."
			t.Focus()
			t.PromptStyle = styles.FocusedInputStyle
			t.TextStyle = styles.FocusedInputStyle
		case pathField:
			t.Placeholder = "e.g., /home/user/project, absolute path or relative cwd"
		case cmdsField:
			t.Placeholder = "cmd1, cmd2... (e.g., yarn install, yarn dev)"
		case iconField:
			t.Placeholder = "e.g., terminal-bash"
		case iconColorField:
			t.Placeholder = "terminal.<color> (e.g., terminal.ansiGreen)"
		}
		model.inputs[i] = t
	}

	return model
}