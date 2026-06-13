package repository

import (
	"os"
	"path/filepath"
	"strings"
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

// -- additional characterization tests for uncovered functions --

func TestReadTasks_returnsAllSavedTasks(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	task1 := testutils.NewTask().WithName("first").Build()
	task2 := testutils.NewTask().WithName("second").Build()

	if err := SaveTask(task1); err != nil {
		t.Fatalf("SaveTask(first) returned unexpected error: %v", err)
	}
	if err := SaveTask(task2); err != nil {
		t.Fatalf("SaveTask(second) returned unexpected error: %v", err)
	}

	all, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks returned unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks from ReadTasks, got %d", len(all))
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

func TestSaveTask_rejectsDuplicateName(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	// The task name is the unique handle every read path keys on, so SaveTask
	// rejects a second task with the same name (mirrors SaveWorkspace).
	task := testutils.NewTask().WithName("duplicated").Build()
	if err := SaveTask(task); err != nil {
		t.Fatalf("first SaveTask returned unexpected error: %v", err)
	}

	err := SaveTask(task)
	if err == nil {
		t.Fatal("second SaveTask with the same name should have returned an error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected 'already exists' in error, got: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks returned unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected the duplicate to be rejected (1 task), got %d", len(tasks))
	}
}

// -- ImportTasks and ValidateTaskBatch tests --

func TestImportTasks_successOnFreshSystem(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{
		testutils.NewTask().WithName("alpha").WithCmds("go run .").Build(),
		testutils.NewTask().WithName("beta").WithCmds("make").Build(),
	}

	if err := ImportTasks(batch); err != nil {
		t.Fatalf("ImportTasks returned unexpected error: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks after ImportTasks returned unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks after import, got %d", len(tasks))
	}
}

func TestImportTasks_appendsToExistingTasks(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	if err := SaveTask(testutils.NewTask().WithName("existing").Build()); err != nil {
		t.Fatalf("setup SaveTask failed: %v", err)
	}

	batch := []models.Task{testutils.NewTask().WithName("new-one").Build()}
	if err := ImportTasks(batch); err != nil {
		t.Fatalf("ImportTasks returned unexpected error: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks returned unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks after import into populated file, got %d", len(tasks))
	}
}

func TestImportTasks_emptyBatchIsNoOp(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	if err := ImportTasks([]models.Task{}); err != nil {
		t.Fatalf("ImportTasks with empty batch returned unexpected error: %v", err)
	}

	tasks, err := ReadTasks()
	if err != nil {
		t.Fatalf("ReadTasks returned unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks for empty batch, got %d", len(tasks))
	}
}

func TestImportTasks_rejectsEmptyName(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{{Name: "", Cmds: []string{"echo hi"}}}
	err := ImportTasks(batch)
	if err == nil {
		t.Fatal("ImportTasks should return an error for a task with empty name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("expected 'name is required' in error, got: %v", err)
	}
}

func TestImportTasks_rejectsEmptyCmds(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{{Name: "no-cmds", Cmds: []string{}}}
	err := ImportTasks(batch)
	if err == nil {
		t.Fatal("ImportTasks should return an error for a task with empty cmds")
	}
	if !strings.Contains(err.Error(), "cmds must not be empty") {
		t.Fatalf("expected 'cmds must not be empty' in error, got: %v", err)
	}
}

func TestImportTasks_rejectsDiskCollision(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	if err := SaveTask(testutils.NewTask().WithName("clash").Build()); err != nil {
		t.Fatalf("setup SaveTask failed: %v", err)
	}

	batch := []models.Task{testutils.NewTask().WithName("clash").Build()}
	err := ImportTasks(batch)
	if err == nil {
		t.Fatal("ImportTasks should return an error for a name that already exists on disk")
	}
	if !strings.Contains(err.Error(), "already exists on disk") {
		t.Fatalf("expected 'already exists on disk' in error, got: %v", err)
	}

	tasks, _ := ReadTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected file unchanged (1 task) after collision rejection, got %d", len(tasks))
	}
}

func TestImportTasks_rejectsIntraBatchCollision(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{
		testutils.NewTask().WithName("dup").Build(),
		testutils.NewTask().WithName("dup").Build(),
	}
	err := ImportTasks(batch)
	if err == nil {
		t.Fatal("ImportTasks should return an error for duplicate names within the batch")
	}
	if !strings.Contains(err.Error(), "duplicate name in batch") {
		t.Fatalf("expected 'duplicate name in batch' in error, got: %v", err)
	}
}

func TestImportTasks_reportsAllErrorsInOneBatch(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{
		{Name: "", Cmds: []string{"ok"}},
		{Name: "valid-name", Cmds: []string{}},
	}
	err := ImportTasks(batch)
	if err == nil {
		t.Fatal("ImportTasks should return errors for both invalid entries")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("expected 'name is required' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "cmds must not be empty") {
		t.Fatalf("expected 'cmds must not be empty' in error, got: %v", err)
	}
}

func TestImportTasks_rejectsInvalidIcon(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{
		testutils.NewTask().WithName("bad-icon").WithIcon("not-a-real-icon").Build(),
	}
	err := ImportTasks(batch)
	if err == nil {
		t.Fatal("ImportTasks should return an error for an invalid icon")
	}
	if !strings.Contains(err.Error(), "not a valid VSCode icon") {
		t.Fatalf("expected 'not a valid VSCode icon' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "not-a-real-icon") {
		t.Fatalf("expected icon name in error, got: %v", err)
	}
}

func TestImportTasks_rejectsInvalidIconColor(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{
		testutils.NewTask().WithName("bad-color").WithIconColor("not-a-real-color").Build(),
	}
	err := ImportTasks(batch)
	if err == nil {
		t.Fatal("ImportTasks should return an error for an invalid iconColor")
	}
	if !strings.Contains(err.Error(), "not a valid VSCode ANSI color") {
		t.Fatalf("expected 'not a valid VSCode ANSI color' in error, got: %v", err)
	}
}

func TestImportTasks_acceptsValidIconAndColor(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{
		testutils.NewTask().WithName("styled").WithIcon("account").WithIconColor("terminal.ansiBlue").Build(),
	}
	if err := ImportTasks(batch); err != nil {
		t.Fatalf("ImportTasks should accept a valid icon and iconColor, got: %v", err)
	}
}

func TestImportTasks_writesNothingOnAnyError(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	if err := SaveTask(testutils.NewTask().WithName("pre-existing").Build()); err != nil {
		t.Fatalf("setup SaveTask failed: %v", err)
	}

	// One valid entry + one invalid entry — no write should occur.
	batch := []models.Task{
		testutils.NewTask().WithName("good").Build(),
		{Name: "bad", Cmds: []string{}},
	}
	if err := ImportTasks(batch); err == nil {
		t.Fatal("ImportTasks should return an error when any entry is invalid")
	}

	tasks, _ := ReadTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected file unchanged (1 task) after partial error, got %d tasks", len(tasks))
	}
}

func TestValidateTaskBatch_returnsNilForValidBatch(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{testutils.NewTask().WithName("ok").Build()}
	if err := ValidateTaskBatch(batch); err != nil {
		t.Fatalf("ValidateTaskBatch returned unexpected error for a valid batch: %v", err)
	}

	tasks, _ := ReadTasks()
	if len(tasks) != 0 {
		t.Fatalf("ValidateTaskBatch should not write anything, but found %d tasks", len(tasks))
	}
}

func TestValidateTaskBatch_returnsErrorsWithoutWriting(t *testing.T) {
	defer redirectTasksSaveFile(t)()

	batch := []models.Task{{Name: "bad", Cmds: []string{}}}
	if err := ValidateTaskBatch(batch); err == nil {
		t.Fatal("ValidateTaskBatch should return an error for invalid batch")
	}

	tasks, _ := ReadTasks()
	if len(tasks) != 0 {
		t.Fatalf("ValidateTaskBatch should not write anything on error, but found %d tasks", len(tasks))
	}
}
