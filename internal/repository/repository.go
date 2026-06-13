package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/samber/lo"
)

// NamedEntity is the constraint for items persisted by JSONRepository: any type
// that exposes its unique name. It is defined here, where it is consumed, per
// the project's Go interface guideline.
type NamedEntity interface {
	GetName() string
}

// JSONRepository persists a slice of T as a single JSON object file whose only
// top-level key is jsonKey, e.g. {"tasks": [...]}. The per-type append rule is
// injected via onAppend (Strategy), keeping the generic core free of any
// type-specific uniqueness logic.
//
// The save-file path is resolved lazily through getSaveFile rather than captured
// at construction time, so callers (notably tests) that reassign the package
// path variable after init() are honored on the next operation.
type JSONRepository[T NamedEntity] struct {
	getSaveFile func() string
	jsonKey     string
	entityLabel string
	onAppend    func(existing []T, item T) ([]T, error)
}

// NewTaskRepository builds the task repository: tasks are keyed under "tasks"
// and appended unconditionally (duplicates are allowed by design).
func NewTaskRepository(getSaveFile func() string) *JSONRepository[models.Task] {
	return &JSONRepository[models.Task]{
		getSaveFile: getSaveFile,
		jsonKey:     "tasks",
		entityLabel: "task",
		onAppend: func(existing []models.Task, task models.Task) ([]models.Task, error) {
			return append(existing, task), nil
		},
	}
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

// ReadAll loads every persisted item, creating the backing file on first use.
// An empty file yields (nil, nil); malformed JSON returns the unmarshal error.
func (r *JSONRepository[T]) ReadAll() ([]T, error) {
	r.ensure()

	data, err := os.ReadFile(r.getSaveFile())
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var content map[string][]T
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, err
	}

	return content[r.jsonKey], nil
}

// FindByName retrieves the item whose name matches name, case-insensitively.
func (r *JSONRepository[T]) FindByName(name string) (*T, error) {
	items, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", r.jsonKey, err)
	}

	item, found := lo.Find(items, func(item T) bool {
		return strings.EqualFold(item.GetName(), name)
	})
	if !found {
		return nil, fmt.Errorf("%s '%s' not found", r.entityLabel, name)
	}

	return &item, nil
}

// Save appends item according to the injected onAppend strategy and persists
// the result. The strategy decides whether a duplicate name is allowed.
func (r *JSONRepository[T]) Save(item T) error {
	items, err := r.ReadAll()
	if err != nil {
		return err
	}

	updated, err := r.onAppend(items, item)
	if err != nil {
		return err
	}

	return r.WriteAll(updated)
}

// Update replaces the item stored under originalName with updated, matching the
// original name case-sensitively. It returns an error when no such item exists.
func (r *JSONRepository[T]) Update(originalName string, updated T) error {
	items, err := r.ReadAll()
	if err != nil {
		return err
	}

	_, index, found := lo.FindIndexOf(items, func(item T) bool {
		return item.GetName() == originalName
	})
	if !found {
		return fmt.Errorf("%s '%s' not found", r.entityLabel, originalName)
	}

	// items is freshly read and owned by this call, so it is safe to mutate in place.
	items[index] = updated
	return r.WriteAll(items)
}

// Delete removes the item whose name equals name, matching case-sensitively.
// Deleting a non-existent name succeeds silently, preserving prior behavior.
func (r *JSONRepository[T]) Delete(name string) error {
	items, err := r.ReadAll()
	if err != nil {
		return err
	}

	remaining := lo.Filter(items, func(item T, _ int) bool {
		return item.GetName() != name
	})
	return r.WriteAll(remaining)
}

// WriteAll replaces the whole file with items, serialized as {jsonKey: items}.
func (r *JSONRepository[T]) WriteAll(items []T) error {
	r.ensure()

	content := map[string][]T{r.jsonKey: items}
	encoded, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return os.WriteFile(r.getSaveFile(), encoded, 0666)
}

// ensure creates the config directory and an empty save file if either is
// missing, so the first read or write on a fresh system does not fail.
func (r *JSONRepository[T]) ensure() {
	saveFile := r.getSaveFile()
	if _, err := os.Stat(saveFile); !os.IsNotExist(err) {
		return
	}
	if err := os.MkdirAll(filepath.Dir(saveFile), 0755); err != nil {
		return
	}
	if _, err := os.Create(saveFile); err != nil {
		return
	}
}
