---
title: Generic JSONRepository[T] and Set[T] for vstr
version: 1.1
date_created: 2026-06-13
last_updated: 2026-06-13
owner: Diego López torres
tags: [architecture, design, refactor, generics, go]
---

> **Changelog**
> - **v1.1 (2026-06-13)**: Post-refactor decision — task names are now unique
>   (duplicates rejected on create and rename), making Task symmetric with
>   Workspace. This intentionally diverges from the v1.0 behavior-preserving
>   baseline. Affected: REQ-009 (split into REQ-009a/REQ-009b), new REQ-015,
>   PAT-001, AC-004 (inverted), new AC-017/AC-018, §4.3 factory example.
> - **v1.0 (2026-06-13)**: Initial behavior-preserving refactor (generic
>   `JSONRepository[T]` + `Set[T]`). Delivered and verified.

# Introduction

This specification defines two related refactors for the `vstr` CLI, both delivered on a single
dedicated branch:

1. **`JSONRepository[T]`** — a generic, JSON-file-backed repository that collapses the ~208 lines of
   near-identical CRUD logic currently duplicated between `repository_tasks.go` and
   `repository_workspaces.go` into one parametrized implementation.
2. **`Set[T comparable]`** — a small generic set type that replaces the ad-hoc `map[string]bool`
   "set" usages, primarily in the workspace task-selector component.

The refactor is **behavior-preserving**. No public function signature in `internal/repository`
changes, no call site changes, and every existing test must stay green without modification. The
existing test suites are the safety net for this refactor.

## 1. Purpose & Scope

### Purpose

- Eliminate duplicated persistence logic by introducing a single generic repository (GoF: a
  parametrized **Repository**, configured via an injected **Strategy** for the append step).
- Replace untyped `map[string]bool` sets with a self-documenting `Set[T]` type.

### Scope

**In scope:**

- New file `internal/repository/repository.go` containing `JSONRepository[T]` and the `NamedEntity`
  constraint.
- New file `pkg/collections/set.go` containing `Set[T comparable]`.
- Adding a `GetName() string` method to `models.Task` and `models.Workspace`.
- Rewriting the bodies of the existing public repository functions to delegate to the generic
  repository (they become thin adapters; their signatures are unchanged).
- Migrating `internal/workspace/components/task_selector.go` to use `Set[string]`.

**Out of scope:**

- Any change to a public function signature in `internal/repository`.
- Any change to the 24 call sites listed in section 4.4.
- Batch-validation logic (`collectTaskBatchErrors`, `collectWorkspaceBatchErrors`,
  `validateSingleTask`, `validateSingleWorkspace`) — these stay as-is. Applying `Set[T]` to their
  internal `existingNames` / `seenInBatch` maps is an OPTIONAL secondary application (see REQ-014),
  not a requirement.
- The on-disk JSON file format. It must remain byte-compatible: `{"tasks": [...]}` and
  `{"workspaces": [...]}`.

### Intended audience

A Go engineer (or a cold AI agent) implementing the refactor with no prior knowledge of the
conversation that produced this spec.

## 2. Definitions

- **vstr**: the CLI binary built from this module (`github.com/DieGopherLT/vscode-terminal-runner`).
- **Adapter (here)**: an existing public function whose body is rewritten to delegate to the generic
  repository while keeping its original signature, so callers do not change.
- **Singleton repo**: a package-level `*JSONRepository[T]` value initialized in `init()`.
- **NamedEntity**: a constraint interface requiring a `GetName() string` method.
- **onAppend strategy**: a per-type callback that decides how a new item is appended (with or without
  a uniqueness check), injected at construction time.
- **Case-sensitive match**: comparison using `==` / `!=` on the `Name` field.
- **Case-insensitive match**: comparison using `strings.EqualFold` on the `Name` field.

## 3. Requirements, Constraints & Guidelines

### Requirements — JSONRepository[T]

