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

var (
	WorkspacesSaveFile string
)

func init() {
	cfgFolder, err := os.UserConfigDir()
	if err != nil {
		panic("could not determine user config directory: " + err.Error())
	}
	WorkspacesSaveFile = path.Join(cfgFolder, "vscode-terminal-runner", "workspaces.json")
}

// ensureWorkspacesSaveFile creates the directory and the workspaces file if they
// do not exist yet. It is called lazily by every public function so that
// importing this package no longer touches the filesystem at init time.
func ensureWorkspacesSaveFile() {
	if _, err := os.Stat(WorkspacesSaveFile); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(WorkspacesSaveFile), 0755); err != nil {
			return
		}
		if _, err := os.Create(WorkspacesSaveFile); err != nil {
			return
		}
	}
}

// WorkspaceSaveFileContent represents the structure of the workspace persistence file.
type WorkspaceSaveFileContent struct {
	Workspaces []models.Workspace `json:"workspaces"`
}

// ReadWorkspaces loads all workspaces from the persistence file.
func ReadWorkspaces() ([]models.Workspace, error) {
	ensureWorkspacesSaveFile()
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
	ensureWorkspacesSaveFile()
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

	updated, err := appendWorkspaceIfUnique(content.Workspaces, workspace)
	if err != nil {
		return err
	}
	content.Workspaces = updated

	newJsonContent, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(WorkspacesSaveFile, newJsonContent, 0666)
}

// appendWorkspaceIfUnique returns a new slice with workspace appended, or an error when
// a workspace with the same name already exists. Pure: no I/O.
func appendWorkspaceIfUnique(workspaces []models.Workspace, workspace models.Workspace) ([]models.Workspace, error) {
	if _, found := lo.Find(workspaces, func(ws models.Workspace) bool {
		return ws.Name == workspace.Name
	}); found {
		return nil, fmt.Errorf("workspace '%s' already exists", workspace.Name)
	}
	return append(workspaces, workspace), nil
}

// DeleteWorkspace removes a workspace from the local configuration file by name.
func DeleteWorkspace(name string) error {
	ensureWorkspacesSaveFile()
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

	content.Workspaces = filterOutWorkspaceByName(content.Workspaces, name)

	encoded, err := json.Marshal(content)
	if err != nil {
		return err
	}

	file.Truncate(0)
	file.Seek(0, 0)
	_, err = file.Write(encoded)
	return err
}

// filterOutWorkspaceByName returns a copy of workspaces without the entry whose Name
// equals name. Pure: no I/O.
func filterOutWorkspaceByName(workspaces []models.Workspace, name string) []models.Workspace {
	return lo.Filter(workspaces, func(ws models.Workspace, _ int) bool {
		return ws.Name != name
	})
}
