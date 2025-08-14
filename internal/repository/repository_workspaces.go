// internal/repository/repository_workspaces.go
package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/samber/lo"
)

var WorkspacesSaveFile = path.Join(os.Getenv("HOME"), ".config/vsct-runner/workspaces.json")

// WorkspaceSaveFileContent represents the structure of the workspace persistence file.
type WorkspaceSaveFileContent struct {
	Workspaces []models.Workspace `json:"workspaces"`
}

// ReadWorkspaces loads all workspaces from the persistence file.
func ReadWorkspaces() ([]models.Workspace, error) {
	file, err := os.OpenFile(WorkspacesSaveFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	jsonContent, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var content WorkspaceSaveFileContent
	if len(jsonContent) > 0 {
		if err = json.Unmarshal(jsonContent, &content); err != nil {
			return nil, err
		}
	}

	return content.Workspaces, nil
}

// FindWorkspaceByName retrieves a workspace by its name from the saved workspaces.
func FindWorkspaceByName(name string) (*models.Workspace, error) {
	workspaces, err := ReadWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("failed to load workspaces: %w", err)
	}

	workspace, found := lo.Find(workspaces, func(ws models.Workspace) bool {
		return strings.EqualFold(ws.Name, name)
	})

	if !found {
		return nil, fmt.Errorf("workspace '%s' not found", name)
	}

	return &workspace, nil
}

// SaveWorkspace saves a workspace to the local configuration file.
func SaveWorkspace(workspace models.Workspace) error {
	if err := os.MkdirAll(path.Dir(WorkspacesSaveFile), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(WorkspacesSaveFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonContent, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var content WorkspaceSaveFileContent
	if len(jsonContent) > 0 {
		if err = json.Unmarshal(jsonContent, &content); err != nil {
			return err
		}
	}

	if _, found := lo.Find(content.Workspaces, func(ws models.Workspace) bool {
		return ws.Name == workspace.Name
	}); found {
		return fmt.Errorf("workspace '%s' already exists", workspace.Name)
	}

	content.Workspaces = append(content.Workspaces, workspace)

	newJsonContent, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(WorkspacesSaveFile, newJsonContent, 0666)
}

// DeleteWorkspace removes a workspace from the local configuration file by name.
func DeleteWorkspace(name string) error {
	if err := os.MkdirAll(path.Dir(WorkspacesSaveFile), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(WorkspacesSaveFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonContent, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var content WorkspaceSaveFileContent
	if len(jsonContent) > 0 {
		if err = json.Unmarshal(jsonContent, &content); err != nil {
			return err
		}
	}

	content.Workspaces = lo.Filter(content.Workspaces, func(ws models.Workspace, _ int) bool {
		return ws.Name != name
	})

	encoded, err := json.Marshal(content)
	if err != nil {
		return err
	}

	file.Truncate(0)
	file.Seek(0, 0)
	_, err = file.Write(encoded)
	return err
}