- **REQ-001**: A generic type `JSONRepository[T NamedEntity]` SHALL be defined in
  `internal/repository/repository.go`.
- **REQ-002**: `NamedEntity` SHALL be defined as `interface { GetName() string }` in the same file
  (the interface is defined where it is consumed, per the project's Go interface guideline).
- **REQ-003**: `models.Task` and `models.Workspace` SHALL each implement `GetName() string` returning
  their `Name` field. These methods go in `internal/models` next to the respective struct.
- **REQ-004**: `JSONRepository[T]` SHALL expose the following methods, each preserving the exact
  observable behavior of the function it replaces (see section 4.1 for signatures):
  `ReadAll`, `FindByName`, `Save`, `Update`, `Delete`, `WriteAll`.
- **REQ-005**: Two package-level singletons, `taskRepo *JSONRepository[models.Task]` and
  `workspaceRepo *JSONRepository[models.Workspace]`, SHALL be constructed in `init()` **after** the
  corresponding `TasksSaveFile` / `WorkspacesSaveFile` path variables are assigned (see CON-002).
- **REQ-006**: The existing public functions SHALL become thin adapters delegating to the singletons.
  The functions and their unchanged signatures are:
  `ReadTasks`, `GetAllTasks`, `FindTaskByName`, `SaveTask`, `UpdateTask`, `DeleteTask`,
  `ReadWorkspaces`, `FindWorkspaceByName`, `SaveWorkspace`, `UpdateWorkspace`, `DeleteWorkspace`.
- **REQ-007**: `ImportTasks` and `ImportWorkspaces` (in `repository_import.go`) SHALL validate the
  batch exactly as today and then persist via `taskRepo.WriteAll` / `workspaceRepo.WriteAll` instead
  of the inline `json.Marshal` + `os.WriteFile` sequence.

### Behavior-preservation requirements (the subtle asymmetries)

These differences are observable and covered by existing tests. They MUST be preserved, not unified.

- **REQ-008** (Find is case-insensitive): `FindByName` SHALL match using `strings.EqualFold` on the
  name. Source of truth: `repository_tasks.go:79-81`, `repository_workspaces.go:78-80`.
- **REQ-009a** (Save uniqueness — symmetric as of v1.1): Both `SaveTask` and `SaveWorkspace` SHALL
  reject a name that already exists using a **case-sensitive** (`==`) comparison, returning
  `fmt.Errorf("task '%s' already exists", name)` and `fmt.Errorf("workspace '%s' already exists", name)`
  respectively. This is each repository's injected `onAppend` strategy (see PAT-001).
  > **v1.0 note (superseded)**: originally `SaveTask` allowed duplicates (no check) while only
  > `SaveWorkspace` rejected them. v1.1 removed that asymmetry because the task name is the unique
  > handle every read path (`FindTaskByName`, run, shell completion, the workspace task selector)
  > keys on; allowing duplicate names left `Find`/`Update` resolving only the first match while
  > `Delete` removed all matches — an incoherent half-state, not a feature.
- **REQ-009b** (Form-level collision guard parity): The duplicate check at the repository layer is a
  case-sensitive backstop. The user-facing uniqueness guarantee (case-insensitive, covering rename)
  is enforced in the TUI submit path — see REQ-015.
- **REQ-010** (Update match is case-sensitive): `Update` SHALL locate the record to replace using a
  **case-sensitive** (`==`) comparison on the original name, and return
  `fmt.Errorf("<entity> '%s' not found", originalName)` when absent. Source of truth:
  `repository_tasks.go:171-188`, `repository_workspaces.go:228-240`.
- **REQ-011** (Delete match is case-sensitive): `Delete` SHALL remove the record whose name equals
  the argument using a **case-sensitive** (`!=`) filter. Deleting a non-existent name SHALL succeed
  silently (current behavior — see the documented debt in `[[task-delete-silent-success-debt]]`; this
  refactor MUST NOT change it). Source of truth: `repository_tasks.go:230-234`,
  `repository_workspaces.go:220-224`.
- **REQ-012** (FindByName error messages): `FindTaskByName` SHALL keep returning
  `fmt.Errorf("task '%s' not found", name)` and `FindWorkspaceByName` →
  `fmt.Errorf("workspace '%s' not found", name)`, each wrapping read errors with
  `fmt.Errorf("failed to load <entity>: %w", err)`.

### Requirements — Set[T]

- **REQ-013**: A generic type `Set[T comparable]` SHALL be defined in `pkg/collections/set.go`
  (package `collections`) exposing: `NewSet[T]()`, `Add`, `Remove`, `Contains`, `Toggle`, `Len`,
  `Clear`. Internally it SHALL use `map[T]struct{}`.
- **REQ-014** (primary application): `TaskSelector.selectedTasks` in
  `internal/workspace/components/task_selector.go:28` SHALL change from `map[string]bool` to
  `*collections.Set[string]`, and every usage site SHALL be migrated to the `Set` API while
  preserving identical observable behavior (see section 4.5 for the per-site mapping). Applying
  `Set[T]` inside `repository_import.go` is OPTIONAL and only permitted if it does not alter the
  `validateSingleTask` / `validateSingleWorkspace` signatures or behavior.

### Requirements — Task name uniqueness (v1.1)

- **REQ-015** (Task form collision guard): `submitTaskCmd` (`internal/task/task_form.go`) SHALL reject
  a name collision before persisting, mirroring `submitWorkspaceCmd`
  (`internal/workspace/workspace_form.go:315-344`): compute
  `isClaimingNewName := !isEditMode || task.Name != originalName`; when true and
  `FindByName(task.Name)` returns no error (the name exists), return
  `taskSaveResultMsg{err: fmt.Errorf("task '%s' already exists", task.Name)}` and do not save. A
  same-name edit is exempt. Because `FindByName` wraps `FindTaskByName` (case-insensitive
  `EqualFold`), this makes the user-facing uniqueness case-insensitive and also blocks a rename from
  silently overwriting a different task (previously a data-loss bug — see the v1.0 pitfall in
  `internal/task/CLAUDE.md`). `UpdateTask` itself is unchanged (no cross-collision check), so this
  form guard is the mechanism that prevents rename overwrites.

### Constraints

- **CON-001**: The on-disk JSON format MUST remain byte-compatible with the current files. The top
  level is an object keyed by `"tasks"` or `"workspaces"` whose value is the array (or `null` when
  empty). Because the JSON key differs per type and Go struct tags cannot be parametrized, the
  generic repository SHALL serialize via `map[string][]T{jsonKey: items}` and deserialize via
  `map[string][]T` (see section 4.2).
- **CON-002**: `init()` ordering — Go initializes package-level `var` declarations before `init()`
  runs. `TasksSaveFile` / `WorkspacesSaveFile` are assigned **inside** `init()`. Therefore the
  singletons MUST be assigned inside `init()` after the path assignment, NOT as standalone `var`
  initializers (which would capture an empty path).
- **CON-003**: Go version is 1.24.5 (`go.mod`). Generics, builtin `min`/`max`, and `any` are
  available.
- **CON-004**: No new third-party dependency. `github.com/samber/lo` is already available and MAY be
  used inside the generic repository (e.g. `lo.Find`, `lo.Filter`, `lo.FindIndexOf`).
- **CON-005**: After any change under `internal/workspace/` or `pkg/tui/`, the
  `tui-integration-tester` agent MUST be run before reporting done, per the project CLAUDE.md.
  `task_selector.go` is under `internal/workspace/`, so this applies.

### Guidelines

- **GUD-001**: Follow the project's Go element-order rule (constants, types, exported methods by
  call hierarchy via the Stepdown Rule, then unexported helpers).
