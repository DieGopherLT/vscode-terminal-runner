package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Runner output styles
var (
	// Header styles
	RunnerHeaderStyle = lipgloss.NewStyle().
				Foreground(VSCodeBlue).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(VSCodeBlue).
				BorderBottom(true).
				Padding(0, 1).
				MarginBottom(1)

	// Info and detail styles
	RunnerInfoStyle = lipgloss.NewStyle().
			Foreground(LightGray).
			PaddingLeft(2)

	RunnerTaskNameStyle = lipgloss.NewStyle().
				Foreground(White).
				Bold(true)

	// Status styles
	RunnerSuccessStyle = lipgloss.NewStyle().
				Foreground(Success).
				Bold(true)

	RunnerErrorStyle = lipgloss.NewStyle().
				Foreground(Error).
				Bold(true)

	RunnerWarningStyle = lipgloss.NewStyle().
				Foreground(Warning).
				Bold(true)

	RunnerProgressStyle = lipgloss.NewStyle().
				Foreground(LightBlue).
				Italic(true)

	// Icon styles
	RunnerIconStyle = lipgloss.NewStyle().
			Foreground(VSCodeBlue)
)

// PrintSuccess prints a success message with icon
func PrintSuccess(message string) {
	fmt.Println(RunnerSuccessStyle.Render(fmt.Sprintf("%s %s", SuccessIcon, message)))
}

// PrintError prints an error message with icon
func PrintError(message string) {
	fmt.Println(RunnerErrorStyle.Render(fmt.Sprintf("%s %s", ErrorIcon, message)))
}

// PrintWarning prints a warning message with icon
func PrintWarning(message string) {
	fmt.Println(RunnerWarningStyle.Render(fmt.Sprintf("%s %s", WarningIcon, message)))
}

// PrintInfo prints an info message with icon
func PrintInfo(message string) {
	fmt.Println(RunnerInfoStyle.Render(fmt.Sprintf("%s %s", InfoIcon, message)))
}

// PrintProgress prints a progress message with icon
func PrintProgress(message string) {
	fmt.Println(RunnerProgressStyle.Render(fmt.Sprintf("%s %s", ProgressIcon, message)))
}

// RenderTaskStatus renders a task with its status
func RenderTaskStatus(name string, icon string, success bool) string {
	statusIcon := SuccessIcon
	statusStyle := RunnerSuccessStyle
	if !success {
		statusIcon = ErrorIcon
		statusStyle = RunnerErrorStyle
	}

	return fmt.Sprintf("%s %s %s %s",
		statusStyle.Render(statusIcon),
		RunnerIconStyle.Render(icon),
		RunnerTaskNameStyle.Render(name),
		statusStyle.Render(""))
}
