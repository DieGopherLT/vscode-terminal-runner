package styles

import "github.com/charmbracelet/lipgloss"

// TableHeaderStyle is used for table headers.
var TableHeaderStyle = lipgloss.NewStyle().
	Foreground(White).
	Background(VSCodeBlue).
	Bold(true).
	Padding(0, 1)

// TableRowStyle is used for even table rows.
var TableRowStyle = lipgloss.NewStyle().
	Foreground(White).
	Background(DarkGray).
	Padding(0, 1)

// TableAltRowStyle is used for odd table rows.
var TableAltRowStyle = lipgloss.NewStyle().
	Foreground(White).
	Background(GrayBlue).
	Padding(0, 1)