- **GUD-002**: Keep `Save`/`Update`/`Delete` implemented as `ReadAll` → mutate the pure slice →
  `WriteAll`. This is observably identical to the originals (which reopen/rewrite the whole file)
  and removes the per-method file-handling boilerplate.
- **GUD-003**: Add GoDoc comments to all exported symbols (`JSONRepository`, `NamedEntity`, each
  method, `Set` and its methods, `GetName`).
- **GUD-004**: Run the `simplify` + `clean-code` quality pass at the end, per the user's task
  execution workflow.

### Patterns

- **PAT-001** (Strategy for append): the per-type append behavior is injected as
  `onAppend func(existing []T, item T) ([]T, error)`. As of v1.1 both Task's and Workspace's
  strategies perform the same case-sensitive uniqueness check (REQ-009a) — the seam still exists so a
  future type could append unconditionally, but the two current types are symmetric. (In v1.0 this
  seam encoded the now-removed Task/Workspace asymmetry.)
- **PAT-002** (Factory constructors): `NewTaskRepository(saveFile string)` and
  `NewWorkspaceRepository(saveFile string)` build the concrete singletons, each wiring its `jsonKey`
  and `onAppend`. The path is passed as a parameter (not read from the global) so construction is
  explicit and testable.

## 4. Interfaces & Data Contracts

