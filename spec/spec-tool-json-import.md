---
title: JSON Import for Tasks and Workspaces (--from-json)
version: 1.0
date_created: 2026-06-12
last_updated: 2026-06-12
owner: DieGopherLT
tags: [tool, cli, import, repository]
---

# Introduction

This specification defines a JSON import capability for the `vstr` CLI. It replaces the
partially implemented `--file` flag with a `--from-json` flag that accepts either a file
path or standard input (`-`), validates the entire payload before writing anything
(all-or-nothing), and extends import coverage to workspaces and to a unified `vstr import`
command that ingests tasks and workspaces together.

## 1. Purpose & Scope

Allow a user to create tasks and/or workspaces in bulk from a JSON source instead of the
interactive TUI. Three entry points are in scope:

1. `vstr task create --from-json <source>` — imports a JSON **array of Task**.
2. `vstr workspace create --from-json <source>` — imports a JSON **array of Workspace**.
3. `vstr import --from-json <source>` — imports a JSON **object** `{ "tasks": [...], "workspaces": [...] }`.

Out of scope: editing/updating existing entries via import, partial/merge imports, and any
network or bridge interaction. Import is a local persistence operation only.

**Disruptive change (approved):** the existing `--file` / `-f` flag on `task create` and its
backing function `repository.SaveFromFile` are **removed and replaced**. The project is
pre-release; no backward-compatibility alias is kept.

## 2. Definitions

- **Source**: the value passed to `--from-json`. Either a filesystem path or the literal `-`.
- **stdin source**: when the source is `-`, JSON is read from standard input (POSIX Utility
  Syntax Guideline 13 convention), enabling `echo '[...]' | vstr task create --from-json -`.
- **All-or-nothing**: the full payload is validated before any write; on any error nothing is
  persisted and every detected problem is reported.
- **Disk collision**: an imported entry whose `name` already exists in the persistence file.
- **Intra-batch collision**: two entries in the same payload sharing the same `name`.
- **Task / Workspace**: data types defined in `internal/models` (see section 4).

## 3. Requirements, Constraints & Guidelines

- **REQ-001**: `vstr task create --from-json <source>` SHALL parse `<source>` as a JSON array
  of Task and persist all of them when valid.
- **REQ-002**: `vstr workspace create --from-json <source>` SHALL parse `<source>` as a JSON
  array of Workspace and persist all of them when valid.
- **REQ-003**: `vstr import --from-json <source>` SHALL parse `<source>` as a JSON object with
  optional `tasks` and `workspaces` arrays and persist both when valid.
- **REQ-004**: `<source>` SHALL be read from the file at that path, except when it is exactly
  `-`, in which case it SHALL be read from `os.Stdin`.
- **REQ-005**: The shorthand for `--from-json` SHALL be `-j` (not `-f`, which belonged to the
  removed `--file` flag).
- **REQ-006**: Validation SHALL run over the entire payload BEFORE any write. On any failure,
  no file SHALL be modified.
- **REQ-007**: On validation failure, the command SHALL print every detected problem (not just
  the first) and exit with a non-zero status.
- **REQ-008**: On success, the command SHALL print a success message and exit zero.
- **VAL-001**: A Task entry is invalid if `name` is empty or `cmds` is empty (zero commands).
- **VAL-002**: A Workspace entry is invalid if `name` is empty.
- **VAL-003**: If a Task entry provides a non-empty `icon`, it MUST exist in
  `styles.VSCodeIcons`; if it provides a non-empty `iconColor`, it MUST exist in
  `styles.VSCodeANSIColors`. Empty values are allowed and left as-is.
- **VAL-004**: A disk collision (VAL: name already present in `tasks.json` / `workspaces.json`)
  SHALL be an error.
- **VAL-005**: An intra-batch collision SHALL be an error.
- **CON-001**: `repository.SaveFromFile` (`internal/repository/repository_tasks.go:134`) and
  `appendTaskBatch` (`:184`) SHALL be removed; the blind-append behavior they implement
  violates VAL-004/VAL-005.
- **CON-002**: The `--file` / `-f` flag registration in `internal/task/task_commands.go:155`
  and its handling block (`:21-31`) SHALL be removed.
- **CON-003**: Reading source and parsing JSON MUST be separated from validation and writing,
  so the same validation path serves file and stdin inputs.
- **GUD-001**: Field validation SHOULD live as pure functions reusable by all three entry
  points — recommended as `Validate() error` methods on `models.Task` and `models.Workspace`,
  or as `repository`-level helpers. Do NOT reuse `TaskModel.isValidTask`
  (`internal/task/task_create.go:75`); it is bound to the TUI `MessageManager`.
- **PAT-001**: Follow the existing persistence read-merge-write pattern already used by
  `SaveTask` / `SaveWorkspace` (read file -> mutate slice -> `os.WriteFile`).

## 4. Interfaces & Data Contracts

### Data types (existing, `internal/models`)

```go
type Task struct {
    Name      string   `json:"name"`
    Path      string   `json:"path"`
    Cmds      []string `json:"cmds"`
    Icon      string   `json:"icon"`
    IconColor string   `json:"iconColor"`
}

type Workspace struct {
    Name  string `json:"name"`
    Tasks []Task `json:"tasks"`
}
```

### Persistence (existing, `internal/repository`)

