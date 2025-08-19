package task

import (
	"fmt"
	"strconv"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/vscode"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/messages"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TaskSelectorModel manages the state for selecting a task from a list.
type TaskSelectorModel struct {
	nav       *tui.FormNavigator
	tasks     []models.Task
	operation string
	messages  *messages.MessageManager
	quitting  bool
}

// NewTaskSelectorModel creates a new task selector model.
func NewTaskSelectorModel(tasks []models.Task, operation string) *TaskSelectorModel {
	return &TaskSelectorModel{
		nav:       tui.NewNavigator(len(tasks)),
		tasks:     tasks,
		operation: operation,
		messages:  messages.NewManager(),
		quitting:  false,
	}
}

// Init initializes the task selector model.
func (m *TaskSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages received by the task selector model.
func (m *TaskSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
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
			selectedTask := m.tasks[m.nav.FocusIndex]
			return m, m.handleTaskAction(selectedTask)

		default:
			// Handle numeric selection
			if num, err := strconv.Atoi(msg.String()); err == nil && num >= 1 && num <= len(m.tasks) {
				m.nav.FocusIndex = num - 1
				selectedTask := m.tasks[m.nav.FocusIndex]
				return m, m.handleTaskAction(selectedTask)
			}
		}
	}

	return m, nil
}

// View renders the task selector view.
func (m *TaskSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	var sections []string

	// Title
	title := fmt.Sprintf("SELECT TASK TO %s", m.operation)
	sections = append(sections, styles.RenderTitle(title))

	// Task list
	for i, task := range m.tasks {
		var taskStyle lipgloss.Style
		taskText := fmt.Sprintf("%d. %s", i+1, task.Name)

		if i == m.nav.FocusIndex {
			taskStyle = styles.FocusedInputStyle
			taskText = "► " + taskText
		} else {
			taskStyle = lipgloss.NewStyle().Foreground(styles.LightGray)
			taskText = "  " + taskText
		}

		taskDescription := fmt.Sprintf("Path: %s | Commands: %v", task.Path, task.Cmds)
		taskContent := lipgloss.JoinVertical(
			lipgloss.Left,
			taskStyle.Render(taskText),
			lipgloss.NewStyle().
				Foreground(styles.LightGray).
				MarginLeft(4).
				Render(taskDescription),
		)

		sections = append(sections, taskContent)
	}

	// Render messages if any exist
	if m.messages.HasMessages() {
		sections = append(sections, m.messages.Render())
	}

	// Help text
	helpText := styles.HelpTextStyle.Render("↑/↓ navigate • 1-9 direct select • enter confirm • esc/q cancel")
	sections = append(sections, helpText)

	return styles.FormContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
}

// handleTaskAction performs the selected operation on the chosen task.
func (m *TaskSelectorModel) handleTaskAction(task models.Task) tea.Cmd {
	switch m.operation {
	case "edit":
		return tea.Sequence(
			tea.Quit,
			func() tea.Msg {
				p := tea.NewProgram(NewEditModel(&task))
				if _, err := p.Run(); err != nil {
					fmt.Printf("Error editing task: %v\n", err)
				}
				return nil
			},
		)

	case "delete":
		return tea.Sequence(
			tea.Quit,
			func() tea.Msg {
				if err := repository.DeleteTask(task.Name); err != nil {
					fmt.Printf("Error deleting task '%s': %v\n", task.Name, err)
				} else {
					fmt.Printf("Task '%s' deleted successfully.\n", task.Name)
				}
				return nil
			},
		)

	case "run":
		return tea.Sequence(
			tea.Quit,
			func() tea.Msg {
				runner, err := vscode.NewRunner()
				if err != nil {
					fmt.Printf("Failed to create runner: %v\n", err)
					return nil
				}

				fmt.Printf("Running task '%s'...\n", task.Name)
				if err := runner.RunTask(task.Name); err != nil {
					fmt.Printf("Error running task: %v\n", err)
				}
				return nil
			},
		)

	default:
		return tea.Quit
	}
}