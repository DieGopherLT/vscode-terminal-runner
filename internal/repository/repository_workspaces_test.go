package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils"
)

// redirectWorkspacesSaveFile points WorkspacesSaveFile at a path inside
// t.TempDir() that does NOT yet exist (neither the directory nor the file).
// It returns a restore function that must be deferred by the caller.
//
// This is the intended injection point for all workspace repository tests: set
// WorkspacesSaveFile before calling any public function;
// ensureWorkspacesSaveFile() will create dir + file on first use, exactly as it
// would on a fresh system.
func redirectWorkspacesSaveFile(t *testing.T) func() {
	t.Helper()
	original := WorkspacesSaveFile
	WorkspacesSaveFile = filepath.Join(t.TempDir(), "vscode-terminal-runner", "workspaces.json")
	return func() { WorkspacesSaveFile = original }
}

// -- characterization tests: lock the fresh-system bootstrap behavior --

func TestReadWorkspaces_createsFileAndReturnsEmptyListOnFreshSystem(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	workspaces, err := ReadWorkspaces()
	if err != nil {
		t.Fatalf("ReadWorkspaces on a fresh path returned an unexpected error: %v", err)
	}
	if len(workspaces) != 0 {
		t.Fatalf("expected 0 workspaces on a fresh system, got %d", len(workspaces))
	}
	if _, err := os.Stat(WorkspacesSaveFile); err != nil {
		t.Fatalf("ReadWorkspaces should have created %s, but stat failed: %v", WorkspacesSaveFile, err)
	}
}

func TestSaveWorkspace_roundTrip(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	ws := models.Workspace{Name: "my-project", Tasks: []models.Task{{Name: "build"}}}
	if err := SaveWorkspace(ws); err != nil {
		t.Fatalf("SaveWorkspace returned unexpected error: %v", err)
	}

	workspaces, err := ReadWorkspaces()
	if err != nil {
		t.Fatalf("ReadWorkspaces after SaveWorkspace returned unexpected error: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("expected 1 workspace after save, got %d", len(workspaces))
	}
	if workspaces[0].Name != ws.Name {
		t.Fatalf("expected workspace name %q, got %q", ws.Name, workspaces[0].Name)
	}
}

func TestSaveWorkspace_duplicateNameReturnsError(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	ws := models.Workspace{Name: "duplicate"}
	if err := SaveWorkspace(ws); err != nil {
		t.Fatalf("first SaveWorkspace returned unexpected error: %v", err)
	}

	err := SaveWorkspace(ws)
	if err == nil {
		t.Fatal("second SaveWorkspace with the same name should have returned an error")
	}
}

func TestDeleteWorkspace_removesExistingWorkspace(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	ws := models.Workspace{Name: "to-delete"}
	if err := SaveWorkspace(ws); err != nil {
		t.Fatalf("setup SaveWorkspace failed: %v", err)
	}

	if err := DeleteWorkspace("to-delete"); err != nil {
		t.Fatalf("DeleteWorkspace returned unexpected error: %v", err)
	}

	workspaces, err := ReadWorkspaces()
	if err != nil {
		t.Fatalf("ReadWorkspaces after DeleteWorkspace returned unexpected error: %v", err)
	}
	if len(workspaces) != 0 {
		t.Fatalf("expected 0 workspaces after deletion, got %d", len(workspaces))
	}
}

// -- additional characterization tests for uncovered functions --

func TestFindWorkspaceByName_returnsWorkspaceOnMatch(t *testing.T) {
	tests := []struct {
		name      string
		savedName string
		queryName string
		wantErr   bool
	}{
		{
			name:      "exact match returns workspace",
			savedName: "my-project",
			queryName: "my-project",
			wantErr:   false,
		},
		{
			name:      "different-case name does not match (case-sensitive)",
			savedName: "My-Project",
			queryName: "my-project",
			wantErr:   true,
		},
		{
			name:      "missing name returns error",
			savedName: "my-project",
			queryName: "nonexistent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer redirectWorkspacesSaveFile(t)()

			ws := testutils.NewWorkspace().WithName(tt.savedName).Build()
			if err := SaveWorkspace(ws); err != nil {
				t.Fatalf("setup SaveWorkspace failed: %v", err)
			}

			found, err := FindWorkspaceByName(tt.queryName)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if found == nil {
				t.Fatal("expected a workspace but got nil")
			}
		})
	}
}

