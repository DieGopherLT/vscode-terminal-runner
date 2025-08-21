package workspace

import (
	"fmt"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/workspace/components"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/messages"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	nameField     = 0 // Workspace name field
	taskListField = 1 // Task selector field
)

var noStyle = lipgloss.NewStyle()

// WorkspaceModel represents the form for creating/editing workspaces.
type WorkspaceModel struct {
	nav                   *tui.FormNavigator
	nameInput            textinput.Model
	taskSelector         *components.TaskSelector
	messages             *messages.MessageManager
	isEditMode           bool
	originalWorkspaceName string
}

// NewWorkspaceModel creates a new workspace creation form.
func NewWorkspaceModel() tea.Model {
	return newWorkspaceModelInternal(nil)
}

// NewEditWorkspaceModel creates a workspace editing form with pre-filled data.
func NewEditWorkspaceModel(workspace *models.Workspace) tea.Model {
	return newWorkspaceModelInternal(workspace)
}

// newWorkspaceModelInternal creates the internal workspace model with optional existing workspace data.
func newWorkspaceModelInternal(workspace *models.Workspace) *WorkspaceModel {
	// Initialize form navigator with 2 fields (name, tasks) + submit handled separately
	nav := tui.NewNavigator(2)

	// Setup name input
	nameInput := textinput.New()
	nameInput.Placeholder = "Enter workspace name..."
	nameInput.Focus()
	nameInput.CharLimit = 50
	nameInput.Width = 90

	// Get all available tasks with proper error handling
	availableTasks := getAvailableTasks()

	// Initialize task selector
	taskSelector := components.NewTaskSelector(availableTasks)

	// Setup edit mode if workspace is provided
	isEditMode := workspace != nil
	originalWorkspaceName := ""

	if isEditMode {
		originalWorkspaceName = workspace.Name
		nameInput.SetValue(workspace.Name)
		taskSelector.SetSelectedTasks(workspace.Tasks)
	}

	return &WorkspaceModel{
		nav:                   nav,
		nameInput:            nameInput,
		taskSelector:         taskSelector,
		messages:             messages.NewManager(),
		isEditMode:           isEditMode,
		originalWorkspaceName: originalWorkspaceName,
	}
}

// Init initializes the workspace form model.
func (w *WorkspaceModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the workspace form state.
func (w *WorkspaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Update the appropriate component based on focus first
	if w.nav.FocusIndex == nameField {
		var cmd tea.Cmd
		w.nameInput, cmd = w.nameInput.Update(msg)
		w.clearMessagesOnInput()
		
		// Check if this is a key message for navigation
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if w.isNavigationKey(keyMsg.String()) {
				return w.handleKeyPress(keyMsg)
			}
		}
		return w, cmd
	} else if w.nav.FocusIndex == taskListField {
		// If in search mode, let the task selector handle input first
		if w.taskSelector.IsInSearchMode() {
			// Let search input handle typing, but check for navigation keys
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				key := keyMsg.String()
				// Only handle navigation keys, let everything else go to search input
				if key == "esc" || key == "enter" {
					return w.handleKeyPress(keyMsg)
				}
			}
			cmd := w.taskSelector.Update(msg)
			return w, cmd
		} else {
			// Handle task selector update and key navigation
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				return w.handleKeyPress(keyMsg)
			}
			cmd := w.taskSelector.Update(msg)
			return w, cmd
		}
	}

	// Handle key messages for other cases
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return w.handleKeyPress(msg)
	}

	return w, nil
}

// isNavigationKey checks if the key is for navigation between fields.
func (w *WorkspaceModel) isNavigationKey(key string) bool {
	navKeys := []string{
		string(tui.KeyUp), string(tui.KeyDown), 
		string(tui.KeyTab), string(tui.KeyShiftTab),
		"enter", "ctrl+c", "esc",
	}
	
	for _, navKey := range navKeys {
		if key == navKey {
			return true
		}
	}
	return false
}

