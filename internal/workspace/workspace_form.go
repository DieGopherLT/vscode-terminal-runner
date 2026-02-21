package workspace

import (
	"fmt"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/workspace/components"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/messages"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	nameField     = 0
	taskListField = 1
)

// WorkspaceModel represents the form for creating/editing workspaces.
type WorkspaceModel struct {
	nav                   *tui.FormNavigator
	nameInput             textinput.Model
	taskSelector          *components.TaskSelector
	messages              *messages.MessageManager
	isEditMode            bool
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
	nav := tui.NewNavigator(2)

	nameInput := textinput.New()
	nameInput.Placeholder = "Enter workspace name..."
	nameInput.Focus()
	nameInput.CharLimit = 50
	nameInput.Width = 90

	availableTasks, err := getAvailableTasks()
	if err != nil {
		availableTasks = []models.Task{}
	}
	taskSelector := components.NewTaskSelector(availableTasks)

	isEditMode := workspace != nil
	originalWorkspaceName := ""

	if isEditMode {
		originalWorkspaceName = workspace.Name
		nameInput.SetValue(workspace.Name)
		taskSelector.SetSelectedTasks(workspace.Tasks)
	}

	return &WorkspaceModel{
		nav:                   nav,
		nameInput:             nameInput,
		taskSelector:          taskSelector,
		messages:              messages.NewManager(),
		isEditMode:            isEditMode,
		originalWorkspaceName: originalWorkspaceName,
	}
}

// Init initializes the workspace form model.
func (w *WorkspaceModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the workspace form state.
func (w *WorkspaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.nameInput.Width = msg.Width - 10
		return w, nil

	case workspaceSaveResultMsg:
		if msg.err != nil {
			w.messages.AddError(fmt.Sprintf("Failed to save workspace: %v", msg.err))
			return w, nil
		}
		successMessage := "Workspace created successfully!"
		if w.isEditMode {
			successMessage = "Workspace updated successfully!"
		}
		w.messages.AddSuccess(successMessage)
		return w, tea.Quit
	}

	if w.nav.FocusIndex == nameField {
		var cmd tea.Cmd
		w.nameInput, cmd = w.nameInput.Update(msg)
		w.clearMessagesOnInput()

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if w.isNavigationKey(keyMsg) {
				return w.handleKeyPress(keyMsg)
			}
		}
		return w, cmd
	}

	if w.nav.FocusIndex == taskListField {
		if w.taskSelector.IsInSearchMode() {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				if key.Matches(keyMsg, tui.DefaultKeys.Quit) || key.Matches(keyMsg, tui.DefaultKeys.Enter) {
					return w.handleKeyPress(keyMsg)
				}
			}
			cmd := w.taskSelector.Update(msg)
			return w, cmd
		}

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return w.handleKeyPress(keyMsg)
		}
		cmd := w.taskSelector.Update(msg)
		return w, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return w.handleKeyPress(keyMsg)
	}

	return w, nil
}

// isNavigationKey checks if the key is for navigation between fields.
func (w *WorkspaceModel) isNavigationKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, tui.DefaultKeys.Up) ||
		key.Matches(msg, tui.DefaultKeys.Down) ||
		key.Matches(msg, tui.DefaultKeys.Tab) ||
		key.Matches(msg, tui.DefaultKeys.ShiftTab) ||
		key.Matches(msg, tui.DefaultKeys.Enter) ||
		key.Matches(msg, tui.DefaultKeys.Quit)
}

// handleKeyPress processes keyboard input for navigation and actions.
func (w *WorkspaceModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, tui.DefaultKeys.Quit):
		if w.nav.FocusIndex == taskListField && w.taskSelector.IsInSearchMode() {
			w.taskSelector.ToggleSearch()
			return w, nil
		}
		return w, tea.Quit

	case key.Matches(msg, tui.DefaultKeys.Up), key.Matches(msg, tui.DefaultKeys.Down):
		return w.handleVerticalNavigation(msg)

	case key.Matches(msg, tui.DefaultKeys.Tab), key.Matches(msg, tui.DefaultKeys.ShiftTab):
		return w.handleTabNavigation(msg)

	case key.Matches(msg, tui.DefaultKeys.Enter):
		return w.handleEnterKey()

	case key.Matches(msg, tui.DefaultKeys.Space):
		return w.handleSpaceKey()

	case key.Matches(msg, tui.DefaultKeys.Search):
		return w.handleSearchToggle()

	case key.Matches(msg, tui.DefaultKeys.SelectAll):
		return w.handleSelectAll()

	case key.Matches(msg, tui.DefaultKeys.DeselectAll):
		return w.handleDeselectAll()
	}

	return w, nil
}

