package task

import (
	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui/suggestions"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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

// Init initializes the TUI model (cursor blinking).
func (t *TaskModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages received by the TUI model and updates the form state.
func (t *TaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for i := range t.inputs {
			t.inputs[i].Width = msg.Width - 10
		}
		suggestionWidth := styles.ResponsiveContainerWidth(msg.Width, 58, 6)
		t.pathSuggestions.SetMaxWidth(suggestionWidth)
		t.iconSuggestions.SetMaxWidth(suggestionWidth)
		t.colorSuggestions.SetMaxWidth(suggestionWidth)
		t.messages.SetMaxWidth(styles.ResponsiveContainerWidth(msg.Width, 70, 6))
		return t, nil

	case suggestions.PathSuggestionsLoadedMsg:
		t.pathSuggestions.ApplyScannedSuggestions(msg)
		return t, nil

	case spinner.TickMsg:
		if !t.isSubmitting {
			return t, nil
		}
		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd

	case taskSaveResultMsg:
		t.isSubmitting = false
		if msg.err != nil {
			t.messages.AddError(msg.err.Error())
			return t, nil
		}
		successMessage := "Task created successfully!"
		if t.isEditMode {
			successMessage = "Task updated successfully!"
		}
		t.messages.AddSuccess(successMessage)
		return t, tea.Quit

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.DefaultKeys.Quit):
			return t, tea.Quit

		case key.Matches(msg, tui.DefaultKeys.Up),
			key.Matches(msg, tui.DefaultKeys.Down),
			key.Matches(msg, tui.DefaultKeys.Tab),
			key.Matches(msg, tui.DefaultKeys.ShiftTab):

			// If Tab pressed and suggestions available, apply suggestion
			if key.Matches(msg, tui.DefaultKeys.Tab) {
				if t.nav.FocusIndex == pathField && t.pathSuggestions.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
					return t, t.pathSuggestions.ApplySelected(&t.inputs[t.nav.FocusIndex])
				}
				if manager := t.getCurrentSuggestionManager(); manager != nil && manager.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
					manager.ApplySelected(&t.inputs[t.nav.FocusIndex])
					return t, nil
				}
			}

			t.nav.HandleNavigation(msg.String())
			// Reset suggestion managers when navigating between fields
			t.iconSuggestions.Reset()
			t.colorSuggestions.Reset()
			t.pathSuggestions.Reset()
			return t.HandleFocus()

		case key.Matches(msg, tui.DefaultKeys.NextSuggestion):
			if t.nav.FocusIndex == pathField {
				t.pathSuggestions.Next()
			} else if manager := t.getCurrentSuggestionManager(); manager != nil {
				manager.Next()
			}
			return t, nil

		case key.Matches(msg, tui.DefaultKeys.PrevSuggestion):
			if t.nav.FocusIndex == pathField {
				t.pathSuggestions.Previous()
			} else if manager := t.getCurrentSuggestionManager(); manager != nil {
				manager.Previous()
			}
			return t, nil

		case key.Matches(msg, tui.DefaultKeys.Enter):
			// If there are suggestions, apply the selected one
			if t.nav.FocusIndex == pathField && t.pathSuggestions.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
				return t, t.pathSuggestions.ApplySelected(&t.inputs[t.nav.FocusIndex])
			}
			if manager := t.getCurrentSuggestionManager(); manager != nil && manager.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
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

			t.isSubmitting = true
			return t, tea.Batch(submitTaskCmd(t, task), t.spinner.Tick)
		}
	}

	// Only process input if it's not suggestion navigation
	cmd := t.HandleInput(msg)
	return t, cmd
}

// submitTaskCmd performs path validation and persistence off the UI goroutine.
func submitTaskCmd(t *TaskModel, task models.Task) tea.Cmd {
	return func() tea.Msg {
		if err := validateTaskPath(task.Path, t.expandPathForValidation); err != nil {
			return taskSaveResultMsg{err: err}
		}
		if err := t.saveTask(task); err != nil {
			return taskSaveResultMsg{err: err}
		}
		return taskSaveResultMsg{}
	}
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
		t.inputs[i].PromptStyle = styles.UnfocusedInputStyle
		t.inputs[i].TextStyle = styles.UnfocusedInputStyle
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
			cmds = append(cmds, t.pathSuggestions.UpdateFilter(t.inputs[i].Value()))
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
	switch {
	case t.isSubmitting:
		button = styles.RenderBlurredButton(t.spinner.View() + "Saving...")
	case t.nav.FocusIndex == len(t.inputs):
		button = styles.RenderFocusedButton("Submit")
	}

	sections = append(sections, button)
	sections = append(sections, t.renderHelpText())

	return styles.FormContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
}

// renderHelpText renders context-sensitive help text based on the focused field.
func (t *TaskModel) renderHelpText() string {
	switch t.nav.FocusIndex {
	case nameField, cmdsField:
		return styles.HelpTextStyle.Render("up/down/tab/shift+tab navigate • esc quit")
	case pathField, iconField, iconColorField:
		return styles.HelpTextStyle.Render("up/down navigate • ctrl+b/n cycle suggestions • tab/enter apply • esc quit")
	default:
		return styles.HelpTextStyle.Render("enter submit • esc quit")
	}
}

// getCurrentSuggestionManager returns the suggestion manager for icon/color fields.
// Path suggestions are handled directly via t.pathSuggestions at each call site.
func (t *TaskModel) getCurrentSuggestionManager() *suggestions.Manager {
	switch t.nav.FocusIndex {
	case iconField:
		return t.iconSuggestions
	case iconColorField:
		return t.colorSuggestions
	default:
		return nil
	}
}
