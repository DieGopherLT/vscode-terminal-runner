# internal/task

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Bubbletea TUI for task CRUD and execution. Exposes Cobra commands and interactive forms that persist tasks via `internal/repository` and run them through `internal/vscode.SecureRunner`.

## Entry Points

- `task.go::NewModel` - TUI model constructor for create mode
- `task.go::NewEditModel` - TUI model constructor for edit mode (pre-fills fields)
- `task_commands.go::CreateCmd` - Cobra: interactive create or batch from `--file`
- `task_commands.go::ListCmd` - Cobra: tabular list or `--only-names` compact view
- `task_commands.go::EditCmd` - Cobra: opens pre-filled form for existing task
- `task_commands.go::DeleteCmd` - Cobra: deletes task by name
- `task_commands.go::RunCmd` - Cobra: executes task via SecureRunner
- `task_create.go::DeleteTask` - Helper wrapping `repository.DeleteTask`
- `task_list.go::FindByName` - Helper wrapping `repository.FindTaskByName`

## Key Files

- **task.go**: `TaskModel` struct, field index constants, constructors, `newModelInternal`
- **task_form.go**: `Init`, `Update`, `View` — core TUI loop, suggestion routing, key handling
- **task_commands.go**: Cobra command definitions with flag registration
- **task_create.go**: Validation (`isValidTask`), form-to-struct (`handleTaskCreation`), path expansion, `saveTask`, batch import
- **task_list.go**: Tabular and name-only list rendering

**Note**: Symbol references use LSP-optimized format (`file::Symbol`) for:

- `goToDefinition`: Jump directly to symbol location
- `findReferences`: Find all real usages (zero false positives)
- `hover`: Get type info and documentation instantly
- `documentSymbol`: Navigate file structure without reading full content

## Business Logic

**Form fields (indices 0-4):**

| Index | Field | Notes |
|-------|-------|-------|
| 0 | Name | Required; unique key for update |
| 1 | Path | Optional; `~`, relative, absolute — validated with `os.Stat` after tilde expansion |
| 2 | Commands | Required; comma-separated, split on save |
| 3 | Icon | Must match `styles.VSCodeIcons`; autocomplete via `IconSuggestions` |
| 4 | IconColor | Must match `styles.VSCodeANSIColors`; autocomplete via `ColorSuggestions` |

Index `5` is the Submit button (navigation wraps 0..5).

**Submission flow:** Enter on Submit -> `handleTaskCreation` -> `isValidTask` (errors to `MessageManager`) -> if valid, `saveTask` (create or update by `originalTaskName`) -> `tea.Quit`.

**Suggestion lifecycle:** typing updates filter; Ctrl+N/Ctrl+B cycles; Tab/Enter applies. Navigation resets manager. `getCurrentSuggestionManager` routes by `FocusIndex`.

**Edit mode:** `isEditMode=true` + `originalTaskName` stored at init; `saveTask` calls `repository.UpdateTask(originalTaskName, ...)` so rename works correctly.

**Batch create:** `--file <path>` skips TUI and calls `repository.SaveFromFile(path)` with a JSON array of tasks.

## Dependencies

**Internal:**

- `internal/models`: `Task` struct
- `internal/repository`: `SaveTask`, `UpdateTask`, `DeleteTask`, `FindTaskByName`, `SaveFromFile`
- `internal/vscode`: `NewSecureRunner` for task execution
- `pkg/tui`: `FormNavigator` for field focus cycling
- `pkg/tui/suggestions`: `Manager` (icon/color), `PathManager` (filesystem dirs)
- `pkg/messages`: `MessageManager` for error/success display
- `pkg/styles`: Lipgloss styles, `VSCodeIcons`, `VSCodeANSIColors`

**External:**

- `charmbracelet/bubbletea`: TUI framework
- `charmbracelet/bubbles/textinput`: Input field components
- `charmbracelet/lipgloss`: Terminal styling
- `samber/lo`: `lo.Map`, `lo.Filter`, `lo.Find`
- `spf13/cobra`: CLI command registration

## Architecture

Strict Bubbletea MUV pattern. `TaskModel` holds all state: `[]textinput.Model`, three suggestion managers, `FormNavigator`, `MessageManager`, and edit-mode flags. No goroutines; all I/O is synchronous inside `Update`. View is pure function of model state.

## Modification Guide

### Adding Features

- **New field**: increment `numberOfFields` in `newModelInternal`; add index constant; add `textinput` init; update labels in `View`; update `handleTaskCreation`; add validation in `isValidTask`; wire suggestion manager in `getCurrentSuggestionManager` if needed
- **New validation rule**: add to `isValidTask` using `t.messages.AddError(msg)`; return `false` if `t.messages.HasErrors()`
- **New suggestion source**: create `suggestions.Manager` with chosen `FilterFunc`; register in `getCurrentSuggestionManager`

### Removing Code

- Removing a field: also remove its index constant, label, validation, and suggestion manager wiring
- Removing a command: also deregister from `cmd/task.go`

### Common Pitfalls

- Tilde path not expanded before `os.Stat` -> validation always fails for `~` paths. Use `expandPathForValidation` first.
- Edit mode: always use `originalTaskName` as key in `UpdateTask`, not the current form value, or rename silently creates a duplicate.
- Suggestion + Tab collision: Tab checks and applies suggestion first, then navigates only if no suggestion was applied. Don't change this order.
- Name uniqueness on rename: `UpdateTask` replaces by `originalTaskName`; there is no uniqueness check, so renaming to an existing task overwrites it silently.

## Usage Examples

```go
// Create mode
p := tea.NewProgram(task.NewModel())
p.Run()

// Edit mode
existing, _ := task.FindByName("build")
p := tea.NewProgram(task.NewEditModel(existing))
p.Run()

// Delete (non-TUI)
err := task.DeleteTask("build")
```

---

## Claude's Navigation Commitment

This CLAUDE.md is my map for navigating this module. I commit to:

- **Update immediately** after any code modification in this module
- **Verify accuracy** of all symbol references after each change
- **Maintain truth** - outdated documentation is a critical bug
- **Treat this as my compass** - if this map is wrong, I'm lost

Last verified: 2026-02-18
