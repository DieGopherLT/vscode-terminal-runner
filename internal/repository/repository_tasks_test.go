package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils"
)

// redirectTasksSaveFile points TasksSaveFile at a path inside t.TempDir() that
// does NOT yet exist (neither the directory nor the file). It returns a restore
// function that must be deferred by the caller.
//
// This is the intended injection point for all repository tests: set
// TasksSaveFile before calling any public function; ensureTasksSaveFile() will
// create dir + file on first use, exactly as it would on a fresh system.
func redirectTasksSaveFile(t *testing.T) func() {
	t.Helper()
	original := TasksSaveFile
	TasksSaveFile = filepath.Join(t.TempDir(), "vscode-terminal-runner", "tasks.json")
	return func() { TasksSaveFile = original }
}

// -- characterization tests: lock the fresh-system bootstrap behavior --

func TestReadTasks_createsFileAndReturnsEmptyListOnFreshSystem(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks on a fresh path returned an unexpected error: %v", err)
	}
	if tasks == nil {
		tasks = []models.Task{}
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks on a fresh system, got %d", len(tasks))
	}
	if _, err := os.Stat(TasksSaveFile); err != nil {
		t.Fatalf("ReadTasks should have created %s, but stat failed: %v", TasksSaveFile, err)
	}
}

func TestSaveTask_roundTrip(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	task := models.Task{Name: "build", Path: "/app", Cmds: []string{"make"}}
	if err := SaveTask(task); err != nil {
		t.Fatalf("SaveTask returned unexpected error: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks after SaveTask returned unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task after save, got %d", len(tasks))
	}
	if tasks[0].Name != task.Name {
		t.Fatalf("expected task name %q, got %q", task.Name, tasks[0].Name)
	}
}

func TestSaveFromFile_emptyTasksJsonReturnsError(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	// Ensure the tasks.json exists but is empty (ReadTasks creates it empty).
	if _, err := ReadTasks(); err != nil {
		t.Fatalf("setup ReadTasks failed: %v", err)
	}

	// Write a valid batch file with one task.
	batchFile := filepath.Join(t.TempDir(), "batch.json")
	batch := []models.Task{{Name: "lint", Path: "/app", Cmds: []string{"golangci-lint run"}}}
	batchBytes, _ := json.Marshal(batch)
	if err := os.WriteFile(batchFile, batchBytes, 0666); err != nil {
		t.Fatalf("failed to write batch file: %v", err)
	}

	// tasks.json is empty at this point → SaveFromFile must return the sentinel error.
	err := SaveFromFile(batchFile)
	if err == nil {
		t.Fatal("SaveFromFile with an empty tasks.json should have returned an error")
	}
	if err.Error() != "Provided file is empty" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTask_returnsErrorWhenNameNotFound(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	updated := models.Task{Name: "new-name", Path: "/app", Cmds: []string{"go build"}}
	err := UpdateTask("nonexistent", updated)
	if err == nil {
		t.Fatal("UpdateTask should return an error for a missing task name")
	}
}

func TestDeleteTask_removesExistingTask(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	if err := SaveTask(models.Task{Name: "to-delete", Path: "/", Cmds: []string{"rm -rf"}}); err != nil {
		t.Fatalf("setup SaveTask failed: %v", err)
	}

	if err := DeleteTask("to-delete"); err != nil {
		t.Fatalf("DeleteTask returned unexpected error: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks after DeleteTask returned unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks after deletion, got %d", len(tasks))
	}
}

// -- pure-transform unit tests: no filesystem required --

func TestAppendTask_addsToSlice(t *testing.T) {
	initial := []models.Task{{Name: "a"}}
	result := appendTask(initial, models.Task{Name: "b"})
	if len(result) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result))
	}
	if result[1].Name != "b" {
		t.Fatalf("expected appended task name %q, got %q", "b", result[1].Name)
	}
}

func TestAppendTaskBatch_appendsAll(t *testing.T) {
	initial := []models.Task{{Name: "a"}}
	batch := []models.Task{{Name: "b"}, {Name: "c"}}
	result := appendTaskBatch(initial, batch)
	if len(result) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(result))
	}
}

