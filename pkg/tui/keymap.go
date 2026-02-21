package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds all key bindings used across task and workspace forms.
type KeyMap struct {
	Up             key.Binding
	Down           key.Binding
	Tab            key.Binding
	ShiftTab       key.Binding
	Enter          key.Binding
	Quit           key.Binding
	NextSuggestion key.Binding
	PrevSuggestion key.Binding
	Space          key.Binding
	Search         key.Binding
	SelectAll      key.Binding
	DeselectAll    key.Binding
}

// DefaultKeys is the default key map used by all TUI forms.
var DefaultKeys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next / apply suggestion"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm / apply suggestion"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("esc", "quit"),
	),
	NextSuggestion: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "next suggestion"),
	),
	PrevSuggestion: key.NewBinding(
		key.WithKeys("ctrl+b"),
		key.WithHelp("ctrl+b", "previous suggestion"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle selection"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	),
	DeselectAll: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "deselect all"),
	),
}