func TestSaveWorkspace_multipleDistinctWorkspacesRoundTrip(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	task := testutils.NewTask().WithName("lint").Build()
	ws1 := testutils.NewWorkspace().WithName("frontend").WithTasks(task).Build()
	ws2 := testutils.NewWorkspace().WithName("backend").Build()

	if err := SaveWorkspace(ws1); err != nil {
		t.Fatalf("SaveWorkspace(frontend) returned unexpected error: %v", err)
	}
	if err := SaveWorkspace(ws2); err != nil {
		t.Fatalf("SaveWorkspace(backend) returned unexpected error: %v", err)
	}

	workspaces, err := ReadWorkspaces()
	if err != nil {
		t.Fatalf("ReadWorkspaces returned unexpected error: %v", err)
	}
	if len(workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(workspaces))
	}
}

// -- error-category: corrupt JSON input --

func TestReadWorkspaces_returnsErrorOnCorruptJSON(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	if _, err := ReadWorkspaces(); err != nil {
		t.Fatalf("setup ReadWorkspaces failed: %v", err)
	}

	if err := os.WriteFile(WorkspacesSaveFile, []byte("{not valid json"), 0666); err != nil {
		t.Fatalf("failed to write corrupt JSON: %v", err)
	}

	_, err := ReadWorkspaces()
	if err == nil {
		t.Fatal("ReadWorkspaces should return an error for corrupt JSON")
	}
}

func TestDeleteWorkspace_returnsErrorOnCorruptJSON(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	if _, err := ReadWorkspaces(); err != nil {
		t.Fatalf("setup ReadWorkspaces failed: %v", err)
	}
	if err := os.WriteFile(WorkspacesSaveFile, []byte("{not valid json"), 0666); err != nil {
		t.Fatalf("failed to write corrupt JSON: %v", err)
	}

	err := DeleteWorkspace("any")
	if err == nil {
		t.Fatal("DeleteWorkspace should return an error when workspaces.json contains corrupt JSON")
	}
}

func TestSaveWorkspace_returnsErrorOnCorruptJSON(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	if _, err := ReadWorkspaces(); err != nil {
		t.Fatalf("setup ReadWorkspaces failed: %v", err)
	}
	if err := os.WriteFile(WorkspacesSaveFile, []byte("{not valid json"), 0666); err != nil {
		t.Fatalf("failed to write corrupt JSON: %v", err)
	}

	ws := testutils.NewWorkspace().WithName("new").Build()
	err := SaveWorkspace(ws)
	if err == nil {
		t.Fatal("SaveWorkspace should return an error when workspaces.json contains corrupt JSON")
	}
}