### 4.1 JSONRepository[T] type and methods

```go
// internal/repository/repository.go
package repository

// NamedEntity is the constraint for items persisted by JSONRepository: any type that
// exposes its unique name. Defined here because this is where it is consumed.
type NamedEntity interface {
    GetName() string
}

// JSONRepository[T] persists a slice of T as a single JSON object file keyed by jsonKey.
type JSONRepository[T NamedEntity] struct {
    saveFile string
    jsonKey  string
    onAppend func(existing []T, item T) ([]T, error)
}

func (r *JSONRepository[T]) ReadAll() ([]T, error)
func (r *JSONRepository[T]) FindByName(name string) (*T, error) // EqualFold (REQ-008)
func (r *JSONRepository[T]) Save(item T) error                  // uses onAppend (REQ-009)
func (r *JSONRepository[T]) Update(originalName string, updated T) error // == match (REQ-010)
func (r *JSONRepository[T]) Delete(name string) error           // != filter (REQ-011)
func (r *JSONRepository[T]) WriteAll(items []T) error           // full replace (REQ-007)

// unexported
func (r *JSONRepository[T]) ensure() // mkdir + create empty file if missing
```

### 4.2 On-disk JSON contract (CON-001)

| Entity    | File path (via `os.UserConfigDir()`)                         | Top-level key  |
| --------- | ------------------------------------------------------------ | -------------- |
| Task      | `<cfg>/vscode-terminal-runner/tasks.json`                    | `"tasks"`      |
| Workspace | `<cfg>/vscode-terminal-runner/workspaces.json`               | `"workspaces"` |

Serialization shape used internally:

```go
// write
content := map[string][]T{r.jsonKey: items}   // -> {"tasks":[...]} or {"tasks":null}
encoded, err := json.Marshal(content)

// read
var content map[string][]T
_ = json.Unmarshal(data, &content)             // empty file -> return nil, nil
return content[r.jsonKey], nil
```

### 4.3 Factory constructors (PAT-002)

```go
func NewTaskRepository(saveFile string) *JSONRepository[models.Task] {
    return &JSONRepository[models.Task]{
        saveFile: saveFile,
        jsonKey:  "tasks",
        onAppend: func(existing []models.Task, t models.Task) ([]models.Task, error) {
            // v1.1: tasks reject duplicates, symmetric with workspaces (REQ-009a)
            if _, found := lo.Find(existing, func(e models.Task) bool {
                return e.Name == t.Name // case-sensitive
            }); found {
                return nil, fmt.Errorf("task '%s' already exists", t.Name)
            }
            return append(existing, t), nil
        },
    }
}

func NewWorkspaceRepository(saveFile string) *JSONRepository[models.Workspace] {
    return &JSONRepository[models.Workspace]{
        saveFile: saveFile,
        jsonKey:  "workspaces",
        onAppend: func(existing []models.Workspace, w models.Workspace) ([]models.Workspace, error) {
            _, found := lo.Find(existing, func(e models.Workspace) bool {
                return e.Name == w.Name // case-sensitive (REQ-009)
            })
            if found {
                return nil, fmt.Errorf("workspace '%s' already exists", w.Name)
            }
            return append(existing, w), nil
        },
    }
}
```

