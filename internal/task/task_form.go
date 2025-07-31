package task

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
			t.HandleNavigate(key)
			return t.HandleFocus()

		case "enter":
			if t.focusIndex != len(t.inputs) {
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

func (t *TaskModel) HandleNavigate(key string) {
	switch key {
	case "up", "sift+tab":
		t.focusIndex--
	case "down", "tab":
		t.focusIndex++
	}

	if t.focusIndex > len(t.inputs) {
		t.focusIndex = 0
	} else if t.focusIndex < 0 {
		t.focusIndex = len(t.inputs)
	}

}

func (t *TaskModel) HandleFocus() (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(t.inputs))

	for i := 0; i < len(t.inputs); i++ {
		if i == t.focusIndex {
			cmds[i] = t.inputs[i].Focus()
			t.inputs[i].PromptStyle = focusedStyle
			t.inputs[i].TextStyle = focusedStyle
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
	b := strings.Builder{}

	for i := range t.inputs {
		b.WriteString(t.inputs[i].View())
		if i < len(t.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if t.focusIndex == len(t.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}
