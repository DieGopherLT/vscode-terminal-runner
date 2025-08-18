package task

import (
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui/suggestions"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	nameField      = 0 // Name field index
	pathField      = 1 // Path field index
	cmdsField      = 2 // Commands field index
	iconField      = 3 // Icon field index
	iconColorField = 4 // Icon color field index
)

// Init initializes the TUI model (cursor blinking).
func (t *TaskModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages received by the TUI model and updates the form state.
func (t *TaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return t, tea.Quit

		case string(tui.KeyUp), string(tui.KeyDown), string(tui.KeyTab), string(tui.KeyShiftTab):
			key := msg.String()
			
			// If there are suggestions and Tab pressed, apply suggestion
			if key == string(tui.KeyTab) {
				if t.nav.FocusIndex == pathField && t.pathSuggestions.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
					t.pathSuggestions.ApplySelected(&t.inputs[t.nav.FocusIndex])
					return t, nil
				} else if manager := t.getCurrentSuggestionManager(); manager != nil && manager.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
					manager.ApplySelected(&t.inputs[t.nav.FocusIndex])
					return t, nil
				}
			}
			
			t.nav.HandleNavigation(key)
			// Reset suggestion managers when navigating between fields
			t.iconSuggestions.Reset()
			t.colorSuggestions.Reset()
			t.pathSuggestions.Reset()
			return t.HandleFocus()

		case "ctrl+n":
			// Navigate down through suggestions (circular)
			if t.nav.FocusIndex == pathField {
				t.pathSuggestions.Next()
			} else if manager := t.getCurrentSuggestionManager(); manager != nil {
				manager.Next()
			}
			return t, nil

		case "ctrl+b":
			// Navigate up through suggestions (circular)
			if t.nav.FocusIndex == pathField {
				t.pathSuggestions.Previous()
			} else if manager := t.getCurrentSuggestionManager(); manager != nil {
				manager.Previous()
			}
			return t, nil

		case "enter":
			// If there are suggestions, apply the selected one
			if t.nav.FocusIndex == pathField && t.pathSuggestions.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
				t.pathSuggestions.ApplySelected(&t.inputs[t.nav.FocusIndex])
				return t, nil
			} else if manager := t.getCurrentSuggestionManager(); manager != nil && manager.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
				manager.ApplySelected(&t.inputs[t.nav.FocusIndex])
				return t, nil
			}
			
			if t.nav.FocusIndex != len(t.inputs) {
				return t, nil
			}
			task := t.handleTaskCreation()

			if !t.isValidTask(task) {
				return t, nil
			}

			if err := t.saveTask(task); err != nil {
				return t, tea.Quit
			}
			
			successMessage := "Task created successfully!"
			if t.isEditMode {
				successMessage = "Task updated successfully!"
			}
			t.messages.AddSuccess(successMessage)
			return t, tea.Quit
		}
	}

	// Only process input if it's not suggestion navigation
	cmd := t.HandleInput(msg)

	return t, cmd
}

// HandleFocus updates the visual focus and style of the form fields.
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

// HandleInput processes text input and updates the suggestion managers.
func (t *TaskModel) HandleInput(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(t.inputs))

	// Clear messages when user starts typing
	if t.messages.HasMessages() {
		t.messages.Clear()
	}

	for i := range t.inputs {
		t.inputs[i], cmds[i] = t.inputs[i].Update(msg)

		// Update suggestion managers based on input changes
		if i == pathField && i == t.nav.FocusIndex {
			// For path suggestions, we use PathManager which provides filesystem-based autocomplete
			t.pathSuggestions.UpdateFilter(t.inputs[i].Value())
		}

		if i == iconField && i == t.nav.FocusIndex {
			t.iconSuggestions.UpdateFilter(t.inputs[i].Value())
		}

		if i == iconColorField && i == t.nav.FocusIndex {
			t.colorSuggestions.UpdateFilter(t.inputs[i].Value())
		}
	}

	return tea.Batch(cmds...)
}

// View renders the TUI form view for creating/editing tasks.
func (t *TaskModel) View() string {
	var sections []string
	
	title := "CREATE TASK"
	if t.isEditMode {
		title = "EDIT TASK"
	}
	sections = append(sections, styles.RenderTitle(title))
	
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
		
		// Show suggestions for the current focused field
		if t.nav.FocusIndex == i {
			var suggestionBox string
			if i == pathField && t.pathSuggestions.ShouldShow(t.inputs[i].Value()) {
				suggestionBox = t.pathSuggestions.Render()
			} else if manager := t.getCurrentSuggestionManager(); manager != nil && manager.ShouldShow(t.inputs[i].Value()) {
				suggestionBox = manager.Render()
			}
			
			if suggestionBox != "" {
				fieldContent = lipgloss.JoinVertical(
					lipgloss.Left,
					fieldContent,
					suggestionBox,
				)
			}
		}
		
		sections = append(sections, styles.FieldContainerStyle.Render(fieldContent))
	}
	
	// Render messages if any exist
	if t.messages.HasMessages() {
		sections = append(sections, t.messages.Render())
	}
	
	button := styles.RenderBlurredButton("Submit")
	if t.nav.FocusIndex == len(t.inputs) {
		button = styles.RenderFocusedButton("Submit")
	}
	
	sections = append(sections, button)
	
	helpText := styles.HelpTextStyle.Render("↑/↓ navigate • ctrl+b/n suggestions • tab/enter apply • esc quit")
	sections = append(sections, helpText)
	
	return styles.FormContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
}

// getCurrentSuggestionManager returns the suggestion manager for the current field.
func (t *TaskModel) getCurrentSuggestionManager() *suggestions.Manager {
	switch t.nav.FocusIndex {
	case pathField:
		return t.pathSuggestions.Manager
	case iconField:
		return t.iconSuggestions
	case iconColorField:
		return t.colorSuggestions
	default:
		return nil
	}
}