func TestReplaceTaskByName_replacesCorrectEntry(t *testing.T) {
	tasks := []models.Task{{Name: "old"}, {Name: "keep"}}
	updated := models.Task{Name: "new"}

	result, err := replaceTaskByName(tasks, "old", updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Name != "new" {
		t.Fatalf("expected first task to be renamed to %q, got %q", "new", result[0].Name)
	}
	if result[1].Name != "keep" {
		t.Fatalf("expected second task to remain %q, got %q", "keep", result[1].Name)
	}
}

func TestReplaceTaskByName_returnsErrorForMissingName(t *testing.T) {
	tasks := []models.Task{{Name: "alpha"}}
	_, err := replaceTaskByName(tasks, "nonexistent", models.Task{})
	if err == nil {
		t.Fatal("expected an error for a missing task name")
	}
}

func TestFilterOutTaskByName_removesMatchingEntry(t *testing.T) {
	tasks := []models.Task{{Name: "remove-me"}, {Name: "keep-me"}}
	result := filterOutTaskByName(tasks, "remove-me")
	if len(result) != 1 {
		t.Fatalf("expected 1 task after filter, got %d", len(result))
	}
	if result[0].Name != "keep-me" {
		t.Fatalf("expected remaining task %q, got %q", "keep-me", result[0].Name)
	}
}

// -- additional characterization tests for uncovered functions --

func TestGetAllTasks_returnsAllSavedTasks(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	task1 := testutils.NewTask().WithName("first").Build()
	task2 := testutils.NewTask().WithName("second").Build()

	if err := SaveTask(task1); err != nil {
		t.Fatalf("SaveTask(first) returned unexpected error: %v", err)
	}
	if err := SaveTask(task2); err != nil {
		t.Fatalf("SaveTask(second) returned unexpected error: %v", err)
	}

	all, err := GetAllTasks()
	if err != nil {
		t.Fatalf("GetAllTasks returned unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks from GetAllTasks, got %d", len(all))
	}
}

func TestFindTaskByName_returnsTaskOnMatch(t *testing.T) {
	tests := []struct {
		name      string
		savedName string
		queryName string
		wantErr   bool
	}{
		{
			name:      "exact match returns task",
			savedName: "build",
			queryName: "build",
			wantErr:   false,
		},
		{
			name:      "case-insensitive match (EqualFold) returns task",
			savedName: "Build",
			queryName: "build",
			wantErr:   false,
		},
		{
			name:      "missing name returns error",
			savedName: "build",
			queryName: "nonexistent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer redirectTasksSaveFile(t)()

			task := testutils.NewTask().WithName(tt.savedName).Build()
			if err := SaveTask(task); err != nil {
				t.Fatalf("setup SaveTask failed: %v", err)
			}

			found, err := FindTaskByName(tt.queryName)
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
				t.Fatal("expected a task but got nil")
			}
		})
	}
}

func TestUpdateTask_successfullyReplacesTask(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	original := testutils.NewTask().WithName("original").WithCmds("make").Build()
	if err := SaveTask(original); err != nil {
		t.Fatalf("setup SaveTask failed: %v", err)
	}

	updated := testutils.NewTask().WithName("renamed").WithCmds("go build").Build()
	if err := UpdateTask("original", updated); err != nil {
		t.Fatalf("UpdateTask returned unexpected error: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks after UpdateTask returned unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task after update, got %d", len(tasks))
	}
	if tasks[0].Name != "renamed" {
		t.Fatalf("expected updated task name %q, got %q", "renamed", tasks[0].Name)
	}
}

// -- error-category: corrupt JSON input --

func TestReadTasks_returnsErrorOnCorruptJSON(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	// Bootstrap the file first so the path exists.
	if _, err := ReadTasks(); err != nil {
		t.Fatalf("setup ReadTasks failed: %v", err)
	}

	// Overwrite with malformed JSON.
	if err := os.WriteFile(TasksSaveFile, []byte("{not valid json"), 0666); err != nil {
		t.Fatalf("failed to write corrupt JSON: %v", err)
	}

	_, err := ReadTasks()
	if err == nil {
		t.Fatal("ReadTasks should return an error for corrupt JSON")
	}
}

func TestUpdateTask_returnsErrorOnCorruptJSON(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	if _, err := ReadTasks(); err != nil {
		t.Fatalf("setup ReadTasks failed: %v", err)
	}
	if err := os.WriteFile(TasksSaveFile, []byte("{not valid json"), 0666); err != nil {
		t.Fatalf("failed to write corrupt JSON: %v", err)
	}

	err := UpdateTask("any", models.Task{Name: "any"})
	if err == nil {
		t.Fatal("UpdateTask should return an error when tasks.json contains corrupt JSON")
	}
}

func TestDeleteTask_returnsErrorOnCorruptJSON(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	if _, err := ReadTasks(); err != nil {
		t.Fatalf("setup ReadTasks failed: %v", err)
	}
	if err := os.WriteFile(TasksSaveFile, []byte("{not valid json"), 0666); err != nil {
		t.Fatalf("failed to write corrupt JSON: %v", err)
	}

	err := DeleteTask("any")
	if err == nil {
		t.Fatal("DeleteTask should return an error when tasks.json contains corrupt JSON")
	}
}

func TestSaveTask_appendsWithoutDeduplication(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	// SaveTask has no duplicate guard (unlike SaveWorkspace). Pin this asymmetry.
	task := testutils.NewTask().WithName("duplicated").Build()
	if err := SaveTask(task); err != nil {
		t.Fatalf("first SaveTask returned unexpected error: %v", err)
	}
	if err := SaveTask(task); err != nil {
		t.Fatalf("second SaveTask returned unexpected error: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks returned unexpected error: %v", err)
	}
	// Characterization: duplicate is accepted. Two entries with the same name exist.
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks (SaveTask has no dedup guard), got %d", len(tasks))
	}
}

