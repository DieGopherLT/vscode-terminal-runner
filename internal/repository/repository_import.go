package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/samber/lo"
)

// OpenSource returns a reader for the given --from-json value: os.Stdin when source == "-",
// otherwise the opened file. The returned ReadCloser is always safe to Close.
func OpenSource(source string) (io.ReadCloser, error) {
	if source == "-" {
		return io.NopCloser(os.Stdin), nil
	}
	f, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("OpenSource: %w", err)
	}
	return f, nil
}

// ReadJSONSource opens source via OpenSource, reads all bytes, and unmarshals them into T.
// It consolidates the open-read-unmarshal sequence shared by all --from-json command handlers.
func ReadJSONSource[T any](source string) (T, error) {
	var zero T

	reader, err := OpenSource(source)
	if err != nil {
		return zero, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return zero, fmt.Errorf("ReadJSONSource: %w", err)
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, fmt.Errorf("ReadJSONSource: invalid JSON: %w", err)
	}

	return result, nil
}

// ValidateTaskBatch checks the batch for all problems (VAL-001, VAL-003, VAL-004, VAL-005)
// without modifying any file. Returns nil when all entries are valid.
func ValidateTaskBatch(tasks []models.Task) error {
	existing, err := ReadTasks()
	if err != nil {
		return fmt.Errorf("ValidateTaskBatch: read existing tasks: %w", err)
	}
	return errors.Join(collectTaskBatchErrors(tasks, existing)...)
}

// ValidateWorkspaceBatch checks the batch (VAL-002, VAL-004, VAL-005) without modifying any file.
func ValidateWorkspaceBatch(workspaces []models.Workspace) error {
	existing, err := ReadWorkspaces()
	if err != nil {
		return fmt.Errorf("ValidateWorkspaceBatch: read existing workspaces: %w", err)
	}
	return errors.Join(collectWorkspaceBatchErrors(workspaces, existing)...)
}

// ImportTasks validates the whole batch (VAL-001, VAL-003, VAL-004, VAL-005) and writes
// all entries only if every entry is valid. On failure it returns an aggregate error and
// writes nothing.
func ImportTasks(tasks []models.Task) error {
	existing, err := ReadTasks()
	if err != nil {
		return fmt.Errorf("ImportTasks: read existing tasks: %w", err)
	}

	if errs := collectTaskBatchErrors(tasks, existing); len(errs) > 0 {
		return errors.Join(errs...)
	}

	content := TaskSaveFileContent{Tasks: append(existing, tasks...)}
	encoded, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("ImportTasks: marshal: %w", err)
	}

	ensureTasksSaveFile()
	return os.WriteFile(TasksSaveFile, encoded, 0666)
}

// ImportWorkspaces mirrors ImportTasks for workspaces (VAL-002, VAL-004, VAL-005).
func ImportWorkspaces(workspaces []models.Workspace) error {
	existing, err := ReadWorkspaces()
	if err != nil {
		return fmt.Errorf("ImportWorkspaces: read existing workspaces: %w", err)
	}

	if errs := collectWorkspaceBatchErrors(workspaces, existing); len(errs) > 0 {
		return errors.Join(errs...)
	}

	content := WorkspaceSaveFileContent{Workspaces: append(existing, workspaces...)}
	encoded, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("ImportWorkspaces: marshal: %w", err)
	}

	ensureWorkspacesSaveFile()
	return os.WriteFile(WorkspacesSaveFile, encoded, 0666)
}

// collectTaskBatchErrors validates each task against existing disk state and intra-batch
// duplicates. Returns every problem found; the caller decides whether to abort.
func collectTaskBatchErrors(tasks []models.Task, existing []models.Task) []error {
	existingNames := make(map[string]bool, len(existing))
	for _, t := range existing {
		existingNames[t.Name] = true
	}

	var errs []error
	seenInBatch := make(map[string]bool, len(tasks))

	for i, task := range tasks {
		errs = append(errs, validateSingleTask(i, task, existingNames, seenInBatch)...)
		if task.Name != "" {
			seenInBatch[task.Name] = true
		}
	}

	return errs
}

// collectWorkspaceBatchErrors validates each workspace against existing disk state and
// intra-batch duplicates. Returns every problem found.
func collectWorkspaceBatchErrors(workspaces []models.Workspace, existing []models.Workspace) []error {
	existingNames := make(map[string]bool, len(existing))
	for _, ws := range existing {
		existingNames[ws.Name] = true
	}

	var errs []error
	seenInBatch := make(map[string]bool, len(workspaces))

	for i, ws := range workspaces {
		errs = append(errs, validateSingleWorkspace(i, ws, existingNames, seenInBatch)...)
		if ws.Name != "" {
			seenInBatch[ws.Name] = true
		}
	}

	return errs
}

// validateSingleTask collects all validation errors for one task entry.
func validateSingleTask(idx int, task models.Task, existingNames map[string]bool, seenInBatch map[string]bool) []error {
	var errs []error

	label := fmt.Sprintf("entry %d", idx)
	if task.Name != "" {
		label = fmt.Sprintf("task %q", task.Name)
	}

	if task.Name == "" {
		errs = append(errs, fmt.Errorf("entry %d: name is required", idx))
	} else {
		if existingNames[task.Name] {
			errs = append(errs, fmt.Errorf("task %q: already exists on disk", task.Name))
		}
		if seenInBatch[task.Name] {
			errs = append(errs, fmt.Errorf("task %q: duplicate name in batch", task.Name))
		}
	}

	if len(task.Cmds) == 0 {
		errs = append(errs, fmt.Errorf("%s: cmds must not be empty", label))
	}

	if task.Icon != "" {
		validIcon := lo.ContainsBy(styles.VSCodeIcons, func(i styles.VSCodeIcon) bool {
			return i.Name == task.Icon
		})
		if !validIcon {
			errs = append(errs, fmt.Errorf("%s: icon %q is not a valid VSCode icon", label, task.Icon))
		}
	}

	if task.IconColor != "" {
		validColor := lo.ContainsBy(styles.VSCodeANSIColors, func(c styles.VSCodeANSIColor) bool {
			return c.Name == task.IconColor
		})
		if !validColor {
			errs = append(errs, fmt.Errorf("%s: iconColor %q is not a valid VSCode ANSI color", label, task.IconColor))
		}
	}

	return errs
}

// validateSingleWorkspace collects all validation errors for one workspace entry.
func validateSingleWorkspace(idx int, ws models.Workspace, existingNames map[string]bool, seenInBatch map[string]bool) []error {
	var errs []error

	if ws.Name == "" {
		return append(errs, fmt.Errorf("entry %d: workspace name is required", idx))
	}

	if existingNames[ws.Name] {
		errs = append(errs, fmt.Errorf("workspace %q: already exists on disk", ws.Name))
	}
	if seenInBatch[ws.Name] {
		errs = append(errs, fmt.Errorf("workspace %q: duplicate name in batch", ws.Name))
	}

	return errs
}
