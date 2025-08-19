package task

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/messages"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	createOption = 0
	listOption   = 1
	editOption   = 2
	deleteOption = 3
	runOption    = 4
	exitOption   = 5
)

// MenuModel manages the state and logic of the main task menu TUI.
type MenuModel struct {
	nav       *tui.FormNavigator
	messages  *messages.MessageManager
	menuItems []MenuOption
	quitting  bool
}

// clearScreen clears the terminal screen like the 'clear' command.
func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// MenuOption represents a selectable option in the menu.
type MenuOption struct {
	title       string
	description string
}

// NewMenuModel initializes and returns the TUI model for the main task menu.
func NewMenuModel() tea.Model {
	menuOptions := []MenuOption{
		{
			title:       "Create Task",
			description: "Create a new task with interactive form",
		},
		{
			title:       "List Tasks",
			description: "Display all configured tasks",
		},
		{
			title:       "Edit Task",
			description: "Edit an existing task",
		},
		{
			title:       "Delete Task",
			description: "Delete an existing task",
		},
		{
			title:       "Run Task",
			description: "Execute a task in VSCode terminal",
		},
		{
			title:       "Exit",
			description: "Exit the task manager",
		},
	}

	return &MenuModel{
		nav:       tui.NewNavigator(len(menuOptions) - 1),
		messages:  messages.NewManager(),
		menuItems: menuOptions,
		quitting:  false,
	}
}

// Init initializes the menu model.
func (m *MenuModel) Init() tea.Cmd {
	return tea.ClearScreen
}

// Update handles messages received by the menu model and updates the state.
func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.QuitMsg:
		m.quitting = true
		return m, tea.Quit
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.quitting = true
			return m, tea.Quit

		case string(tui.KeyUp), string(tui.KeyDown):
			key := msg.String()
			m.nav.HandleNavigation(key)
			return m, nil

		case "enter":
			return m, m.handleMenuAction(m.nav.FocusIndex)

		case "1", "2", "3", "4", "5", "6":
			// Direct number selection
			if num, err := strconv.Atoi(msg.String()); err == nil && num >= 1 && num <= len(m.menuItems) {
				m.nav.FocusIndex = num - 1
				return m, m.handleMenuAction(m.nav.FocusIndex)
			}
		}
	}

	return m, nil
}

// View renders the menu view.
func (m *MenuModel) View() string {
	if m.quitting {
		return ""
	}

	var sections []string

	// Title
	sections = append(sections, styles.RenderTitle("TASK MANAGER"))

	// Menu options
	for i, option := range m.menuItems {
		var optionStyle lipgloss.Style
		optionText := fmt.Sprintf("%d. %s", i+1, option.title)

		if i == m.nav.FocusIndex {
			optionStyle = styles.FocusedInputStyle
			optionText = "► " + optionText
		} else {
			optionStyle = lipgloss.NewStyle().Foreground(styles.LightGray)
			optionText = "  " + optionText
		}

		optionContent := lipgloss.JoinVertical(
			lipgloss.Left,
			optionStyle.Render(optionText),
			lipgloss.NewStyle().
				Foreground(styles.LightGray).
				MarginLeft(4).
				Render(option.description),
		)

		sections = append(sections, optionContent)
	}

	// Render messages if any exist
	if m.messages.HasMessages() {
		sections = append(sections, m.messages.Render())
	}

	// Help text
	helpText := styles.HelpTextStyle.Render("↑/↓ navigate • 1-6 direct select • enter confirm • esc/q quit")
	sections = append(sections, helpText)

	return styles.FormContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
}

// handleMenuAction handles the selected menu action.
func (m *MenuModel) handleMenuAction(index int) tea.Cmd {
	switch index {
	case createOption:
		return func() tea.Msg {
			clearScreen()
			p := tea.NewProgram(NewModel())
			_, err := p.Run()
			if err != nil {
				fmt.Printf("Error creating task: %v\n", err)
			}
			return tea.Quit
		}

	case listOption:
		return func() tea.Msg {
			clearScreen()
			if err := listAllTasks(); err != nil {
				fmt.Printf("Error listing tasks: %v\n", err)
			}
			return tea.Quit
		}

	case editOption:
		return m.showTaskSelector("edit")

	case deleteOption:
		return m.showTaskSelector("delete")

	case runOption:
		return m.showTaskSelector("run")

	case exitOption:
		return tea.Quit

	default:
		return nil
	}
}

// showTaskSelector displays a task selector for edit/delete/run operations.
func (m *MenuModel) showTaskSelector(operation string) tea.Cmd {
	return func() tea.Msg {
		clearScreen()
		
		// Get all tasks
		tasks, err := repository.ReadTasks()
		if err != nil {
			fmt.Printf("Error loading tasks: %v\n", err)
			return tea.Quit
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found. Create a task first.")
			return tea.Quit
		}

		// Create task selector
		p := tea.NewProgram(NewTaskSelectorModel(tasks, operation))
		_, err = p.Run()
		if err != nil {
			fmt.Printf("Error with task selector: %v\n", err)
		}
		return tea.Quit
	}
}
