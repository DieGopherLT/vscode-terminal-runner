package tui

import (
	"testing"

	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui/suggestions"
)

func TestSuggestionManager_FilterSuggestions(t *testing.T) {
	tests := []struct {
		name            string
		allSuggestions  []string
		filterText      string
		expectedResults []string
	}{
		{
			name:            "filters with exact match",
			allSuggestions:  []string{"task1", "task2", "test1"},
			filterText:      "task",
			expectedResults: []string{"task1", "task2"},
		},
		{
			name:            "returns all when filter is empty",
			allSuggestions:  []string{"task1", "task2", "test1"},
			filterText:      "",
			expectedResults: []string{"task1", "task2", "test1"},
		},
		{
			name:            "returns empty when no matches",
			allSuggestions:  []string{"task1", "task2", "test1"},
			filterText:      "xyz",
			expectedResults: []string{},
		},
		{
			name:            "handles case insensitive filtering",
			allSuggestions:  []string{"Task1", "TASK2", "test1"},
			filterText:      "task",
			expectedResults: []string{"Task1", "TASK2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager := suggestions.NewManager(tt.allSuggestions, 10, suggestions.StartsWithFilter)

			// Act
			manager.UpdateFilter(tt.filterText)
			results := manager.GetVisible()

			// Assert
			if len(results) != len(tt.expectedResults) {
				t.Errorf("expected %d results, got %d", len(tt.expectedResults), len(results))
				return
			}

			for i, expected := range tt.expectedResults {
				if i >= len(results) || results[i] != expected {
					t.Errorf("expected result[%d] to be '%s', got '%s'", i, expected, results[i])
				}
			}
		})
	}
}

func TestSuggestionManager_Navigation(t *testing.T) {
	tests := []struct {
		name              string
		suggestions       []string
		initialIndex      int
		navigationKey     string
		expectedIndex     int
		expectedSelection string
	}{
		{
			name:              "navigates down in suggestions",
			suggestions:       []string{"option1", "option2", "option3"},
			initialIndex:      0,
			navigationKey:     "down",
			expectedIndex:     1,
			expectedSelection: "option2",
		},
		{
			name:              "wraps to top when navigating past last suggestion",
			suggestions:       []string{"option1", "option2", "option3"},
			initialIndex:      2,
			navigationKey:     "down",
			expectedIndex:     0,
			expectedSelection: "option1",
		},
		{
			name:              "navigates up in suggestions",
			suggestions:       []string{"option1", "option2", "option3"},
			initialIndex:      2,
			navigationKey:     "up",
			expectedIndex:     1,
			expectedSelection: "option2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager := suggestions.NewManager(tt.suggestions, 10, suggestions.StartsWithFilter)

			// Set initial index by navigating
			for i := 0; i < tt.initialIndex; i++ {
				manager.Next()
			}

			// Act
			if tt.navigationKey == "down" {
				manager.Next()
			} else if tt.navigationKey == "up" {
				manager.Previous()
			}

			// Assert
			selection := manager.GetSelected()
			if selection != tt.expectedSelection {
				t.Errorf("expected selection to be '%s', got '%s'", tt.expectedSelection, selection)
			}
		})
	}
}

func TestSuggestionManager_Selection(t *testing.T) {
	tests := []struct {
		name                string
		suggestions         []string
		selectedIndex       int
		expectedSelection   string
		expectedShouldClose bool
	}{
		{
			name:                "selects suggestion at index",
			suggestions:         []string{"option1", "option2", "option3"},
			selectedIndex:       1,
			expectedSelection:   "option2",
			expectedShouldClose: true,
		},
		{
			name:                "handles empty suggestions",
			suggestions:         []string{},
			selectedIndex:       0,
			expectedSelection:   "",
			expectedShouldClose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager := suggestions.NewManager(tt.suggestions, 10, suggestions.StartsWithFilter)

			// Navigate to selected index
			for i := 0; i < tt.selectedIndex; i++ {
				manager.Next()
			}

			// Act
			selection := manager.GetSelected()

			// Assert
			if selection != tt.expectedSelection {
				t.Errorf("expected selection to be '%s', got '%s'", tt.expectedSelection, selection)
			}
		})
	}
}

