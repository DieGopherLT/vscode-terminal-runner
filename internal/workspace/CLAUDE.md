# internal/workspace

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Bubbletea TUI for workspace CRUD and execution. A workspace is a named group of tasks run together. Exposes Cobra commands and an interactive form with a reusable multi-select task selector component.

## Entry Points

- `workspace_commands.go::CreateCmd` - Cobra: opens TUI to create a new workspace
- `workspace_commands.go::RunCmd` - Cobra: runs workspace by name via `SecureRunner`
- `workspace_commands.go::ListCmd` - Cobra: lists workspaces (stub, not yet implemented)
- `workspace_form.go::NewWorkspaceModel` - TUI model constructor for create mode
- `workspace_form.go::NewEditWorkspaceModel` - TUI model constructor for edit mode (pre-fills fields)
- `workspace_create.go::CreateWorkspaceCommand` - Wraps TUI program; entry point from CLI
- `workspace_create.go::EditWorkspaceCommand` - Loads workspace, wraps TUI program for editing
- `components/task_selector.go::NewTaskSelector` - Reusable multi-select component with search

## Key Files

- **workspace_form.go**: `WorkspaceModel` struct; `Init`, `Update`, `View`; validation; save logic
- **workspace_commands.go**: Cobra command definitions with argument validation
- **workspace_create.go**: Runs `tea.NewProgram`; bridges CLI args to TUI
- **components/task_selector.go**: `TaskSelector`; search, filter, toggle, select-all, scroll

**Note**: Symbol references use LSP-optimized format (`file::Symbol`) for:

- `goToDefinition`: Jump directly to symbol location
- `findReferences`: Find all real usages (zero false positives)
- `hover`: Get type info and documentation instantly
- `documentSymbol`: Navigate file structure without reading full content

## Business Logic

**Form fields (FocusIndex):**

| Index | Field | Notes |
|-------|-------|-------|
| 0 | Workspace Name | Required; validated for empty and duplicates |
| 1 | Task Selector | Multi-select; warns but allows empty selection |
| 2 | Submit button | `FocusIndex == elementCount` (2) triggers submit on Enter |

**Submission flow:** Enter on Submit -> `isValidWorkspace` -> clear messages -> validate name -> warn if no tasks -> if valid, `saveWorkspace` -> `repository.SaveWorkspace` -> `tea.Quit`.

**Edit mode:** if name changed, `saveWorkspace` calls `repository.DeleteWorkspace(originalName)` first, then saves new. `originalWorkspaceName` stored at init.

**Task selector interaction:**

- `/` key: toggles search mode; all keystrokes go to search input
- Space: `ToggleSelected` on focused item
- Ctrl+A / Ctrl+D: `SelectAll` / `DeselectAll`
- Esc/Enter: exits search mode; resets to full list
- Search matches task name and path (case-insensitive)
- Scrolls when list > `maxHeight=6` items

**Validation guard clauses:**

1. Name empty -> error + refocus to name field
2. Duplicate name (skipped in edit if unchanged) -> error + refocus
3. No tasks selected -> warning (non-blocking)

**Messages:** cleared when user types in name field; persist across navigation; `Clear()` called on each submit attempt before re-validation.

## Dependencies

**Internal:**

- `internal/models`: `Task`, `Workspace` structs
- `internal/repository`: `GetAllTasks`, `FindWorkspaceByName`, `SaveWorkspace`, `DeleteWorkspace`
- `internal/vscode`: `NewSecureRunner` -> `RunWorkspace` for execution
- `pkg/tui`: `FormNavigator` for field focus cycling
- `pkg/messages`: `MessageManager` — `AddError`, `AddWarning`, `AddSuccess`, `Render`
- `pkg/styles`: `RenderTitle`, `FieldLabelStyle`, `RenderFocusedButton`, `RenderBlurredButton`, `FocusedTaskStyle`, `SelectedTaskStyle`, `TaskSelectorContainerStyle`

**External:**

- `charmbracelet/bubbletea`: TUI framework
- `charmbracelet/bubbles/textinput`: Name input and search input fields
- `charmbracelet/lipgloss`: Styling and vertical layout
- `samber/lo`: `lo.Filter`, `lo.CountBy`, `lo.Find`
- `spf13/cobra`: CLI command registration

## Architecture

- **Two constructors, one internal**: `NewWorkspaceModel` and `NewEditWorkspaceModel` both delegate to `newWorkspaceModelInternal(workspace)` — reuse with optional pre-fill.
- **TaskSelector as sub-component**: encapsulates all selection/search state; `WorkspaceModel` calls `taskSelector.Update(msg)` and `taskSelector.View()` without owning selection logic.
- **Submit button at `FocusIndex == elementCount`**: navigator cycles 0->1->2->0; rendering checks `FocusIndex >= elementCount` for button focus state.
- **Search mode gates navigation**: `Update` checks `taskSelector.IsInSearchMode()` early; if true, only Esc/Enter bypass the search input.

## Modification Guide

### Adding Features

- **New form field**: increment `elementCount` in `newWorkspaceModelInternal`; add field index constant; add input init; update `View` labels; update `handleSubmit`; add validation in `isValidWorkspace`
- **New validation rule**: add guard clause in `isValidWorkspace` using `w.messages.AddError(msg)` + early return `false`
- **New task selector key binding**: add case in `TaskSelector.Update` following existing toggle/select-all pattern

### Removing Code

- Removing `ListCmd`: also remove registration in `cmd/workspace.go`
- Removing `TaskSelector` search: remove search input, `showSearch` flag, and search key handlers from `Update`

### Common Pitfalls

- `focusedIndex` is into `filteredTasks`, not `availableTasks` -> always index `ts.filteredTasks[focusedIndex]`; use task name to cross-reference with `availableTasks`
- Edit mode rename: `saveWorkspace` must delete old name first; if delete errors, save still proceeds (partial failure possible)
- Focus index `== elementCount` is valid (submit button); rendering must handle `FocusIndex >= elementCount` explicitly
- Search mode blocks navigation: always check `IsInSearchMode()` in `WorkspaceModel.Update` before routing key to navigator

## Usage Examples

```go
// Create mode
p := tea.NewProgram(workspace.NewWorkspaceModel())
p.Run()

// Edit mode
ws, _ := repository.FindWorkspaceByName("dev-setup")
p := tea.NewProgram(workspace.NewEditWorkspaceModel(ws))
p.Run()

// Run via CLI
runner, err := vscode.NewSecureRunner()
if err != nil {
    return err
}
return runner.RunWorkspace("dev-setup")
```

---

## Claude's Navigation Commitment

This CLAUDE.md is my map for navigating this module. I commit to:

- **Update immediately** after any code modification in this module
- **Verify accuracy** of all symbol references after each change
- **Maintain truth** - outdated documentation is a critical bug
- **Treat this as my compass** - if this map is wrong, I'm lost

Last verified: 2026-02-18
