package components

import (
	"fmt"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

// UI Layout constants for grid-like alignment
const (
	maxSearchInputLength = 50  // Maximum characters in search input
	maxVisibleTasks     = 6   // Maximum tasks visible at once
	maxPathDisplayLength = 50  // Maximum characters displayed for path
	taskNameColumnWidth  = 25  // Fixed width for task name column
	separatorLineLength  = 88  // Length of separator line in search mode
)

// TaskSelector provides multi-select functionality for tasks with search capabilities.
type TaskSelector struct {
	availableTasks    []models.Task
	selectedTasks     map[string]bool
	filteredTasks     []models.Task
	focusedIndex      int
	searchInput       textinput.Model
	showSearch        bool
	maxHeight         int
}

// NewTaskSelector creates a new task selector with the given available tasks.
func NewTaskSelector(availableTasks []models.Task) *TaskSelector {
	searchInput := textinput.New()
	searchInput.Placeholder = "Search tasks..."
	searchInput.CharLimit = maxSearchInputLength

	return &TaskSelector{
		availableTasks: availableTasks,
		selectedTasks:  make(map[string]bool),
		filteredTasks:  availableTasks,
		focusedIndex:   0,
		searchInput:    searchInput,
		showSearch:     false,
		maxHeight:      maxVisibleTasks,
	}
}

// GetSelectedTasks returns a slice of currently selected tasks.
func (ts *TaskSelector) GetSelectedTasks() []models.Task {
	return lo.Filter(ts.availableTasks, func(task models.Task, _ int) bool {
		return ts.selectedTasks[task.Name]
	})
}

// SetSelectedTasks sets the initially selected tasks.
func (ts *TaskSelector) SetSelectedTasks(tasks []models.Task) {
	ts.selectedTasks = make(map[string]bool)
	for _, task := range tasks {
		ts.selectedTasks[task.Name] = true
	}
}

// GetSelectedCount returns the number of currently selected tasks using functional approach.
func (ts *TaskSelector) GetSelectedCount() int {
	return lo.CountBy(lo.Values(ts.selectedTasks), func(isSelected bool) bool {
		return isSelected
	})
}

// ToggleSearch toggles the search input visibility and focus.
func (ts *TaskSelector) ToggleSearch() {
	ts.showSearch = !ts.showSearch
	if ts.showSearch {
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
	return ts.showSearch
}

// SelectAll selects all currently visible (filtered) tasks.
func (ts *TaskSelector) SelectAll() {
	for _, task := range ts.filteredTasks {
		ts.selectedTasks[task.Name] = true
	}
}

// DeselectAll deselects all currently selected tasks.
func (ts *TaskSelector) DeselectAll() {
	ts.selectedTasks = make(map[string]bool)
}

// ToggleSelected toggles the selection state of the currently focused task.
func (ts *TaskSelector) ToggleSelected() {
	if len(ts.filteredTasks) == 0 {
		return
	}
	
	if ts.focusedIndex >= 0 && ts.focusedIndex < len(ts.filteredTasks) {
		task := ts.filteredTasks[ts.focusedIndex]
		ts.selectedTasks[task.Name] = !ts.selectedTasks[task.Name]
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
	if ts.showSearch {
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
	
	// Header with counter
	selectedCount := ts.GetSelectedCount()
	totalTasks := len(ts.availableTasks)
	header := fmt.Sprintf("Select Tasks:                      [%d/%d selected]", selectedCount, totalTasks)
	
	// Search box (if enabled)
	if ts.showSearch {
		searchBox := styles.TextInputStyle.Render(ts.searchInput.View())
		sections = append(sections, searchBox)
		sections = append(sections, strings.Repeat("─", separatorLineLength))
	}
	
	// Task list
	taskList := ts.renderTaskList()
	sections = append(sections, taskList)
	
	// Add spacing before help text
	sections = append(sections, "")
	
	// Help text
	helpText := ts.renderHelpText()
	sections = append(sections, helpText)
	
	// Combine all sections
	content := strings.Join(sections, "\n")
	
	// Container with header
	container := fmt.Sprintf("%s\n%s",
		styles.LightGrayStyle.Render(header),
		styles.TaskSelectorContainerStyle.Render(content),
	)
	
	return container
}

// renderTaskList renders the scrollable list of tasks.
func (ts *TaskSelector) renderTaskList() string {
	if len(ts.filteredTasks) == 0 {
		return styles.LightGrayStyle.Render("No tasks match your search.")
	}
	
	var items []string
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
	
	return strings.Join(items, "\n")
}

// renderTaskItem renders a single task item with checkbox and styling.
func (ts *TaskSelector) renderTaskItem(task models.Task, index int) string {
	// Checkbox state
	checkbox := "☐"
	if ts.selectedTasks[task.Name] {
		checkbox = "☑"
	}
	
	// Focus indicator and styling
	focusPrefix := "  "
	itemStyle := lipgloss.NewStyle()
	
	if index == ts.focusedIndex {
		focusPrefix = "▶ "
		itemStyle = styles.FocusedTaskStyle
	} else if ts.selectedTasks[task.Name] {
		itemStyle = styles.SelectedTaskStyle
	}
	
	// Truncate path if too long
	displayPath := truncatePath(task.Path, maxPathDisplayLength)
	
	// Format item text with fixed-width columns (grid-like alignment)
	// Focus(2) + Checkbox(2) + TaskName(25) + Icon(4) + Path(remaining)
	paddedTaskName := padRight(task.Name, taskNameColumnWidth)
	itemText := fmt.Sprintf("%s%s %s 📁 %s", 
		focusPrefix, checkbox, paddedTaskName, displayPath)
		
	return itemStyle.Render(itemText)
}

// renderEmptyState renders the state when no tasks are available.
func (ts *TaskSelector) renderEmptyState() string {
	content := styles.LightGrayStyle.Render("No tasks available.\nCreate some tasks first to add them to workspaces.")
	return styles.TaskSelectorContainerStyle.Render(content)
}

// renderHelpText renders context-sensitive help text.
func (ts *TaskSelector) renderHelpText() string {
	if ts.showSearch {
		return styles.LightGrayStyle.Render("esc exit search • enter confirm")
	}
	return styles.LightGrayStyle.Render("↑/↓ navigate • space toggle • /search • ctrl+a select all • tab/shift+tab navigate")
}

// truncatePath truncates a path to fit within the specified width, adding ellipsis if needed.
func truncatePath(path string, maxWidth int) string {
	if len(path) <= maxWidth {
		return path
	}
	
	if maxWidth <= 3 {
		return "..."
	}
	
	return "..." + path[len(path)-maxWidth+3:]
}

// padRight pads a string to a fixed width using spaces, truncating with ellipsis if necessary.
func padRight(text string, width int) string {
	if len(text) > width {
		return text[:width-3] + "..."
	}
	
	paddingNeeded := width - len(text)
	if paddingNeeded <= 0 {
		return text
	}
	
	return text + strings.Repeat(" ", paddingNeeded)
}