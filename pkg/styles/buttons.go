package styles

import "github.com/charmbracelet/lipgloss"

// Button styles for consistent UI
var (
	// Primary button (focused state)
	FocusedButton = lipgloss.NewStyle().
			Foreground(White).
			Background(VSCodeBlue).
			Padding(0, 2)

	// Secondary button (blurred state)
	BlurredButton = lipgloss.NewStyle().
			Foreground(LightGray).
			Border(lipgloss.NormalBorder()).
			BorderForeground(LightGray).
			Padding(0, 2)

	// Danger button for destructive actions
	DangerButton = lipgloss.NewStyle().
			Foreground(White).
			Background(Error).
			Padding(0, 2)

	// Success button for positive actions
	SuccessButton = lipgloss.NewStyle().
			Foreground(White).
			Background(Success).
			Padding(0, 2)
)

// Renders a focused button with text
func RenderFocusedButton(text string) string {
	return FocusedButton.Render(text)
}

// Renders a not focused button with text
func RenderBlurredButton(text string) string {
	return BlurredButton.Render(text)
}

func RenderDangerButton(text string) string {
	return DangerButton.Render(text)
}

func RenderSuccessButton(text string) string {
	return SuccessButton.Render(text)
}
