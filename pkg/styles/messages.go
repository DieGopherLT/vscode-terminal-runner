package styles

import "github.com/charmbracelet/lipgloss"

// Message styles for different types of notifications
var (
	// Error message styles
	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(Error).
				Bold(true).
				PaddingLeft(1)

	ErrorContainerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2D1B1B")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Error).
				Padding(0, 1).
				MarginBottom(1).
				Width(70)

	// Success message styles  
	SuccessMessageStyle = lipgloss.NewStyle().
				Foreground(Success).
				Bold(true).
				PaddingLeft(1)

	SuccessContainerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1B2D1B")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Success).
				Padding(0, 1).
				MarginBottom(1).
				Width(70)

	// Warning message styles
	WarningMessageStyle = lipgloss.NewStyle().
				Foreground(Warning).
				Bold(true).
				PaddingLeft(1)

	WarningContainerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2D241B")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Warning).
				Padding(0, 1).
				MarginBottom(1).
				Width(70)

	// Info message styles
	InfoMessageStyle = lipgloss.NewStyle().
			Foreground(VSCodeBlue).
			Bold(true).
			PaddingLeft(1)

	InfoContainerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1B1F2D")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(VSCodeBlue).
			Padding(0, 1).
			MarginBottom(1).
			Width(70)
)

// Message icons for different types
const (
	ErrorIcon   = "✗"
	SuccessIcon = "✓"
	WarningIcon = "⚠"
	InfoIcon    = "ℹ"
)

// RenderErrorMessage renders a styled error message with container.
func RenderErrorMessage(content string) string {
	message := ErrorMessageStyle.Render(ErrorIcon + " " + content)
	return ErrorContainerStyle.Render(message)
}

// RenderSuccessMessage renders a styled success message with container.
func RenderSuccessMessage(content string) string {
	message := SuccessMessageStyle.Render(SuccessIcon + " " + content)
	return SuccessContainerStyle.Render(message)
}

// RenderWarningMessage renders a styled warning message with container.
func RenderWarningMessage(content string) string {
	message := WarningMessageStyle.Render(WarningIcon + " " + content)
	return WarningContainerStyle.Render(message)
}

// RenderInfoMessage renders a styled info message with container.
func RenderInfoMessage(content string) string {
	message := InfoMessageStyle.Render(InfoIcon + " " + content)
	return InfoContainerStyle.Render(message)
}