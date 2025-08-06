package suggestions

import (
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

// FilterFunc defines how suggestions are filtered based on input
type FilterFunc func(suggestion, input string) bool

// Manager handles autocomplete suggestions with navigation and rendering
type Manager struct {
	allSuggestions      []string   // All available suggestions
	filteredSuggestions []string   // Current filtered suggestions
	selectedIndex       int        // Currently selected suggestion index
	maxVisible          int        // Maximum suggestions to show
	filterFunc          FilterFunc // Function to filter suggestions
	lastInput           string     // Last input used for filtering
	showOnEmpty         bool       // Whether to show suggestions when input is empty
}

// NewManager creates a new suggestion manager (shows suggestions on empty input by default)
func NewManager(suggestions []string, maxVisible int, filterFunc FilterFunc) *Manager {
	return NewManagerWithOptions(suggestions, maxVisible, filterFunc, true)
}

// NewManagerWithOptions creates a new suggestion manager with custom options
func NewManagerWithOptions(suggestions []string, maxVisible int, filterFunc FilterFunc, showOnEmpty bool) *Manager {
	if filterFunc == nil {
		filterFunc = StartsWithFilter
	}

	return &Manager{
		allSuggestions:      suggestions,
		filteredSuggestions: suggestions,
		selectedIndex:       0,
		maxVisible:          maxVisible,
		filterFunc:          filterFunc,
		showOnEmpty:         showOnEmpty,
	}
}

// SetSuggestions updates all available suggestions
func (sm *Manager) SetSuggestions(suggestions []string) {
	sm.allSuggestions = suggestions
	sm.filteredSuggestions = suggestions
	sm.selectedIndex = 0
	sm.lastInput = ""
}

// UpdateFilter filters suggestions based on input and resets selection only if input changed
func (sm *Manager) UpdateFilter(input string) {
	// Only update if input actually changed
	if input == sm.lastInput {
		return
	}

	sm.lastInput = input

	if input == "" {
		sm.filteredSuggestions = sm.allSuggestions
	} else {
		sm.filteredSuggestions = make([]string, 0)
		sm.filteredSuggestions = lo.Filter(sm.allSuggestions, func(suggestion string, _ int) bool {
			return sm.filterFunc(suggestion, input)
		})
	}

	// Reset selection only when filter actually changes
	sm.selectedIndex = 0
}

// Next moves to the next suggestion (circular)
func (sm *Manager) Next() {
	visible := sm.GetVisible()
	if len(visible) > 0 {
		sm.selectedIndex = (sm.selectedIndex + 1) % len(visible)
	}
}

// Previous moves to the previous suggestion (circular)
func (sm *Manager) Previous() {
	visible := sm.GetVisible()
	if len(visible) > 0 {
		sm.selectedIndex = (sm.selectedIndex - 1 + len(visible)) % len(visible)
	}
}

// Reset resets the selected index to 0
func (sm *Manager) Reset() {
	sm.selectedIndex = 0
}

// GetSelected returns the currently selected suggestion
func (sm *Manager) GetSelected() string {
	visible := sm.GetVisible()
	if len(visible) > 0 && sm.selectedIndex < len(visible) {
		return visible[sm.selectedIndex]
	}
	return ""
}

// GetVisible returns the suggestions that should be displayed (limited by maxVisible)
func (sm *Manager) GetVisible() []string {
	if len(sm.filteredSuggestions) <= sm.maxVisible {
		return sm.filteredSuggestions
	}
	return sm.filteredSuggestions[:sm.maxVisible]
}

// ShouldShow determines if suggestions should be shown based on current state and input
func (sm *Manager) ShouldShow(input string) bool {
	// Don't show if configured not to show on empty input
	if !sm.showOnEmpty && input == "" {
		return false
	}

	visible := sm.GetVisible()

	// Don't show if no suggestions
	if len(visible) == 0 {
		return false
	}

	// Don't show if only one suggestion and it's an exact match
	if len(visible) == 1 && visible[0] == input {
		return false
	}

	return true
}

// ApplySelected applies the selected suggestion to the given textinput
func (sm *Manager) ApplySelected(input *textinput.Model) {
	selected := sm.GetSelected()
	if selected == "" {
		return
	}
	input.SetValue(selected)
	input.SetCursor(len(selected))
	sm.Reset()
}

// Render returns the rendered suggestions as a string
func (sm *Manager) Render() string {
	visible := sm.GetVisible()
	if len(visible) == 0 {
		return ""
	}

	suggestionLines := lo.Map(visible, func(suggestion string, i int) string {
		if i == sm.selectedIndex {
			return styles.SuggestionHighlightStyle.Render("• " + suggestion)
		}
		return styles.SuggestionItemStyle.Render("• " + suggestion)
	})

	return styles.SuggestionContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, suggestionLines...),
	)
}