// -- SaveFromFile tests --

func TestSaveFromFile_returnsErrorForNonexistentBatchFile(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	err := SaveFromFile("/nonexistent/path/batch.json")
	if err == nil {
		t.Fatal("SaveFromFile should return an error for a nonexistent batch file")
	}
}

func TestSaveFromFile_returnsErrorForMalformedBatchFile(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batchFile := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(batchFile, []byte("{not an array}"), 0666); err != nil {
		t.Fatalf("failed to write malformed batch file: %v", err)
	}

	err := SaveFromFile(batchFile)
	if err == nil {
		t.Fatal("SaveFromFile should return an error for a malformed batch file")
	}
}

// TestSaveFromFile_succeedsWhenDestinationIsPrePopulated confirms that SaveFromFile
// works correctly when the destination tasks.json already has existing content.
// This is the only path that succeeds because SaveFromFile refuses empty destinations.
func TestSaveFromFile_succeedsWhenDestinationIsPrePopulated(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	// Pre-populate the destination with an existing task so it is not empty.
	existing := testutils.NewTask().WithName("existing").Build()
	if err := SaveTask(existing); err != nil {
		t.Fatalf("setup SaveTask failed: %v", err)
	}

	batch := []models.Task{
		testutils.NewTask().WithName("from-file-1").Build(),
		testutils.NewTask().WithName("from-file-2").Build(),
	}
	batchBytes, _ := json.Marshal(batch)
	batchFile := filepath.Join(t.TempDir(), "batch.json")
	if err := os.WriteFile(batchFile, batchBytes, 0666); err != nil {
		t.Fatalf("failed to write batch file: %v", err)
	}

	if err := SaveFromFile(batchFile); err != nil {
		t.Fatalf("SaveFromFile returned unexpected error: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks after SaveFromFile returned unexpected error: %v", err)
	}
	// 1 pre-existing + 2 from batch = 3
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks after SaveFromFile, got %d", len(tasks))
	}
}

// BUG: SaveFromFile refuses to import a batch into a fresh (empty) tasks.json.
// The function checks whether the destination file is empty and returns
// "Provided file is empty", but on a fresh install ensureTasksSaveFile() always
// creates an empty file, making it impossible to ever use SaveFromFile on a
// newly set-up system.
//
// Intended behavior: a valid batch file should import successfully even when
// tasks.json is empty (or equivalently, treats an empty file as zero tasks).
//
// Current behavior: returns error "Provided file is empty".
func TestSaveFromFile_BUG_shouldSucceedOnFreshSystem(t *testing.T) {
	t.Skip("BUG: SaveFromFile always fails on fresh install — empty tasks.json triggers 'Provided file is empty' even for a valid batch")

	defer redirectTasksSaveFile(t)()

	// Fresh system: tasks.json does not exist yet.
	batch := []models.Task{testutils.NewTask().WithName("from-file").Build()}
	batchBytes, _ := json.Marshal(batch)
	batchFile := filepath.Join(t.TempDir(), "batch.json")
	if err := os.WriteFile(batchFile, batchBytes, 0666); err != nil {
		t.Fatalf("failed to write batch file: %v", err)
	}

	if err := SaveFromFile(batchFile); err != nil {
		t.Fatalf("SaveFromFile on a fresh system should succeed, got: %v", err)
	}
}

// -- pure-transform edge cases --

func TestAppendTask_onNilSliceCreatesNewSlice(t *testing.T) {
	result := appendTask(nil, models.Task{Name: "first"})
	if len(result) != 1 {
		t.Fatalf("expected 1 task from nil base, got %d", len(result))
	}
}

func TestAppendTaskBatch_onEmptyBatchReturnsUnchanged(t *testing.T) {
	initial := []models.Task{{Name: "only"}}
	result := appendTaskBatch(initial, []models.Task{})
	if len(result) != 1 {
		t.Fatalf("expected 1 task after empty batch append, got %d", len(result))
	}
}

func TestFilterOutTaskByName_onNonMatchingNameReturnsAll(t *testing.T) {
	tasks := []models.Task{{Name: "a"}, {Name: "b"}}
	result := filterOutTaskByName(tasks, "nonexistent")
	if len(result) != 2 {
		t.Fatalf("expected 2 tasks when filter name does not match, got %d", len(result))
	}
}

func TestFilterOutTaskByName_onEmptySliceReturnsEmpty(t *testing.T) {
	result := filterOutTaskByName(nil, "any")
	if len(result) != 0 {
		t.Fatalf("expected 0 tasks for nil input, got %d", len(result))
	}
}