func TestUpdateWorkspace_successfullyReplacesWorkspace(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	original := testutils.NewWorkspace().WithName("original").Build()
	if err := SaveWorkspace(original); err != nil {
		t.Fatalf("setup SaveWorkspace failed: %v", err)
	}

	updated := testutils.NewWorkspace().WithName("renamed").Build()
	if err := UpdateWorkspace("original", updated); err != nil {
		t.Fatalf("UpdateWorkspace returned unexpected error: %v", err)
	}

	workspaces, err := ReadWorkspaces()
	if err != nil {
		t.Fatalf("ReadWorkspaces after UpdateWorkspace returned unexpected error: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("expected 1 workspace after update, got %d", len(workspaces))
	}
	if workspaces[0].Name != "renamed" {
		t.Fatalf("expected updated workspace name %q, got %q", "renamed", workspaces[0].Name)
	}
}

func TestUpdateWorkspace_returnsErrorWhenNameNotFound(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	updated := testutils.NewWorkspace().WithName("new-name").Build()
	err := UpdateWorkspace("nonexistent", updated)
	if err == nil {
		t.Fatal("UpdateWorkspace should return an error for a missing workspace name")
	}
}

func TestUpdateWorkspace_returnsErrorOnCorruptJSON(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	if _, err := ReadWorkspaces(); err != nil {
		t.Fatalf("setup ReadWorkspaces failed: %v", err)
	}
	if err := os.WriteFile(WorkspacesSaveFile, []byte("{not valid json"), 0666); err != nil {
		t.Fatalf("failed to write corrupt JSON: %v", err)
	}

	err := UpdateWorkspace("any", models.Workspace{Name: "any"})
	if err == nil {
		t.Fatal("UpdateWorkspace should return an error when workspaces.json contains corrupt JSON")
	}
}

// -- ImportWorkspaces and ValidateWorkspaceBatch tests --

func TestImportWorkspaces_successOnFreshSystem(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	batch := []models.Workspace{
		testutils.NewWorkspace().WithName("dev").Build(),
		testutils.NewWorkspace().WithName("staging").Build(),
	}

	if err := ImportWorkspaces(batch); err != nil {
		t.Fatalf("ImportWorkspaces returned unexpected error: %v", err)
	}

	workspaces, err := ReadWorkspaces()
	if err != nil {
		t.Fatalf("ReadWorkspaces after ImportWorkspaces returned unexpected error: %v", err)
	}
	if len(workspaces) != 2 {
		t.Fatalf("expected 2 workspaces after import, got %d", len(workspaces))
	}
}

func TestImportWorkspaces_emptyBatchIsNoOp(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	if err := ImportWorkspaces([]models.Workspace{}); err != nil {
		t.Fatalf("ImportWorkspaces with empty batch returned unexpected error: %v", err)
	}

	workspaces, err := ReadWorkspaces()
	if err != nil {
		t.Fatalf("ReadWorkspaces returned unexpected error: %v", err)
	}
	if len(workspaces) != 0 {
		t.Fatalf("expected 0 workspaces for empty batch, got %d", len(workspaces))
	}
}

func TestImportWorkspaces_rejectsEmptyName(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	batch := []models.Workspace{{Name: ""}}
	err := ImportWorkspaces(batch)
	if err == nil {
		t.Fatal("ImportWorkspaces should return an error for a workspace with empty name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("expected 'name is required' in error, got: %v", err)
	}
}

func TestImportWorkspaces_rejectsDiskCollision(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	if err := SaveWorkspace(testutils.NewWorkspace().WithName("clash").Build()); err != nil {
		t.Fatalf("setup SaveWorkspace failed: %v", err)
	}

	batch := []models.Workspace{testutils.NewWorkspace().WithName("clash").Build()}
	err := ImportWorkspaces(batch)
	if err == nil {
		t.Fatal("ImportWorkspaces should return an error for a name that already exists on disk")
	}
	if !strings.Contains(err.Error(), "already exists on disk") {
		t.Fatalf("expected 'already exists on disk' in error, got: %v", err)
	}

	workspaces, _ := ReadWorkspaces()
	if len(workspaces) != 1 {
		t.Fatalf("expected file unchanged (1 workspace) after collision rejection, got %d", len(workspaces))
	}
}

func TestImportWorkspaces_rejectsIntraBatchCollision(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	batch := []models.Workspace{
		testutils.NewWorkspace().WithName("dup").Build(),
		testutils.NewWorkspace().WithName("dup").Build(),
	}
	err := ImportWorkspaces(batch)
	if err == nil {
		t.Fatal("ImportWorkspaces should return an error for duplicate names within the batch")
	}
	if !strings.Contains(err.Error(), "duplicate name in batch") {
		t.Fatalf("expected 'duplicate name in batch' in error, got: %v", err)
	}
}

func TestImportWorkspaces_writesNothingOnAnyError(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	if err := SaveWorkspace(testutils.NewWorkspace().WithName("pre-existing").Build()); err != nil {
		t.Fatalf("setup SaveWorkspace failed: %v", err)
	}

	batch := []models.Workspace{
		testutils.NewWorkspace().WithName("good").Build(),
		{Name: ""},
	}
	if err := ImportWorkspaces(batch); err == nil {
		t.Fatal("ImportWorkspaces should return an error when any entry is invalid")
	}

	workspaces, _ := ReadWorkspaces()
	if len(workspaces) != 1 {
		t.Fatalf("expected file unchanged (1 workspace) after partial error, got %d", len(workspaces))
	}
}

func TestValidateWorkspaceBatch_returnsNilForValidBatch(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	batch := []models.Workspace{testutils.NewWorkspace().WithName("ok").Build()}
	if err := ValidateWorkspaceBatch(batch); err != nil {
		t.Fatalf("ValidateWorkspaceBatch returned unexpected error for a valid batch: %v", err)
	}

	workspaces, _ := ReadWorkspaces()
	if len(workspaces) != 0 {
		t.Fatalf("ValidateWorkspaceBatch should not write anything, but found %d workspaces", len(workspaces))
	}
}

func TestValidateWorkspaceBatch_returnsErrorsWithoutWriting(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()

	batch := []models.Workspace{{Name: ""}}
	if err := ValidateWorkspaceBatch(batch); err == nil {
		t.Fatal("ValidateWorkspaceBatch should return an error for invalid batch")
	}

	workspaces, _ := ReadWorkspaces()
	if len(workspaces) != 0 {
		t.Fatalf("ValidateWorkspaceBatch should not write anything on error, but found %d workspaces", len(workspaces))
	}
}