// handleVerticalNavigation processes up/down arrow keys.
func (w *WorkspaceModel) handleVerticalNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if w.nav.FocusIndex == taskListField && !w.taskSelector.IsInSearchMode() {
		direction := -1
		if key.Matches(msg, tui.DefaultKeys.Down) {
			direction = 1
		}
		w.taskSelector.MoveFocus(direction)
		return w, nil
	}

	w.nav.HandleNavigation(msg.String())
	w.applyFocus()
	return w, nil
}

// handleTabNavigation processes tab and shift+tab keys.
func (w *WorkspaceModel) handleTabNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	w.nav.HandleNavigation(msg.String())
	w.applyFocus()
	return w, nil
}

// handleEnterKey processes enter key based on current context.
func (w *WorkspaceModel) handleEnterKey() (tea.Model, tea.Cmd) {
	if w.nav.FocusIndex == taskListField {
		if w.taskSelector.IsInSearchMode() {
			w.taskSelector.ToggleSearch()
			return w, nil
		}
		w.nav.FocusIndex = w.nav.GetElementCount()
		w.applyFocus()
		return w, nil
	}

	if w.nav.FocusIndex >= w.nav.GetElementCount() {
		return w.handleSubmit()
	}

	w.nav.HandleNavigation("down")
	w.applyFocus()
	return w, nil
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

// applyFocus updates the visual focus state of form components.
func (w *WorkspaceModel) applyFocus() {
	switch w.nav.FocusIndex {
	case nameField:
		w.nameInput.Focus()
		w.nameInput.PromptStyle = styles.FocusedInputStyle
		w.nameInput.TextStyle = styles.FocusedInputStyle
	default:
		w.nameInput.Blur()
		w.nameInput.PromptStyle = styles.UnfocusedInputStyle
		w.nameInput.TextStyle = styles.UnfocusedInputStyle
	}
}

// handleSubmit validates and kicks off async workspace persistence.
func (w *WorkspaceModel) handleSubmit() (tea.Model, tea.Cmd) {
	workspace := w.createWorkspaceFromForm()

	if !w.isValidWorkspace(workspace) {
		return w, nil
	}

	return w, submitWorkspaceCmd(w, workspace)
}

// submitWorkspaceCmd performs duplicate-check and save off the UI goroutine.
func submitWorkspaceCmd(w *WorkspaceModel, workspace models.Workspace) tea.Cmd {
	isEditMode := w.isEditMode
	originalName := w.originalWorkspaceName

	return func() tea.Msg {
		// Duplicate check (file I/O off UI thread)
		isRenamingOrCreating := !isEditMode || workspace.Name != originalName
		if isRenamingOrCreating {
			_, err := repository.FindWorkspaceByName(workspace.Name)
			if err == nil {
				return workspaceSaveResultMsg{err: fmt.Errorf("workspace name already exists")}
			}
		}

		// Save first
		if err := repository.SaveWorkspace(workspace); err != nil {
			return workspaceSaveResultMsg{err: err}
		}

		// Only delete old record after save succeeds
		if isEditMode && workspace.Name != originalName {
			if err := repository.DeleteWorkspace(originalName); err != nil {
				return workspaceSaveResultMsg{err: fmt.Errorf("saved workspace but failed to remove old record '%s': %w", originalName, err)}
			}
		}

		return workspaceSaveResultMsg{}
	}
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
		w.applyFocus()
		return false
	}

	if len(workspace.Tasks) == 0 {
		w.messages.AddWarning("No tasks selected. Workspace will be empty.")
	}

	return !w.messages.HasErrors()
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

	title := "CREATE WORKSPACE"
	if w.isEditMode {
		title = "EDIT WORKSPACE"
	}
	sections = append(sections, styles.RenderTitle(title))

	nameFieldContent := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.FieldLabelStyle.Render("Workspace Name:"),
		w.nameInput.View(),
	)
	sections = append(sections, styles.FieldContainerStyle.Render(nameFieldContent))

	sections = append(sections, w.taskSelector.View())

	if w.messages.HasMessages() {
		sections = append(sections, w.messages.Render())
	}

	button := styles.RenderBlurredButton("Submit")
	if w.nav.FocusIndex >= w.nav.GetElementCount() {
		button = styles.RenderFocusedButton("Submit")
	}
	sections = append(sections, button)

	sections = append(sections, w.renderHelpText())

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
		return styles.HelpTextStyle.Render("up/down navigate • space toggle • /search • ctrl+a/d select/deselect all • tab/shift+tab navigate")
	}

	return styles.HelpTextStyle.Render("up/down/tab/shift+tab navigate • enter submit • esc quit")
}

// getAvailableTasks retrieves all available tasks with proper error handling.
func getAvailableTasks() ([]models.Task, error) {
	availableTasks, err := repository.GetAllTasks()
	if err != nil {
		return nil, fmt.Errorf("getAvailableTasks: %w", err)
	}
	return availableTasks, nil
}
