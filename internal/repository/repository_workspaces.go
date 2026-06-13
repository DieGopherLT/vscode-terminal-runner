package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/samber/lo"
)

var (
	// WorkspacesSaveFile holds the absolute path to the workspaces.json file in the user's config directory.
	WorkspacesSaveFile string

	// workspaceRepo is the singleton generic repository backing every workspace adapter below.
	workspaceRepo *JSONRepository[models.Workspace]
)

func init() {
	cfgFolder, err := os.UserConfigDir()
	if err != nil {
		panic("could not determine user config directory: " + err.Error())
	}
	WorkspacesSaveFile = filepath.Join(cfgFolder, "vscode-terminal-runner", "workspaces.json")
	workspaceRepo = NewWorkspaceRepository(func() string { return WorkspacesSaveFile })
}

// NewWorkspaceRepository builds the workspace repository: workspaces are keyed
// under "workspaces" and rejected on a case-sensitive name collision.
func NewWorkspaceRepository(getSaveFile func() string) *JSONRepository[models.Workspace] {
	return &JSONRepository[models.Workspace]{
		getSaveFile: getSaveFile,
		jsonKey:     "workspaces",
		entityLabel: "workspace",
		onAppend: func(existing []models.Workspace, workspace models.Workspace) ([]models.Workspace, error) {
			_, found := lo.Find(existing, func(candidate models.Workspace) bool {
				return candidate.Name == workspace.Name
			})
			if found {
				return nil, fmt.Errorf("workspace '%s' already exists", workspace.Name)
			}
			return append(existing, workspace), nil
		},
	}
}

// ReadWorkspaces loads all workspaces from the persistence file.
func ReadWorkspaces() ([]models.Workspace, error) {
	return workspaceRepo.ReadAll()
}

// FindWorkspaceByName retrieves a workspace by its name from the saved workspaces.
func FindWorkspaceByName(name string) (*models.Workspace, error) {
	return workspaceRepo.FindByName(name)
}

// SaveWorkspace saves a workspace to the local configuration file.
func SaveWorkspace(workspace models.Workspace) error {
	return workspaceRepo.Save(workspace)
}

// UpdateWorkspace replaces the workspace stored under originalName with updatedWorkspace.
// The replacement is a single atomic write, so a rename never leaves the file with a
// duplicate or a missing record.
func UpdateWorkspace(originalName string, updatedWorkspace models.Workspace) error {
	return workspaceRepo.Update(originalName, updatedWorkspace)
}

// DeleteWorkspace removes a workspace from the local configuration file by name.
func DeleteWorkspace(name string) error {
	return workspaceRepo.Delete(name)
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

// filterOutWorkspaceByName returns a copy of workspaces without the entry whose Name
// equals name. Pure: no I/O.
func filterOutWorkspaceByName(workspaces []models.Workspace, name string) []models.Workspace {
	return lo.Filter(workspaces, func(ws models.Workspace, _ int) bool {
		return ws.Name != name
	})
}

// replaceWorkspaceByName returns a copy of workspaces with the entry matching originalName
// replaced by updatedWorkspace. Returns an error when the name is not found. Pure: no I/O.
func replaceWorkspaceByName(workspaces []models.Workspace, originalName string, updatedWorkspace models.Workspace) ([]models.Workspace, error) {
	_, index, found := lo.FindIndexOf(workspaces, func(ws models.Workspace) bool {
		return ws.Name == originalName
	})
	if !found {
		return nil, fmt.Errorf("workspace '%s' not found", originalName)
	}

	result := make([]models.Workspace, len(workspaces))
	copy(result, workspaces)
	result[index] = updatedWorkspace
	return result, nil
}