// TestSuggestionManager_ShouldShow tests suggestion visibility logic
func TestSuggestionManager_ShouldShow(t *testing.T) {
	tests := []struct {
		name                 string
		inputText            string
		availableSuggestions []string
		showOnEmpty          bool
		expectedVisible      bool
	}{
		{
			name:                 "shows suggestions when input has matches",
			inputText:            "tas",
			availableSuggestions: []string{"task1", "task2"},
			showOnEmpty:          false,
			expectedVisible:      true,
		},
		{
			name:                 "hides when input is empty and showOnEmpty is false",
			inputText:            "",
			availableSuggestions: []string{"task1", "task2"},
			showOnEmpty:          false,
			expectedVisible:      false,
		},
		{
			name:                 "shows when input is empty and showOnEmpty is true",
			inputText:            "",
			availableSuggestions: []string{"task1", "task2"},
			showOnEmpty:          true,
			expectedVisible:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager := suggestions.NewManagerWithOptions(
				tt.availableSuggestions,
				10,
				suggestions.StartsWithFilter,
				tt.showOnEmpty,
			)

			// Act
			manager.UpdateFilter(tt.inputText)
			shouldShow := manager.ShouldShow(tt.inputText)

			// Assert
			if shouldShow != tt.expectedVisible {
				t.Errorf("expected ShouldShow to return %v, got %v", tt.expectedVisible, shouldShow)
			}
		})
	}
}

// TestSuggestionManager_Filters tests different filter functions
func TestSuggestionManager_Filters(t *testing.T) {
	tests := []struct {
		name            string
		suggestions     []string
		filterFunc      func(string, string) bool
		input           string
		expectedResults []string
	}{
		{
			name:            "StartsWithFilter matches prefixes",
			suggestions:     []string{"task-one", "task-two", "test-one"},
			filterFunc:      suggestions.StartsWithFilter,
			input:           "task",
			expectedResults: []string{"task-one", "task-two"},
		},
		{
			name:            "ContainsFilter matches anywhere",
			suggestions:     []string{"my-task-one", "task-two", "test-task"},
			filterFunc:      suggestions.ContainsFilter,
			input:           "task",
			expectedResults: []string{"my-task-one", "task-two", "test-task"},
		},
		{
			name:            "WordBoundaryFilter matches word boundaries",
			suggestions:     []string{"terminal-bash", "bash-script", "mybash"},
			filterFunc:      suggestions.WordBoundaryFilter,
			input:           "bash",
			expectedResults: []string{"terminal-bash", "bash-script"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager := suggestions.NewManager(tt.suggestions, 10, tt.filterFunc)

			// Act
			manager.UpdateFilter(tt.input)
			results := manager.GetVisible()

			// Assert
			if len(results) != len(tt.expectedResults) {
				t.Errorf("expected %d results, got %d", len(tt.expectedResults), len(results))
				return
			}

			for i, expected := range tt.expectedResults {
				if i >= len(results) || results[i] != expected {
					t.Errorf("expected result[%d] to be '%s', got '%s'", i, expected, results[i])
				}
			}
		})
	}
}

// TestSuggestionManager_MaxVisible tests the maxVisible limitation
func TestSuggestionManager_MaxVisible(t *testing.T) {
	// Arrange
	allSuggestions := []string{"task1", "task2", "task3", "task4", "task5"}
	maxVisible := 3
	manager := suggestions.NewManager(allSuggestions, maxVisible, suggestions.StartsWithFilter)

	// Act
	manager.UpdateFilter("task") // Should match all 5
	visible := manager.GetVisible()

	// Assert
	if len(visible) != maxVisible {
		t.Errorf("expected %d visible suggestions, got %d", maxVisible, len(visible))
	}
}

// TestSuggestionManager_Reset tests the reset functionality
func TestSuggestionManager_Reset(t *testing.T) {
	// Arrange
	manager := suggestions.NewManager([]string{"option1", "option2", "option3"}, 10, suggestions.StartsWithFilter)
	manager.Next() // Move to index 1
	manager.Next() // Move to index 2

	// Act
	manager.Reset()
	selection := manager.GetSelected()

	// Assert
	if selection != "option1" {
		t.Errorf("expected selection to be 'option1' after reset, got '%s'", selection)
	}
}