// handleKeyPress processes keyboard input for navigation and actions.
func (w *WorkspaceModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c", "esc":
		// Exit search mode if active, otherwise quit
		if w.nav.FocusIndex == taskListField && w.taskSelector.IsInSearchMode() {
			w.taskSelector.ToggleSearch()
			return w, nil
		}
		return w, tea.Quit

	case string(tui.KeyUp), string(tui.KeyDown):
		return w.handleVerticalNavigation(key)

	case string(tui.KeyTab), string(tui.KeyShiftTab):
		return w.handleTabNavigation(key)

	case "enter":
		return w.handleEnterKey()

	case " ":
		return w.handleSpaceKey()

	case "/":
		return w.handleSearchToggle()

	case "ctrl+a":
		return w.handleSelectAll()

	case "ctrl+d":
		return w.handleDeselectAll()
	}

	return w, nil
}

// handleVerticalNavigation processes up/down arrow keys.
func (w *WorkspaceModel) handleVerticalNavigation(key string) (tea.Model, tea.Cmd) {
	if w.nav.FocusIndex == taskListField && !w.taskSelector.IsInSearchMode() {
		// Navigate within task list
		direction := -1 // Up
		if key == string(tui.KeyDown) {
			direction = 1 // Down
		}
		w.taskSelector.MoveFocus(direction)
		return w, nil
	}

	// Regular form navigation between fields
	w.nav.HandleNavigation(key)
	return w.handleFocus()
}

// handleTabNavigation processes tab and shift+tab keys.
func (w *WorkspaceModel) handleTabNavigation(key string) (tea.Model, tea.Cmd) {
	w.nav.HandleNavigation(key)
	return w.handleFocus()
}

// handleEnterKey processes enter key based on current context.
func (w *WorkspaceModel) handleEnterKey() (tea.Model, tea.Cmd) {
	if w.nav.FocusIndex == taskListField {
		if w.taskSelector.IsInSearchMode() {
			// Exit search mode
			w.taskSelector.ToggleSearch()
			return w, nil
		}
		// Move to submit button (outside navigator)
		w.nav.FocusIndex = w.nav.GetElementCount()
		return w.handleFocus()
	}

	if w.nav.FocusIndex >= w.nav.GetElementCount() {
		// Submit the workspace
		return w.handleSubmit()
	}

	// Move to next field
	w.nav.HandleNavigation("down")
	return w.handleFocus()
}

// handleSpaceKey processes space key for task selection.
func (w *WorkspaceModel) handleSpaceKey() (tea.Model, tea.Cmd) {
	if w.nav.FocusIndex == taskListField && !w.taskSelector.IsInSearchMode() {
		w.taskSelector.ToggleSelected()
	}
	return w, nil
}

// handleSearchToggle toggles search mode in task selector.
func (w *WorkspaceModel) handleSearchToggle() (tea.Model, tea.Cmd) {
	if w.nav.FocusIndex == taskListField {
		w.taskSelector.ToggleSearch()
	}
	return w, nil
}

// handleSelectAll selects all visible tasks.
func (w *WorkspaceModel) handleSelectAll() (tea.Model, tea.Cmd) {
	if w.nav.FocusIndex == taskListField {
		w.taskSelector.SelectAll()
	}
	return w, nil
}

// handleDeselectAll deselects all tasks.
func (w *WorkspaceModel) handleDeselectAll() (tea.Model, tea.Cmd) {
	if w.nav.FocusIndex == taskListField {
		w.taskSelector.DeselectAll()
	}
	return w, nil
}

// handleFocus updates the focus state of form components.
func (w *WorkspaceModel) handleFocus() (tea.Model, tea.Cmd) {
	switch w.nav.FocusIndex {
	case nameField:
		w.nameInput.Focus()
		w.nameInput.PromptStyle = styles.FocusedInputStyle
		w.nameInput.TextStyle = styles.FocusedInputStyle
	default:
		w.nameInput.Blur()
		w.nameInput.PromptStyle = noStyle
		w.nameInput.TextStyle = noStyle
	}

	return w, nil
}

// handleSubmit processes workspace creation/update.
func (w *WorkspaceModel) handleSubmit() (tea.Model, tea.Cmd) {
	workspace := w.createWorkspaceFromForm()

	if !w.isValidWorkspace(workspace) {
		return w, nil
	}

	if err := w.saveWorkspace(workspace); err != nil {
		w.messages.AddError(fmt.Sprintf("Failed to save workspace: %v", err))
		return w, nil
	}

	successMessage := "Workspace created successfully!"
	if w.isEditMode {
		successMessage = "Workspace updated successfully!"
	}
	w.messages.AddSuccess(successMessage)

	return w, tea.Quit
}

