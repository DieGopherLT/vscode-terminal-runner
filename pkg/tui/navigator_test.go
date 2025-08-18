package tui

import (
	"testing"
)

func TestNewNavigator(t *testing.T) {
	tests := []struct {
		name               string
		elementCount       int
		expectedFocusIndex int
	}{
		{
			name:               "creates navigator with single element",
			elementCount:       1,
			expectedFocusIndex: 0,
		},
		{
			name:               "creates navigator with multiple elements",
			elementCount:       5,
			expectedFocusIndex: 0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Act
			navigator := NewNavigator(testCase.elementCount)

			// Assert
			if navigator.FocusIndex != testCase.expectedFocusIndex {
				t.Errorf("expected FocusIndex to be %d, got %d", testCase.expectedFocusIndex, navigator.FocusIndex)
			}
			if navigator.elementCount != testCase.elementCount {
				t.Errorf("expected elementCount to be %d, got %d", testCase.elementCount, navigator.elementCount)
			}
		})
	}
}

func TestFormNavigator_HandleNavigation(t *testing.T) {
	tests := []struct {
		name          string
		elementCount  int
		initialIndex  int
		navigationKey string
		expectedIndex int
	}{
		{
			name:          "navigates down from first element",
			elementCount:  3,
			initialIndex:  0,
			navigationKey: "down",
			expectedIndex: 1,
		},
		{
			name:          "wraps to first when navigating down from last element",
			elementCount:  3,
			initialIndex:  2,
			navigationKey: "down",
			expectedIndex: 3,
		},
		{
			name:          "navigates up from last element",
			elementCount:  3,
			initialIndex:  2,
			navigationKey: "up",
			expectedIndex: 1,
		},
		{
			name:          "wraps to last when navigating up from first element",
			elementCount:  3,
			initialIndex:  0,
			navigationKey: "up",
			expectedIndex: 3,
		},
		{
			name:          "tab moves to next element",
			elementCount:  3,
			initialIndex:  0,
			navigationKey: "tab",
			expectedIndex: 1,
		},
		{
			name:          "shift+tab moves to previous element",
			elementCount:  3,
			initialIndex:  1,
			navigationKey: "shift+tab",
			expectedIndex: 0,
		},
		{
			name:          "shift+tab wraps to last from first element",
			elementCount:  3,
			initialIndex:  0,
			navigationKey: "shift+tab",
			expectedIndex: 3,
		},
		{
			name:          "ignores unknown navigation key",
			elementCount:  3,
			initialIndex:  1,
			navigationKey: "unknown",
			expectedIndex: 1,
		},
		{
			name:          "handles single element navigation",
			elementCount:  1,
			initialIndex:  0,
			navigationKey: "down",
			expectedIndex: 1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			navigator := NewNavigator(testCase.elementCount)
			navigator.FocusIndex = testCase.initialIndex

			// Act
			navigator.HandleNavigation(testCase.navigationKey)

			// Assert
			if navigator.FocusIndex != testCase.expectedIndex {
				t.Errorf("expected FocusIndex to be %d, got %d", testCase.expectedIndex, navigator.FocusIndex)
			}
		})
	}
}

func TestFormNavigator_GetFocusIndex(t *testing.T) {
	// Arrange
	navigator := NewNavigator(5)
	navigator.FocusIndex = 3

	// Act
	focusIndex := navigator.GetFocusIndex()

	// Assert
	if focusIndex != 3 {
		t.Errorf("expected GetFocusIndex to return 3, got %d", focusIndex)
	}
}

func TestFormNavigator_SetFocusIndex(t *testing.T) {
	tests := []struct {
		name           string
		elementCount   int
		indexToSet     int
		expectedResult bool
		expectedIndex  int
	}{
		{
			name:           "sets valid index within bounds",
			elementCount:   5,
			indexToSet:     3,
			expectedResult: true,
			expectedIndex:  3,
		},
		{
			name:           "rejects negative index",
			elementCount:   5,
			indexToSet:     -1,
			expectedResult: false,
			expectedIndex:  0,
		},
		{
			name:           "rejects index equal to element count",
			elementCount:   5,
			indexToSet:     5,
			expectedResult: false,
			expectedIndex:  0,
		},
		{
			name:           "rejects index greater than element count",
			elementCount:   5,
			indexToSet:     10,
			expectedResult: false,
			expectedIndex:  0,
		},
		{
			name:           "sets first index",
			elementCount:   5,
			indexToSet:     0,
			expectedResult: true,
			expectedIndex:  0,
		},
		{
			name:           "sets last valid index",
			elementCount:   5,
			indexToSet:     4,
			expectedResult: true,
			expectedIndex:  4,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			navigator := NewNavigator(testCase.elementCount)

			// Act
			result := navigator.SetFocusIndex(testCase.indexToSet)

			// Assert
			if result != testCase.expectedResult {
				t.Errorf("expected SetFocusIndex to return %v, got %v", testCase.expectedResult, result)
			}
			if navigator.FocusIndex != testCase.expectedIndex {
				t.Errorf("expected FocusIndex to be %d, got %d", testCase.expectedIndex, navigator.FocusIndex)
			}
		})
	}
}

func TestFormNavigator_GetElementCount(t *testing.T) {
	tests := []struct {
		name         string
		elementCount int
	}{
		{"returns count for single element", 1},
		{"returns count for multiple elements", 5},
		{"returns count for zero elements", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			navigator := NewNavigator(tt.elementCount)

			// Act
			count := navigator.GetElementCount()

			// Assert
			if count != tt.elementCount {
				t.Errorf("expected GetElementCount to return %d, got %d", tt.elementCount, count)
			}
		})
	}
}

// BenchmarkFormNavigator_HandleNavigation measures navigation performance.
func BenchmarkFormNavigator_HandleNavigation(b *testing.B) {
	navigator := NewNavigator(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		navigator.HandleNavigation("down")
	}
}
