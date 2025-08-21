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

	// Suggestion styles
	SuggestionContainerStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#1a1a1a")).
					Foreground(LightGray).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#333333")).
					Padding(0, 1).
					MarginTop(0).
					Width(58)

	SuggestionItemStyle = lipgloss.NewStyle().
				Foreground(LightGray).
				PaddingLeft(1)

	SuggestionHighlightStyle = lipgloss.NewStyle().
					Foreground(VSCodeBlue).
					Bold(true).
					PaddingLeft(1)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(VSCodeBlue).
			Bold(true).
			MarginBottom(1)

	// Task selector specific styles
	TaskSelectorContainerStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(GrayBlue).
					Padding(0, 1).
					Width(90).
					Height(9)

	SelectedTaskStyle = lipgloss.NewStyle().
				Foreground(VSCodeBlue).
				Bold(true)

	FocusedTaskStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2d2d2d")).
				Foreground(White).
				Bold(true)

	TextInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GrayBlue).
			Padding(0, 1).
			Width(86)
	
	// Light gray style for text
	LightGrayStyle = lipgloss.NewStyle().
			Foreground(LightGray)
)

// ASCII title template
const ASCIITitleTemplate = `
 ╔══════════════════════════════════════╗
 ║              	%s             ║
 ╚══════════════════════════════════════╝`

// RenderTitle renders a title with ASCII box border and proper centering.
func RenderTitle(title string) string {
	// Calculate the inner width (total width minus border characters)
	const totalWidth = 40
	const borderWidth = 2 // "║" on each side
	const innerWidth = totalWidth - borderWidth
	
	// Center the title within the inner width
	paddedTitle := centerText(title, innerWidth)
	
	asciiTitle := "╔══════════════════════════════════════╗\n" +
		"║" + paddedTitle + "║\n" +
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
