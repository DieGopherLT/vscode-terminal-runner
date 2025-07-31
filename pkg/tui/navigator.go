package tui

type FormNavigator struct {
	FocusIndex   int
	elementCount int
}

type NavigationKeys string

const (
	KeyUp       NavigationKeys = "up"
	KeyDown     NavigationKeys = "down"
	KeyTab      NavigationKeys = "tab"
	KeyShiftTab NavigationKeys = "shift+tab"
)

func NewNavigator(elementCount int) *FormNavigator {
	return &FormNavigator{
		FocusIndex:   0,
		elementCount: elementCount,
	}
}

func (n *FormNavigator) HandleNavigation(key string) {
	switch NavigationKeys(key) {
	case KeyUp, KeyShiftTab:
		n.FocusIndex--
	case KeyDown, KeyTab:
		n.FocusIndex++
	}

	if n.FocusIndex > n.elementCount {
		n.FocusIndex = 0
	} else if n.FocusIndex < 0 {
		n.FocusIndex = n.elementCount
	}

}
