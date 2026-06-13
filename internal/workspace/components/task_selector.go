package components

import (
	"fmt"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/collections"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

// UI Layout constants for grid-like alignment
const (
	maxSearchInputLength = 50 // Maximum characters in search input
	maxVisibleTasks      = 6  // Maximum tasks visible at once
	maxPathDisplayLength = 50 // Maximum characters displayed for path
	taskNameColumnWidth  = 25 // Fixed width for task name column

	defaultSelectorWidth     = 90 // Container width when the terminal size is unknown
	selectorHorizontalMargin = 4  // Reserved space for form padding and container border
)

// TaskSelector provides multi-select functionality for tasks with search capabilities.
type TaskSelector struct {
	availableTasks []models.Task
	selectedTasks  *collections.Set[string]
	filteredTasks  []models.Task
	focusedIndex   int
	searchInput    textinput.Model
	isInSearchMode bool
	maxHeight      int
	width          int // Terminal width; 0 means unset (use defaults)
}

// NewTaskSelector creates a new task selector with the given available tasks.
func NewTaskSelector(availableTasks []models.Task) *TaskSelector {
	searchInput := textinput.New()
	searchInput.Placeholder = "Search tasks..."
	searchInput.CharLimit = maxSearchInputLength

	return &TaskSelector{
		availableTasks: availableTasks,
		selectedTasks:  collections.NewSet[string](),
		filteredTasks:  availableTasks,
		focusedIndex:   0,
		searchInput:    searchInput,
		isInSearchMode: false,
		maxHeight:      maxVisibleTasks,
	}
}

// SetWidth stores the terminal width so the selector container, search input,
// and separator can shrink to fit narrow terminals.
func (ts *TaskSelector) SetWidth(width int) {
	ts.width = width
}

// containerWidth returns the selector container width, clamped to the terminal.
func (ts *TaskSelector) containerWidth() int {
	return styles.ResponsiveContainerWidth(ts.width, defaultSelectorWidth, selectorHorizontalMargin)
}

// GetSelectedTasks returns a slice of currently selected tasks.
func (ts *TaskSelector) GetSelectedTasks() []models.Task {
	return lo.Filter(ts.availableTasks, func(task models.Task, _ int) bool {
		return ts.selectedTasks.Contains(task.Name)
	})
}

// SetSelectedTasks sets the initially selected tasks.
func (ts *TaskSelector) SetSelectedTasks(tasks []models.Task) {
	ts.selectedTasks = collections.NewSet[string]()
	for _, task := range tasks {
		ts.selectedTasks.Add(task.Name)
	}
}

// GetSelectedCount returns the number of currently selected tasks.
func (ts *TaskSelector) GetSelectedCount() int {
	return ts.selectedTasks.Len()
}

// ToggleSearch toggles the search input visibility and focus.
func (ts *TaskSelector) ToggleSearch() {
	ts.isInSearchMode = !ts.isInSearchMode
	if ts.isInSearchMode {
		ts.searchInput.Focus()
	} else {
		ts.searchInput.Blur()
		ts.searchInput.SetValue("")
		ts.filteredTasks = ts.availableTasks
		ts.focusedIndex = 0
	}
}

// IsInSearchMode returns true if the search mode is active.
func (ts *TaskSelector) IsInSearchMode() bool {
	return ts.isInSearchMode
}

// SelectAll selects all currently visible (filtered) tasks.
func (ts *TaskSelector) SelectAll() {
	for _, task := range ts.filteredTasks {
		ts.selectedTasks.Add(task.Name)
	}
}

// DeselectAll deselects all currently visible (filtered) tasks.
func (ts *TaskSelector) DeselectAll() {
	for _, task := range ts.filteredTasks {
		ts.selectedTasks.Remove(task.Name)
	}
}

// ToggleSelected toggles the selection state of the currently focused task.
func (ts *TaskSelector) ToggleSelected() {
	if len(ts.filteredTasks) == 0 {
		return
	}

	if ts.focusedIndex >= 0 && ts.focusedIndex < len(ts.filteredTasks) {
		task := ts.filteredTasks[ts.focusedIndex]
		ts.selectedTasks.Toggle(task.Name)
	}
}

// MoveFocus moves the focus up or down within the visible task list.
func (ts *TaskSelector) MoveFocus(direction int) {
	if len(ts.filteredTasks) == 0 {
		return
	}

	ts.focusedIndex += direction

	if ts.focusedIndex < 0 {
		ts.focusedIndex = len(ts.filteredTasks) - 1
	} else if ts.focusedIndex >= len(ts.filteredTasks) {
		ts.focusedIndex = 0
	}
}

// UpdateFilter updates the task filter based on search input.
func (ts *TaskSelector) UpdateFilter() {
	query := strings.ToLower(strings.TrimSpace(ts.searchInput.Value()))

	if query == "" {
		ts.filteredTasks = ts.availableTasks
	} else {
		ts.filteredTasks = lo.Filter(ts.availableTasks, func(task models.Task, _ int) bool {
			nameMatch := strings.Contains(strings.ToLower(task.Name), query)
			pathMatch := strings.Contains(strings.ToLower(task.Path), query)
			return nameMatch || pathMatch
		})
	}

	// Reset focus to first item after filtering
	ts.focusedIndex = 0
}

// Update handles updates for the task selector component.
func (ts *TaskSelector) Update(msg tea.Msg) tea.Cmd {
	if ts.isInSearchMode {
		var cmd tea.Cmd
		ts.searchInput, cmd = ts.searchInput.Update(msg)
		ts.UpdateFilter()
		return cmd
	}
	return nil
}

// View renders the task selector component.
func (ts *TaskSelector) View() string {
	if len(ts.availableTasks) == 0 {
		return ts.renderEmptyState()
	}

	var sections []string

	containerWidth := ts.containerWidth()

	// Header with counter
	selectedCount := ts.GetSelectedCount()
	totalTasks := len(ts.availableTasks)
	header := fmt.Sprintf("Select Tasks:                      [%d/%d selected]", selectedCount, totalTasks)

	// Search box (if enabled)
	if ts.isInSearchMode {
		searchBox := styles.TextInputStyle.Width(containerWidth - 4).Render(ts.searchInput.View())
		sections = append(sections, searchBox)
		sections = append(sections, strings.Repeat("─", containerWidth-2))
	}

	// Task list
	taskList := ts.renderTaskList()
	sections = append(sections, taskList)

	// Help text is rendered by the parent form (WorkspaceModel.renderHelpText)
	// to avoid duplicating the shortcut hints inside and outside the container.

	// Combine all sections
	content := strings.Join(sections, "\n")

	// Container with header
	container := fmt.Sprintf("%s\n%s",
		styles.LightGrayStyle.Render(header),
		styles.TaskSelectorContainerStyle.Width(containerWidth).Render(content),
	)

	return container
}

// renderTaskList renders the scrollable list of tasks.
func (ts *TaskSelector) renderTaskList() string {
	var items []string

	if len(ts.filteredTasks) == 0 {
		items = append(items, styles.LightGrayStyle.Render("No tasks match your search."))
	} else {
		startIndex := 0
		endIndex := len(ts.filteredTasks)

		// Handle scrolling for long lists
		if len(ts.filteredTasks) > ts.maxHeight {
			if ts.focusedIndex >= ts.maxHeight {
				startIndex = ts.focusedIndex - ts.maxHeight + 1
			}
			endIndex = startIndex + ts.maxHeight
			if endIndex > len(ts.filteredTasks) {
				endIndex = len(ts.filteredTasks)
			}
		}

		for i := startIndex; i < endIndex; i++ {
			item := ts.renderTaskItem(ts.filteredTasks[i], i)
			items = append(items, item)
		}
	}

	// Pad to a stable height so the bordered container does not resize (and clip)
	// as the filtered list grows or shrinks.
	for len(items) < ts.maxHeight {
		items = append(items, "")
	}

	return strings.Join(items, "\n")
}

// renderTaskItem renders a single task item with checkbox and styling.
func (ts *TaskSelector) renderTaskItem(task models.Task, index int) string {
	// Checkbox state
	checkbox := "☐"
	if ts.selectedTasks.Contains(task.Name) {
		checkbox = "☑"
	}

	// Focus indicator and styling
	focusPrefix := "  "
	itemStyle := styles.UnfocusedTaskStyle

	if index == ts.focusedIndex {
		focusPrefix = "▶ "
		itemStyle = styles.FocusedTaskStyle
	} else if ts.selectedTasks.Contains(task.Name) {
		itemStyle = styles.SelectedTaskStyle
	}

	// Truncate path if too long
	displayPath := truncatePath(task.Path, maxPathDisplayLength)

	// Format item text with fixed-width columns (grid-like alignment)
	// Focus(2) + Checkbox(2) + TaskName(25) + Separator(7) + Path(remaining)
	paddedTaskName := padRight(task.Name, taskNameColumnWidth)
	itemText := fmt.Sprintf("%s%s %s [path] %s",
		focusPrefix, checkbox, paddedTaskName, displayPath)

	return itemStyle.Render(itemText)
}

// renderEmptyState renders the state when no tasks are available.
func (ts *TaskSelector) renderEmptyState() string {
	content := styles.LightGrayStyle.Render("No tasks available.\nCreate some tasks first to add them to workspaces.")
	return styles.TaskSelectorContainerStyle.Width(ts.containerWidth()).Render(content)
}

// truncatePath truncates a path to fit within the specified width, adding ellipsis if needed.
// Width is measured in runes so multi-byte characters do not break column alignment.
func truncatePath(path string, maxWidth int) string {
	runes := []rune(path)
	if len(runes) <= maxWidth {
		return path
	}

	if maxWidth <= 3 {
		return "..."
	}

	return "..." + string(runes[len(runes)-maxWidth+3:])
}

// padRight pads a string to a fixed width using spaces, truncating with ellipsis if necessary.
// Width is measured in runes so multi-byte characters do not break column alignment.
func padRight(text string, width int) string {
	runes := []rune(text)
	if len(runes) > width {
		return string(runes[:width-3]) + "..."
	}

	paddingNeeded := width - len(runes)
	if paddingNeeded <= 0 {
		return text
	}

	return text + strings.Repeat(" ", paddingNeeded)
}
