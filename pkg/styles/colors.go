package styles

import "github.com/charmbracelet/lipgloss"

// Color palette inspired by VSCode and Go
var (
	// Primary colors
	VSCodeBlue = lipgloss.Color("#007ACC") // VSCode primary blue
	LightBlue  = lipgloss.Color("#40A9FF") // Light blue accent
	GrayBlue   = lipgloss.Color("#6A9FDB") // Muted blue for labels
	
	// Neutral colors
	DarkGray  = lipgloss.Color("#3C3C3C") // VSCode dark gray
	LightGray = lipgloss.Color("#858585") // Light gray text
	White     = lipgloss.Color("#FFFFFF") // Pure white
	
	// Semantic colors
	Success = lipgloss.Color("#4CAF50") // Green for success states
	Warning = lipgloss.Color("#FF9800") // Orange for warnings
	Error   = lipgloss.Color("#F44336") // Red for errors
)