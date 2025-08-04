package styles

import "github.com/charmbracelet/lipgloss"

// Form and input styles
var (
	// Input field styles
	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(VSCodeBlue).
				Bold(true)

	BlurredInputStyle = lipgloss.NewStyle().
				Foreground(LightGray)

	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(LightGray)

	// Label styles
	FieldLabelStyle = lipgloss.NewStyle().
			Foreground(GrayBlue).
			Bold(true)

	// Container styles
	FieldContainerStyle = lipgloss.NewStyle().
				MarginBottom(0)

	FormContainerStyle = lipgloss.NewStyle().
				Padding(0, 1).
				MarginTop(1)

	// Help and navigation text
	HelpTextStyle = lipgloss.NewStyle().
			Foreground(LightGray).
			MarginTop(1)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(VSCodeBlue).
			Bold(true).
			MarginBottom(1)
)

// ASCII title template
const ASCIITitleTemplate = `
 ╔══════════════════════════════════════╗
 ║              	%s             ║
 ╚══════════════════════════════════════╝`

// Function to render title with custom text
func RenderTitle(title string) string {
	// Pad title to center it (adjust padding as needed)
	paddedTitle := centerText(title, 34) // 34 chars to fit in the box
	asciiTitle := "╔══════════════════════════════════════╗\n" +
		"║              " + paddedTitle + "             ║\n" +
		"╚══════════════════════════════════════╝"
	return TitleStyle.Render(asciiTitle)
}

// Helper function to center text within given width
func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}

	totalPadding := width - len(text)
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	result := ""
	for i := 0; i < leftPadding; i++ {
		result += " "
	}
	result += text
	for i := 0; i < rightPadding; i++ {
		result += " "
	}

	return result
}
