package task

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

// TestCompleteTaskNames_filtersByPrefixAndFormatsDescription locks the core
// completion contract: only names that start with toComplete come back, each
// carrying a "name\tjoined-commands" description, always with the
// no-file-completion directive.
func TestCompleteTaskNames_filtersByPrefixAndFormatsDescription(t *testing.T) {
	defer redirectTasksSaveFile(t)()
	seedTasks(t,
		testutils.NewTask().WithName("build").WithCmds("go build ./...").Build(),
		testutils.NewTask().WithName("build-frontend").WithCmds("npm run build").Build(),
		testutils.NewTask().WithName("test").WithCmds("go test ./...").Build(),
		testutils.NewTask().WithName("deploy").WithCmds("build", "push").Build(),
	)

	cases := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{
			name:       "empty prefix returns every task",
			toComplete: "",
			expected: []string{
				"build\tgo build ./...",
				"build-frontend\tnpm run build",
				"test\tgo test ./...",
				"deploy\tbuild && push",
			},
		},
		{
			name:       "prefix narrows to the matching subset",
			toComplete: "build",
			expected: []string{
				"build\tgo build ./...",
				"build-frontend\tnpm run build",
			},
		},
		{
			name:       "prefix matching a single task",
			toComplete: "te",
			expected:   []string{"test\tgo test ./..."},
		},
		{
			name:       "prefix with no match returns no candidates",
			toComplete: "zzz",
			expected:   []string{},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			completions, directive := completeTaskNames(nil, nil, testCase.toComplete)

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Fatalf("expected ShellCompDirectiveNoFileComp, got %v", directive)
			}
			if !reflect.DeepEqual(completions, testCase.expected) {
				t.Fatalf("completions mismatch\n got: %#v\nwant: %#v", completions, testCase.expected)
			}
		})
	}
}

// TestCompleteTaskNames_skipsRepositoryWhenArgumentPresent verifies that once
// the single positional name is supplied there is nothing left to complete, and
// the repository is never read.
func TestCompleteTaskNames_skipsRepositoryWhenArgumentPresent(t *testing.T) {
	defer redirectTasksSaveFile(t)()
	writeCorruptTasksFile(t) // would surface as an error directive if it were read

	completions, directive := completeTaskNames(nil, []string{"build"}, "")

	if completions != nil {
		t.Fatalf("expected nil completions when an argument is already present, got %#v", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

// TestCompleteTaskNames_returnsErrorDirectiveOnCorruptRepository ensures a
// failed read degrades into ShellCompDirectiveError rather than panicking or
// returning garbage candidates.
func TestCompleteTaskNames_returnsErrorDirectiveOnCorruptRepository(t *testing.T) {
	defer redirectTasksSaveFile(t)()
	writeCorruptTasksFile(t)

	completions, directive := completeTaskNames(nil, nil, "")

	if completions != nil {
		t.Fatalf("expected nil completions on read failure, got %#v", completions)
	}
	if directive != cobra.ShellCompDirectiveError {
		t.Fatalf("expected ShellCompDirectiveError, got %v", directive)
	}
}

// redirectTasksSaveFile points repository.TasksSaveFile at a fresh path inside
// t.TempDir() and returns a restore function the caller must defer. It mirrors
// the seam the repository's own tests use, driven here from the task package
// through the exported global.
func redirectTasksSaveFile(t *testing.T) func() {
	t.Helper()
	original := repository.TasksSaveFile
	repository.TasksSaveFile = filepath.Join(t.TempDir(), "vscode-terminal-runner", "tasks.json")
	return func() { repository.TasksSaveFile = original }
}

func seedTasks(t *testing.T, tasks ...models.Task) {
	t.Helper()
	for _, task := range tasks {
		if err := repository.SaveTask(task); err != nil {
			t.Fatalf("seedTasks: failed to save %q: %v", task.Name, err)
		}
	}
}

func writeCorruptTasksFile(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(repository.TasksSaveFile), 0o755); err != nil {
		t.Fatalf("writeCorruptTasksFile: mkdir: %v", err)
	}
	if err := os.WriteFile(repository.TasksSaveFile, []byte("{ not valid json"), 0o644); err != nil {
		t.Fatalf("writeCorruptTasksFile: write: %v", err)
	}
}