### 4.4 Adapter wiring — call sites that MUST NOT change

The 24 call sites below keep compiling unchanged because only the function bodies are rewritten:

```
cmd/import.go:32,33,44,49
internal/task/task_commands.go:29,77,102
internal/task/task_completion.go:27
internal/task/task_create.go:41,43,109
internal/task/task_list.go:15,46,67
internal/vscode/vscode_runner.go:150,157
internal/workspace/workspace_commands.go:62,101,106
internal/workspace/workspace_completion.go:28
internal/workspace/workspace_create.go:35
internal/workspace/workspace_form.go:324,332,338,433
```

Adapter example (signature identical to current `repository_tasks.go:73`):

```go
func FindTaskByName(name string) (*models.Task, error) {
    return taskRepo.FindByName(name)
}
```

`init()` shape (CON-002), consolidated or per-file — either is acceptable as long as the path is
assigned before the singleton:

```go
func init() {
    cfgFolder, err := os.UserConfigDir()
    if err != nil {
        panic("could not determine user config directory: " + err.Error())
    }
    TasksSaveFile = filepath.Join(cfgFolder, "vscode-terminal-runner", "tasks.json")
    taskRepo = NewTaskRepository(TasksSaveFile)
}
```

### 4.5 Set[T] API and task_selector.go migration map

```go
// pkg/collections/set.go
package collections

type Set[T comparable] struct {
    items map[T]struct{}
}

func NewSet[T comparable]() *Set[T]
func (s *Set[T]) Add(item T)
func (s *Set[T]) Remove(item T)
func (s *Set[T]) Contains(item T) bool
func (s *Set[T]) Toggle(item T) // present -> Remove, absent -> Add
func (s *Set[T]) Len() int
func (s *Set[T]) Clear()
```

Per-site migration in `task_selector.go` (field at line 28, init at line 45):

| Current code                                                  | Line(s) | Replacement                          |
| ------------------------------------------------------------- | ------- | ------------------------------------ |
| `selectedTasks map[string]bool`                               | 28      | `selectedTasks *collections.Set[string]` |
| `make(map[string]bool)`                                       | 45, 74  | `collections.NewSet[string]()`       |
| `ts.selectedTasks[task.Name]` (read)                          | 68, 260, 271 | `ts.selectedTasks.Contains(task.Name)` |
| `ts.selectedTasks[task.Name] = true`                          | 76, 112 | `ts.selectedTasks.Add(task.Name)`    |
| `delete(ts.selectedTasks, task.Name)`                         | 119     | `ts.selectedTasks.Remove(task.Name)` |
| `ts.selectedTasks[task.Name] = !ts.selectedTasks[task.Name]`  | 131     | `ts.selectedTasks.Toggle(task.Name)` |
| `GetSelectedCount` loop counting `true`                       | 81-89   | `return ts.selectedTasks.Len()`      |

## 5. Acceptance Criteria

- **AC-001**: Given the full existing test suite, When the refactor is complete, Then
  `go test ./...` passes with **zero modifications** to any existing test file.
- **AC-002**: Given an existing `tasks.json` written by the pre-refactor binary, When the refactored
  binary reads and rewrites it, Then the resulting bytes are equivalent (`{"tasks":[...]}`), i.e. the
  format is unchanged.
- **AC-003**: Given a workspaces file already containing `"dev"`, When `SaveWorkspace` is called with
  name `"dev"`, Then it returns `workspace 'dev' already exists` and writes nothing.