- `TasksSaveFile` (`repository_tasks.go:29`) — path to `tasks.json`.
- `TaskSaveFileContent` (`repository_tasks.go:47`) — `{ Tasks []models.Task }`.
- `ReadTasks() ([]models.Task, error)` (`repository_tasks.go:52`).
- `ReadWorkspaces() ([]models.Workspace, error)` (`repository_workspaces.go:48`).
- `SaveWorkspace(models.Workspace) error` (`repository_workspaces.go:90`) — already enforces
  name uniqueness via `appendWorkspaceIfUnique` (`:171`); reuse its collision semantics.

### New repository functions (to implement)

```go
// ImportTasks validates the whole batch (VAL-001, VAL-003, VAL-004, VAL-005) and writes
// all entries only if every entry is valid. On failure it returns an error that aggregates
// every problem and writes nothing.
func ImportTasks(tasks []models.Task) error

// ImportWorkspaces mirrors ImportTasks for workspaces (VAL-002, VAL-004, VAL-005).
func ImportWorkspaces(workspaces []models.Workspace) error
```

### New shared source reader (to implement)

```go
// OpenSource returns a reader for the given --from-json value: os.Stdin when source == "-",
// otherwise the opened file. Caller closes when it is not stdin.
func OpenSource(source string) (io.ReadCloser, error)
```

### CLI surface

```
vstr task create --from-json <path|->        # -j alias
vstr workspace create --from-json <path|->   # -j alias
vstr import --from-json <path|->             # -j alias
```

### Unified import payload (`vstr import`)

```json
{
  "tasks":      [ { "name": "...", "cmds": ["..."], "icon": "", "iconColor": "", "path": "" } ],
  "workspaces": [ { "name": "...", "tasks": [ { "name": "..." } ] } ]
}
```

Both keys are optional; an absent key imports nothing for that type.

## 5. Acceptance Criteria

- **AC-001**: Given a valid JSON array of tasks in a file, When `vstr task create --from-json
  ./tasks.json` runs, Then every task is appended to `tasks.json` and a success message prints.
- **AC-002**: Given a JSON array piped to stdin, When `echo '[...]' | vstr task create
  --from-json -` runs, Then the tasks are imported identically to the file case.
- **AC-003**: Given a payload where one task name already exists on disk, When import runs,
  Then nothing is written and the error names the colliding task.
- **AC-004**: Given a payload containing two tasks with the same name, When import runs, Then
  nothing is written and the error names the duplicated name.
- **AC-005**: Given a payload where one task has empty `cmds`, When import runs, Then nothing
  is written and the error identifies that entry.
- **AC-006**: Given multiple problems in one payload, When import runs, Then ALL problems are
  reported in a single run, not just the first.
- **AC-007**: Given `vstr import --from-json` with both `tasks` and `workspaces` valid, When it
  runs, Then both files receive their entries.
- **AC-008**: The `--file` flag SHALL no longer be accepted by `vstr task create` (unknown flag).
- **AC-009**: Given an invalid icon on a task, When import runs, Then nothing is written and
  the error names the invalid icon and the offending task.

## 6. Test Automation Strategy

- **Test Levels**: Unit (repository import + validation), Integration (Cobra command wiring).
- **Frameworks**: Go standard `testing`. Follow existing table-driven style in
  `internal/repository/*_test.go` and `internal/task/task_completion_test.go`.
- **Test Data Management**: Use `pkg/testutils` helpers and temp config dirs; never touch the
  real `$XDG_CONFIG_HOME`. Override the save-file paths as the existing repo tests do.
- **Coverage Requirements**: All validation branches (VAL-001..VAL-005) and the source reader
  (`-` vs path vs missing file) must be exercised.
- **CI/CD Integration**: `go test ./...` must pass.

## 7. Rationale & Context

The current `--file` flag (`SaveFromFile`) blind-appends via `appendTaskBatch`, allowing
silent duplicates and invalid entries — the opposite of the desired all-or-nothing contract.
Rather than patch it, it is replaced by a validated `--from-json` that also covers workspaces
and a unified import. `-` for stdin follows POSIX Guideline 13, keeping a single flag for both
"string via pipe" and "file" inputs. Validation is extracted to pure functions because the
existing `isValidTask` is coupled to the TUI `MessageManager` and cannot be reused headless.

## 8. Dependencies & External Integrations

### Technology Platform Dependencies

- **PLT-001**: Go 1.24+, `spf13/cobra` for command/flag registration.

### Data Dependencies

- **DAT-001**: `tasks.json` and `workspaces.json` under `$XDG_CONFIG_HOME/vscode-terminal-runner/`.

## 9. Examples & Edge Cases

```bash
# File import
vstr task create --from-json ./tasks.json

# Stdin import (string via pipe)
echo '[{"name":"api","cmds":["go run ."]}]' | vstr task create --from-json -

# Unified import
vstr import --from-json ./project-config.json
```

Edge cases:

- Empty array `[]` -> valid no-op, success with "0 imported".
- `-` with no piped data -> read yields empty input -> JSON parse error reported.
- Missing file path -> open error reported, nothing written.
- `vstr import` with neither `tasks` nor `workspaces` -> valid no-op.
- Cross-file atomicity in `vstr import`: validate BOTH arrays fully before writing either; if
  the second write fails after the first succeeded, report the partial state explicitly.

## 10. Validation Criteria

- `--file` removed; `--from-json` (`-f`) present on `task create`, `workspace create`, and `import`.
- No path writes occur on any validation failure (verified by asserting file contents unchanged).
- Every problem in a multi-error payload is surfaced in one run.

## 11. Related Specifications / Further Reading

- `spec-tool-status-command.md`
- `spec-tool-version-command.md`
- POSIX Base Definitions, Utility Syntax Guidelines, Guideline 13 (stdin via `-`).
