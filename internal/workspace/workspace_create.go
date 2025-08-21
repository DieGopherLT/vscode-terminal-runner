package workspace

import (
	"fmt"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// CreateWorkspaceCommand creates a new workspace using the TUI form.
func CreateWorkspaceCommand() error {
	model := NewWorkspaceModel()
	
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	
	if err != nil {
		return fmt.Errorf("failed to run workspace creation form: %w", err)
	}
	
	// Check if the program completed successfully
	if workspaceModel, ok := finalModel.(*WorkspaceModel); ok {
		if workspaceModel.messages.HasNonErrorMessages() {
			// Success message was already shown in the TUI
			return nil
		}
	}
	
	return nil
}

// EditWorkspaceCommand edits an existing workspace using the TUI form.
func EditWorkspaceCommand(workspaceName string) error {
	// Load the existing workspace
	workspace, err := repository.FindWorkspaceByName(workspaceName)
	if err != nil {
		styles.PrintError(fmt.Sprintf("Workspace '%s' not found: %v", workspaceName, err))
		return err
	}
	
	model := NewEditWorkspaceModel(workspace)
	
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	
	if err != nil {
		return fmt.Errorf("failed to run workspace edit form: %w", err)
	}
	
	// Check if the program completed successfully
	if workspaceModel, ok := finalModel.(*WorkspaceModel); ok {
		if workspaceModel.messages.HasNonErrorMessages() {
			// Success message was already shown in the TUI
			return nil
		}
	}
	
	return nil
}