// createWorkspaceFromForm creates a workspace model from current form state.
func (w *WorkspaceModel) createWorkspaceFromForm() models.Workspace {
	return models.Workspace{
		Name:  strings.TrimSpace(w.nameInput.Value()),
		Tasks: w.taskSelector.GetSelectedTasks(),
	}
}

// isValidWorkspace validates the workspace data and shows appropriate messages.
func (w *WorkspaceModel) isValidWorkspace(workspace models.Workspace) bool {
	w.messages.Clear()

	if workspace.Name == "" {
		w.messages.AddError("Workspace name is required")
		w.nav.FocusIndex = nameField
		w.handleFocus()
		return false
	}

	// Check for duplicate workspace names
	if !w.isValidWorkspaceName(workspace.Name) {
		return false
	}

	if len(workspace.Tasks) == 0 {
		w.messages.AddWarning("No tasks selected. Workspace will be empty.")
		// Allow empty workspaces but warn user
	}

	return !w.messages.HasErrors()
}

// saveWorkspace saves the workspace to the repository.
func (w *WorkspaceModel) saveWorkspace(workspace models.Workspace) error {
	if w.isEditMode {
		// Delete old workspace if name changed
		if workspace.Name != w.originalWorkspaceName {
			if err := repository.DeleteWorkspace(w.originalWorkspaceName); err != nil {
				return fmt.Errorf("failed to delete old workspace: %w", err)
			}
		}
	}

	return repository.SaveWorkspace(workspace)
}

// clearMessagesOnInput clears messages when user starts typing.
func (w *WorkspaceModel) clearMessagesOnInput() {
	if w.messages.HasMessages() {
		w.messages.Clear()
	}
}

// View renders the workspace form.
func (w *WorkspaceModel) View() string {
	var sections []string

	// Title
	title := "CREATE WORKSPACE"
	if w.isEditMode {
		title = "EDIT WORKSPACE"
	}
	sections = append(sections, styles.RenderTitle(title))

	// Workspace name field
	nameFieldContent := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.FieldLabelStyle.Render("Workspace Name:"),
		w.nameInput.View(),
	)
	sections = append(sections, styles.FieldContainerStyle.Render(nameFieldContent))

	// Task selector field
	sections = append(sections, w.taskSelector.View())

	// Messages
	if w.messages.HasMessages() {
		sections = append(sections, w.messages.Render())
	}

	// Submit button
	button := styles.RenderBlurredButton("Submit")
	if w.nav.FocusIndex >= w.nav.GetElementCount() {
		button = styles.RenderFocusedButton("Submit")
	}
	sections = append(sections, button)

	// Help text
	helpText := w.renderHelpText()
	sections = append(sections, helpText)

	return styles.FormContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
}

// renderHelpText renders context-sensitive help text.
func (w *WorkspaceModel) renderHelpText() string {
	if w.nav.FocusIndex == taskListField && w.taskSelector.IsInSearchMode() {
		return styles.HelpTextStyle.Render("type to search • esc exit search • enter confirm")
	}

	if w.nav.FocusIndex == taskListField {
		return styles.HelpTextStyle.Render("↑/↓ navigate • space toggle • /search • ctrl+a/d select/deselect all • tab/shift+tab navigate")
	}

	return styles.HelpTextStyle.Render("↑/↓/tab/shift+tab navigate • enter submit • esc quit")
}

// getAvailableTasks retrieves all available tasks with proper error handling.
func getAvailableTasks() []models.Task {
	availableTasks, err := repository.GetAllTasks()
	if err != nil {
		// Return empty slice on error - the UI will show "No tasks available" message
		return []models.Task{}
	}
	return availableTasks
}

// isValidWorkspaceName checks if workspace name is valid and available using guard clauses.
func (w *WorkspaceModel) isValidWorkspaceName(workspaceName string) bool {
	// Skip validation if editing with same name
	isEditingWithSameName := w.isEditMode && workspaceName == w.originalWorkspaceName
	if isEditingWithSameName {
		return true
	}
	
	_, err := repository.FindWorkspaceByName(workspaceName)
	if err == nil {
		w.messages.AddError("Workspace name already exists")
		w.nav.FocusIndex = nameField
		w.handleFocus()
		return false
	}
	
	return true
}