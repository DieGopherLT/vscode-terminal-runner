package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils"
	"github.com/spf13/cobra"
)

// TestCompleteWorkspaceNames_filtersByPrefixAndFormatsDescription locks the core
// completion contract: only names that start with toComplete come back, each
// carrying a "name\tcomma-separated-task-names" description (a workspace has no
// command or path of its own), always with the no-file-completion directive.
func TestCompleteWorkspaceNames_filtersByPrefixAndFormatsDescription(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()
	seedWorkspaces(t,
		testutils.NewWorkspace().WithName("backend").WithTasks(
			testutils.NewTask().WithName("build").Build(),
			testutils.NewTask().WithName("test").Build(),
		).Build(),
		testutils.NewWorkspace().WithName("frontend").WithTasks(
			testutils.NewTask().WithName("build-frontend").Build(),
		).Build(),
		testutils.NewWorkspace().WithName("empty").WithNoTasks().Build(),
	)

	cases := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{
			name:       "empty prefix returns every workspace",
			toComplete: "",
			expected: []string{
				"backend\tbuild, test",
				"frontend\tbuild-frontend",
				"empty\t",
			},
		},
		{
			name:       "prefix narrows to a single workspace",
			toComplete: "back",
			expected:   []string{"backend\tbuild, test"},
		},
		{
			name:       "workspace with no tasks yields an empty description",
			toComplete: "empty",
			expected:   []string{"empty\t"},
		},
		{
			name:       "prefix with no match returns no candidates",
			toComplete: "zzz",
			expected:   []string{},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			completions, directive := completeWorkspaceNames(nil, nil, testCase.toComplete)

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Fatalf("expected ShellCompDirectiveNoFileComp, got %v", directive)
			}
			if !reflect.DeepEqual(completions, testCase.expected) {
				t.Fatalf("completions mismatch\n got: %#v\nwant: %#v", completions, testCase.expected)
			}
		})
	}
}

// TestCompleteWorkspaceNames_skipsRepositoryWhenArgumentPresent verifies that
// once the single positional name is supplied there is nothing left to
// complete, and the repository is never read.
func TestCompleteWorkspaceNames_skipsRepositoryWhenArgumentPresent(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()
	writeCorruptWorkspacesFile(t) // would surface as an error directive if it were read

	completions, directive := completeWorkspaceNames(nil, []string{"backend"}, "")

	if completions != nil {
		t.Fatalf("expected nil completions when an argument is already present, got %#v", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

// TestCompleteWorkspaceNames_returnsErrorDirectiveOnCorruptRepository ensures a
// failed read degrades into ShellCompDirectiveError rather than panicking or
// returning garbage candidates.
func TestCompleteWorkspaceNames_returnsErrorDirectiveOnCorruptRepository(t *testing.T) {
	defer redirectWorkspacesSaveFile(t)()
	writeCorruptWorkspacesFile(t)

	completions, directive := completeWorkspaceNames(nil, nil, "")

	if completions != nil {
		t.Fatalf("expected nil completions on read failure, got %#v", completions)
	}
	if directive != cobra.ShellCompDirectiveError {
		t.Fatalf("expected ShellCompDirectiveError, got %v", directive)
	}
}

// redirectWorkspacesSaveFile points repository.WorkspacesSaveFile at a fresh
// path inside t.TempDir() and returns a restore function the caller must defer.
// It mirrors the seam the repository's own tests use, driven here from the
// workspace package through the exported global.
func redirectWorkspacesSaveFile(t *testing.T) func() {
	t.Helper()
	original := repository.WorkspacesSaveFile
	repository.WorkspacesSaveFile = filepath.Join(t.TempDir(), "vscode-terminal-runner", "workspaces.json")
	return func() { repository.WorkspacesSaveFile = original }
}

func seedWorkspaces(t *testing.T, workspaces ...models.Workspace) {
	t.Helper()
	for _, workspace := range workspaces {
		if err := repository.SaveWorkspace(workspace); err != nil {
			t.Fatalf("seedWorkspaces: failed to save %q: %v", workspace.Name, err)
		}
	}
}

func writeCorruptWorkspacesFile(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(repository.WorkspacesSaveFile), 0o755); err != nil {
		t.Fatalf("writeCorruptWorkspacesFile: mkdir: %v", err)
	}
	if err := os.WriteFile(repository.WorkspacesSaveFile, []byte("{ not valid json"), 0o644); err != nil {
		t.Fatalf("writeCorruptWorkspacesFile: write: %v", err)
	}
}