- **AC-004** (inverted in v1.1): Given a tasks file already containing `"build"`, When `SaveTask` is
  called with name `"build"`, Then it returns `task 'build' already exists` and writes nothing (the
  file still contains exactly one `"build"`). Mirrors AC-003.
- **AC-005**: Given a task named `"Build"`, When `FindTaskByName("build")` is called, Then it returns
  the task (case-insensitive match, REQ-008).
- **AC-006**: Given no task named `"ghost"`, When `DeleteTask("ghost")` is called, Then it returns
  `nil` and the file is unchanged (silent success preserved, REQ-011).
- **AC-007**: Given an `UpdateTask("Old", t)` where `"Old"` (exact case) is absent, Then it returns
  `task 'Old' not found` (case-sensitive, REQ-010).
- **AC-008**: Given the task selector with three available tasks, When the user toggles, selects-all,
  deselects-all, and queries the count, Then the observable behavior (checkbox state, counter) is
  identical to the pre-refactor component, verified end-to-end by `tui-integration-tester`.
- **AC-009**: The system shall expose no changed public signature in `internal/repository`; a
  `git diff` of exported function signatures in that package shows only body changes.
- **AC-010**: `Set[T].Toggle` shall add an absent element and remove a present one;
  `Len` shall reflect only present elements.

### CRUD coverage matrix (both entity types)

Every CRUD operation MUST be proven to work end-to-end for **both** `Task` and `Workspace` through
their public functions. The matrix below is exhaustive; no cell may be left unverified.

- **AC-011** (Create): Given an empty store, When `SaveTask(t)` / `SaveWorkspace(w)` is called with a
  valid item, Then `ReadTasks()` / `ReadWorkspaces()` returns exactly that one item and the on-disk
  file reflects it.
- **AC-012** (Read all): Given a store with N items, When `ReadTasks()` / `ReadWorkspaces()` (and the
  `GetAllTasks()` alias) is called, Then it returns all N items in insertion order.
- **AC-013** (Read one): Given a stored item named `"x"`, When `FindTaskByName("x")` /
  `FindWorkspaceByName("x")` is called, Then it returns a non-nil pointer to that item; and given an
  absent name, Then it returns the not-found error of REQ-012.
- **AC-014** (Update): Given a stored item named `"x"`, When `UpdateTask("x", t2)` /
  `UpdateWorkspace("x", w2)` is called, Then the stored item is replaced by the new value, the item
  count is unchanged, and a subsequent read returns the updated fields (covers both same-name update
  and rename).
- **AC-015** (Delete): Given a store containing `"x"` among others, When `DeleteTask("x")` /
  `DeleteWorkspace("x")` is called, Then `"x"` is gone and all other items remain; deleting an absent
  name is a silent no-op (REQ-011).
- **AC-016** (Round-trip persistence): For each entity type, a Create → Read → Update → Read →
  Delete → Read sequence shall leave the store in the expected state at every step, with the JSON
  file remaining valid and format-compatible (CON-001) throughout.

### Task name uniqueness criteria (v1.1)

- **AC-017** (Create collision, form): Given a task named `"build"` exists, When the user submits the
  create form with name `"build"`, Then the form shows `task 'build' already exists`, does not quit,
  and the store still holds exactly one `"build"`. Verified end-to-end by `tui-integration-tester`.
- **AC-018** (Rename collision + same-name exemption, form): Given tasks `"build"` and `"lint"` exist,
  When the user edits `"lint"` and renames it to `"build"`, Then the form shows
  `task 'build' already exists` and `"build"` is not overwritten (its `Cmds` are intact); AND When the
  user edits `"build"` keeping the name `"build"` and changes only its commands, Then the update
  succeeds with no false collision error. Verified end-to-end by `tui-integration-tester`.

## 6. Test Automation Strategy

- **Test Levels**: Unit (repository CRUD, Set behavior), Integration (TUI via `tui-integration-tester`).
- **Frameworks**: Go standard `testing`. Existing helpers in `pkg/testutils`. The existing
  `repository_tasks_test.go` and `repository_workspaces_test.go` are the primary characterization
  net and run unchanged.
