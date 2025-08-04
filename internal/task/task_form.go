package task

import (
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	nameField      = 0
	pathField      = 1
	cmdsField      = 2
	iconField      = 3
	iconColorField = 4
)

func (t *TaskModel) Init() tea.Cmd {
	return textinput.Blink
}

func (t *TaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return t, tea.Quit

		case "up", "down", "tab", "shift+tab":
			key := msg.String()
			t.nav.HandleNavigation(key)
			return t.HandleFocus()

		case "enter":
			if t.nav.FocusIndex != len(t.inputs) {
				return t, nil
			}
			task := t.HandleTaskCreation()
			if err := SaveTask(task); err != nil {
				return t, tea.Quit
			}
			return t, tea.Quit
		}
	}

	cmd := t.HandleInput(msg)

	return t, cmd
}

func (t *TaskModel) HandleFocus() (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(t.inputs))

	for i := 0; i < len(t.inputs); i++ {
		if i == t.nav.FocusIndex {
			cmds[i] = t.inputs[i].Focus()
			t.inputs[i].PromptStyle = styles.FocusedInputStyle
			t.inputs[i].TextStyle = styles.FocusedInputStyle
			continue
		}
		t.inputs[i].Blur()
		t.inputs[i].PromptStyle = noStyle
		t.inputs[i].TextStyle = noStyle
	}

	return t, tea.Batch(cmds...)
}

func (t *TaskModel) HandleInput(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(t.inputs))

	for i := range t.inputs {
		t.inputs[i], cmds[i] = t.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (t *TaskModel) View() string {
	var sections []string
	
	sections = append(sections, styles.RenderTitle("CREATE TASK"))
	
	labels := []string{
		"Task Name:",
		"Project Path:",
		"Commands:",
		"Icon:",
		"Icon Color:",
	}
	
	for i := range t.inputs {
		fieldContent := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.FieldLabelStyle.Render(labels[i]),
			t.inputs[i].View(),
		)
		sections = append(sections, styles.FieldContainerStyle.Render(fieldContent))
	}
	
	button := styles.RenderBlurredButton("Submit")
	if t.nav.FocusIndex == len(t.inputs) {
		button = styles.RenderFocusedButton("Submit")
	}
	
	sections = append(sections, button)
	
	helpText := styles.HelpTextStyle.Render("↑/↓ navigate • enter submit • esc quit")
	sections = append(sections, helpText)
	
	return styles.FormContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
}
