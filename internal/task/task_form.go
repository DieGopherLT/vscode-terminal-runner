package task

import (
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/vscode"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
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
			
			// Si hay sugerencias y Tab, aplicar sugerencia
			if key == "tab" && t.hasSuggestions() {
				t.applySuggestion()
				return t, nil
			}
			
			t.nav.HandleNavigation(key)
			t.suggestionIndex = 0  // Reset suggestion index when navigating
			return t.HandleFocus()

		case "ctrl+n":
			// Navegar hacia abajo en sugerencias (circular)
			if t.hasSuggestions() {
				suggestions := t.getVisibleSuggestions()
				t.suggestionIndex = (t.suggestionIndex + 1) % len(suggestions)
			}
			return t, nil

		case "ctrl+b":
			// Navegar hacia arriba en sugerencias (circular)
			if t.hasSuggestions() {
				suggestions := t.getVisibleSuggestions()
				t.suggestionIndex = (t.suggestionIndex - 1 + len(suggestions)) % len(suggestions)
			}
			return t, nil

		case "enter":
			// Si hay sugerencias, aplicar la seleccionada
			if t.hasSuggestions() {
				t.applySuggestion()
				return t, nil
			}
			
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

	// Solo procesar input si no es navegación de sugerencias
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

	// Store previous values to detect actual changes
	var previousValues []string
	for i := range t.inputs {
		previousValues = append(previousValues, t.inputs[i].Value())
	}

	for i := range t.inputs {
		t.inputs[i], cmds[i] = t.inputs[i].Update(msg)

		// Only reset suggestion index if content actually changed
		contentChanged := t.inputs[i].Value() != previousValues[i]

		if i == iconField {
			iconsNames := lo.Map(vscode.Icons, func(i vscode.Icon, _ int) string { return i.Name })
			suggestions := lo.Filter(iconsNames, func(icon string, _ int) bool {
				return strings.Contains(strings.ToLower(icon), strings.ToLower(t.inputs[i].Value()))
			})
			t.inputs[i].SetSuggestions(suggestions)
			
			// Reset suggestion index only when input content changes
			if i == t.nav.FocusIndex && contentChanged {
				t.suggestionIndex = 0
			}
		}

		if i == iconColorField {
			colorsNames := lo.Map(vscode.ANSIColors, func(c vscode.ANSIColor, _ int) string { return c.Name })
			suggestions := lo.Filter(colorsNames, func(color string, _ int) bool {
				return strings.Contains(strings.ToLower(color), strings.ToLower(t.inputs[i].Value()))
			})
			t.inputs[i].SetSuggestions(suggestions)
			
			// Reset suggestion index only when input content changes
			if i == t.nav.FocusIndex && contentChanged {
				t.suggestionIndex = 0
			}
		}
	}

	return tea.Batch(cmds...)
}

// hasSuggestions checks if the current focused input has suggestions available
func (t *TaskModel) hasSuggestions() bool {
	if t.nav.FocusIndex >= len(t.inputs) {
		return false
	}
	
	currentInput := t.inputs[t.nav.FocusIndex]
	if !currentInput.ShowSuggestions || len(currentInput.AvailableSuggestions()) == 0 {
		return false
	}
	
	// If there's only one suggestion and it's an exact match, don't show suggestions
	suggestions := currentInput.AvailableSuggestions()
	if len(suggestions) == 1 && suggestions[0] == currentInput.Value() {
		return false
	}
	
	return true
}

// getVisibleSuggestions returns the limited suggestions that are actually shown in the UI
func (t *TaskModel) getVisibleSuggestions() []string {
	if !t.hasSuggestions() {
		return nil
	}
	
	suggestions := t.inputs[t.nav.FocusIndex].AvailableSuggestions()
	maxSuggestions := 3
	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}
	return suggestions
}

// applySuggestion applies the currently selected suggestion to the focused input
func (t *TaskModel) applySuggestion() {
	suggestions := t.getVisibleSuggestions()
	if len(suggestions) == 0 {
		return
	}
	
	if t.suggestionIndex < len(suggestions) {
		selectedSuggestion := suggestions[t.suggestionIndex]
		t.inputs[t.nav.FocusIndex].SetValue(selectedSuggestion)
		t.suggestionIndex = 0
	}
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
		
		// Show suggestions for inputs that have ShowSuggestions enabled
		if t.inputs[i].ShowSuggestions && len(t.inputs[i].AvailableSuggestions()) > 0 && t.nav.FocusIndex == i {
			suggestions := t.getVisibleSuggestions()
			
			var suggestionLines []string
			for j, suggestion := range suggestions {
				if j == t.suggestionIndex {
					// Highlight selected suggestion
					suggestionLines = append(suggestionLines, styles.SuggestionHighlightStyle.Render("• "+suggestion))
				} else {
					suggestionLines = append(suggestionLines, styles.SuggestionItemStyle.Render("• "+suggestion))
				}
			}
			
			suggestionBox := styles.SuggestionContainerStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left, suggestionLines...),
			)
			
			fieldContent = lipgloss.JoinVertical(
				lipgloss.Left,
				fieldContent,
				suggestionBox,
			)
		}
		
		sections = append(sections, styles.FieldContainerStyle.Render(fieldContent))
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