- **New unit tests**:
  - `pkg/collections/set_test.go` — Add/Remove/Contains/Toggle/Len/Clear, including idempotent Add
    and Toggle round-trips.
  - Optionally a small `repository_test.go` exercising the generic directly with a fake `NamedEntity`
    to document the constraint, but this is secondary to keeping the existing suites green.
- **Test Data Management**: existing tests already redirect `TasksSaveFile`/`WorkspacesSaveFile` to
  temp paths; the singletons must honor those paths. NOTE (see EDGE-003): if a test reassigns the
  global path after `init()`, the singleton captured the old path. Section 9 specifies the required
  handling.
- **CI/CD Integration**: `go build -o bin/vstr` and `go test ./...` must both pass.
- **Coverage Requirements**: no regression versus current coverage; new `Set` type covered ≥ 90%.
- **Performance Testing**: not applicable (local file I/O, single-threaded CLI).

## 7. Rationale & Context

- The two repository files are ~232 and ~243 lines and ~89% structurally identical. Generics are the
  exact tool for "same logic, different type". The duplication was independently flagged by multiple
  exploration passes as the highest-ROI candidate.
- The append asymmetry (REQ-009) is the only true behavioral divergence between Task and Workspace
  persistence. Encoding it as an injected Strategy (PAT-001) keeps the generic core pure and avoids a
  type switch.
- The `map[string][]T` serialization (CON-001) is the idiomatic way to keep a dynamic top-level JSON
  key while staying generic; a parametrized struct tag is impossible in Go.
- `ReadAll → mutate → WriteAll` (GUD-002) is chosen over replicating the original per-method
  open/seek/truncate dance because the observable result (whole-file rewrite) is identical and the
  code is dramatically simpler. The original `DeleteTask` used `Truncate+Seek+Write` while `SaveTask`
  used `os.WriteFile`; unifying on whole-file write removes that internal inconsistency with no
  external effect.
- `Set[T]` is low-risk, high-clarity: it turns `map[string]bool` (which conflates "absent" and
  "present-but-false") into an explicit set whose `Toggle`/`Len` read as intent.
- Pre-existing internal inconsistencies that this refactor incidentally cleans up (all
  behavior-neutral): `path` vs `filepath` package usage between the two files, and manual index loop
  (`replaceTaskByName`) vs `lo.FindIndexOf` (`replaceWorkspaceByName`).

## 8. Dependencies & External Integrations

### Technology Platform Dependencies
- **PLT-001**: Go 1.24.5 — required for generics. Constraint from `go.mod`.

### Data Dependencies
- **DAT-001**: `tasks.json` / `workspaces.json` in the user config dir — format MUST stay
  byte-compatible (CON-001).

### Third-Party Services
- **SVC-001**: `github.com/samber/lo` — already vendored; reused inside the generic repository. No
  new dependency introduced (CON-004).

### Compliance Dependencies
- **COM-001**: Project CLAUDE.md mandates running `tui-integration-tester` after changes under
  `internal/workspace/` (CON-005).

## 9. Examples & Edge Cases

- **EDGE-001 (empty file)**: `ReadAll` on an empty or whitespace-only file returns `(nil, nil)` —
  matches `repository_tasks.go:62-69` which leaves `content.Tasks` as the zero (nil) slice.
- **EDGE-002 (invalid JSON)**: `ReadAll` returns the `json.Unmarshal` error unwrapped, matching the
  current behavior (`repository_tasks.go:64-66`).
