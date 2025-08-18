package tui

// FormNavigator manages focus and navigation between elements in a TUI form.
type FormNavigator struct {
	FocusIndex   int // Index of the currently focused element
	elementCount int // Total number of navigable elements
}

// NavigationKeys represents the supported navigation keys.
type NavigationKeys string

const (
	KeyUp       NavigationKeys = "up"        // Navigate up
	KeyDown     NavigationKeys = "down"      // Navigate down
	KeyTab      NavigationKeys = "tab"       // Navigate to the next element
	KeyShiftTab NavigationKeys = "shift+tab" // Navigate to the previous element
)

// NewNavigator creates a new FormNavigator for a form with the given number of elements.
func NewNavigator(elementCount int) *FormNavigator {
	return &FormNavigator{
		FocusIndex:   0,
		elementCount: elementCount,
	}
}

// HandleNavigation updates the focus index according to the received navigation key.
func (n *FormNavigator) HandleNavigation(key string) {
	switch NavigationKeys(key) {
	case KeyUp, KeyShiftTab:
		n.FocusIndex--
	case KeyDown, KeyTab:
		n.FocusIndex++
	}

	// Original logic was correct for your application
	if n.FocusIndex > n.elementCount {
		n.FocusIndex = 0
	} else if n.FocusIndex < 0 {
		n.FocusIndex = n.elementCount
	}
}

// GetFocusIndex returns the current focus index.
func (n *FormNavigator) GetFocusIndex() int {
	return n.FocusIndex
}

// SetFocusIndex sets the focus index to a specific value if within bounds.
func (n *FormNavigator) SetFocusIndex(index int) bool {
	if index < 0 || index >= n.elementCount {
		return false
	}
	n.FocusIndex = index
	return true
}

// GetElementCount returns the total number of navigable elements.
func (n *FormNavigator) GetElementCount() int {
	return n.elementCount
}
