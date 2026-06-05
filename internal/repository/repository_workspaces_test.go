package repository

import (
	"os"
	"path/filepath"
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

// -- pure-transform unit tests: no filesystem required --

func TestAppendWorkspaceIfUnique_addsWhenNameAbsent(t *testing.T) {
	initial := []models.Workspace{{Name: "alpha"}}
	ws := models.Workspace{Name: "beta"}

	result, err := appendWorkspaceIfUnique(initial, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(result))
	}
	if result[1].Name != "beta" {
		t.Fatalf("expected appended workspace name %q, got %q", "beta", result[1].Name)
	}
}

func TestAppendWorkspaceIfUnique_returnsErrorForDuplicateName(t *testing.T) {
	initial := []models.Workspace{{Name: "alpha"}}
	_, err := appendWorkspaceIfUnique(initial, models.Workspace{Name: "alpha"})
	if err == nil {
		t.Fatal("expected an error when workspace name already exists")
	}
}

func TestFilterOutWorkspaceByName_removesMatchingEntry(t *testing.T) {
	workspaces := []models.Workspace{{Name: "remove-me"}, {Name: "keep-me"}}
	result := filterOutWorkspaceByName(workspaces, "remove-me")
	if len(result) != 1 {
		t.Fatalf("expected 1 workspace after filter, got %d", len(result))
	}
	if result[0].Name != "keep-me" {
		t.Fatalf("expected remaining workspace %q, got %q", "keep-me", result[0].Name)
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
			name:      "case-insensitive match (EqualFold) returns workspace",
			savedName: "My-Project",
			queryName: "my-project",
			wantErr:   false,
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

// -- pure-transform edge cases --

func TestAppendWorkspaceIfUnique_onEmptySliceAddsWorkspace(t *testing.T) {
	ws := testutils.NewWorkspace().WithName("first").Build()
	result, err := appendWorkspaceIfUnique(nil, ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 workspace from nil base, got %d", len(result))
	}
}

func TestFilterOutWorkspaceByName_onNonMatchingNameReturnsAll(t *testing.T) {
	workspaces := []models.Workspace{{Name: "a"}, {Name: "b"}}
	result := filterOutWorkspaceByName(workspaces, "nonexistent")
	if len(result) != 2 {
		t.Fatalf("expected 2 workspaces when filter name does not match, got %d", len(result))
	}
}

func TestFilterOutWorkspaceByName_onEmptySliceReturnsEmpty(t *testing.T) {
	result := filterOutWorkspaceByName(nil, "any")
	if len(result) != 0 {
		t.Fatalf("expected 0 workspaces for nil input, got %d", len(result))
	}
}