- **EDGE-003 (test path reassignment / singleton staleness)**: Existing tests set the package global
  `TasksSaveFile`/`WorkspacesSaveFile` to a temp path. Two acceptable resolutions, pick one and apply
  consistently:
  - **(a) Preferred:** make the adapter read the path lazily — store a `getSaveFile func() string` in
    the repository instead of a captured string, wired as `func() string { return TasksSaveFile }`,
    so a test that reassigns the global is honored.
  - **(b) Alternative:** if and only if every existing test sets the path *before* the first
    repository call within the same process and never relies on post-`init` reassignment, the
    captured-string form is acceptable. VERIFY against `repository_tasks_test.go` /
    `repository_workspaces_test.go` before choosing this; if any test mutates the global between
    calls, option (a) is mandatory.

  This is the single highest-risk integration detail. Resolve it first, before writing the generic.
- **EDGE-004 (Toggle on map[string]bool semantics)**: the original `ToggleSelected` sets
  `m[name] = !m[name]`, leaving a `false` entry in the map; `GetSelectedCount` only counts `true`
  and `GetSelectedTasks` only keeps `true`. `Set.Toggle` (remove-if-present) is observably
  equivalent because both `Contains` and `Len` treat a removed key and a `false` key identically.

```go
// Example: generic Update body (case-sensitive match, REQ-010)
func (r *JSONRepository[T]) Update(originalName string, updated T) error {
    items, err := r.ReadAll()
    if err != nil {
        return err
    }
    _, index, found := lo.FindIndexOf(items, func(item T) bool {
        return item.GetName() == originalName // case-sensitive
    })
    if !found {
        return fmt.Errorf("%s '%s' not found", r.jsonKey /* or a stored entity label */, originalName)
    }
    result := make([]T, len(items))
    copy(result, items)
    result[index] = updated
    return r.WriteAll(result)
}
```

NOTE on the error label: the originals say `"task '%s' not found"` / `"workspace '%s' not found"`
(singular), whereas `jsonKey` is plural (`"tasks"`). To preserve REQ-010/REQ-012 messages exactly,
store a singular `entityLabel` (`"task"` / `"workspace"`) on the struct, OR keep the precise message
in the adapter rather than the generic. The exact strings in AC-003/AC-007 are the contract.

## 10. Validation Criteria

1. `go build -o bin/vstr` succeeds.
2. `go test ./...` passes with no edits to existing test files (AC-001).
3. `git diff` shows no signature change for any exported symbol in `internal/repository` (AC-009).
4. The on-disk format is unchanged (AC-002).
5. All asymmetry ACs pass: AC-003, AC-004, AC-005, AC-006, AC-007.
6. **Every cell of the CRUD coverage matrix passes for both entity types**: AC-011 through AC-016.
   Each of the five operations (Create, Read-all, Read-one, Update, Delete) is exercised against both
   `Task` and `Workspace` via their public functions, plus the full round-trip (AC-016). No
   operation/type cell may be left unverified.
7. `Set` unit tests pass (AC-010) at ≥ 90% coverage.
8. `tui-integration-tester` confirms the task selector behaves identically (AC-008).
9. `simplify` + `clean-code` pass run with no outstanding findings (GUD-004).

## 11. Related Specifications / Further Reading

- `internal/repository/repository_tasks.go`, `repository_workspaces.go`, `repository_import.go` —
  the code being refactored.
- `internal/workspace/components/task_selector.go` — the `Set[T]` consumer.
- `internal/models` — `Task`, `Workspace` (gain `GetName()`).
- Project memory `[[task-delete-silent-success-debt]]` — documents the silent-delete behavior that
  REQ-011 must preserve.
- Project CLAUDE.md — TUI test mandate (CON-005) and Go standards.

## Branch & Commit Strategy

Both components ship on **one dedicated branch** (name to be set via the `branching` skill at
implementation time). Suggested bisectable commit groups:

1. `feat: add Set[T] generic in pkg/collections` (+ its unit tests) — independent, lands first.
2. `feat: add JSONRepository[T] and NamedEntity` (generic core + GetName on models + factories).
3. `refactor: delegate repository public funcs to JSONRepository` (adapters + init wiring + Import).
4. `refactor: migrate task selector to Set[string]`.

Each group must leave `go build` and `go test ./...` green.
