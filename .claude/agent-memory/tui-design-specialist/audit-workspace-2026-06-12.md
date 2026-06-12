---
name: audit-workspace-2026-06-12
description: Recurring issues found during workspace CRUD/run TUI audit — IDs W-001 through W-008
metadata:
  type: project
---

## Audit: internal/workspace — 2026-06-12

### Issues found

- **W-001 (high)**: `workspace_form.go:104-114` — Navigation key handling for nameField is inverted: `nameInput.Update(msg)` is called first for ALL messages (including navigation keys), then `handleKeyPress` is called again for the same event. Navigation keystrokes feed the text input AND trigger navigation in the same Update cycle.
- **W-002 (high)**: `workspace_form.go:87-89` — `tea.WindowSizeMsg` does not propagate to `taskSelector`; `TaskSelectorContainerStyle` has a hard-coded `Width(90)` in forms.go:66 and `separatorLineLength = 88` in task_selector.go:20. Form does not respond to narrow terminals.
- **W-003 (medium)**: `forms.go:62-67` — `TaskSelectorContainerStyle` sets explicit `Height(9)` on a Lipgloss bordered style. This violates Golden Rule #1 (never set Height on a bordered style — fill content to exact height instead). Clipping occurs when search box is shown inside the container.
- **W-004 (medium)**: `workspace_commands.go:20` — `RunCmd` uses `vscode.NewRunner()` while the CLAUDE.md for internal/workspace says `NewSecureRunner`; potential security regression vs task package.
- **W-005 (medium)**: `workspace_form.go:353-386` — `View()` renders help text at the bottom of the form, but the help bar in the task selector (`renderHelpText` in task_selector.go) also renders duplicated help text inside the container. Both are visible simultaneously; information is repeated.
- **W-006 (low)**: `task_selector.go:238-265` — `renderTaskItem` uses `len(task.Name)` and `len(task.Path)` for padding/truncation. Multi-byte Unicode characters (e.g., CJK) will misalign columns because `len()` counts bytes, not runes or display cells.
- **W-007 (low)**: `workspace_commands.go:39` — `ListCmd` prints to stdout via `fmt.Println` inside a Cobra handler. This is consistent with the project's current pattern for non-TUI commands, so not a Bubbletea violation, but the command is an unimplemented stub with a hardcoded TODO message.
- **W-008 (low)**: `workspace_form.go:376` — Submit button has no spinner; when `submitWorkspaceCmd` is executing (file I/O in-flight), the form is interactive (user can press keys again). The task package has the same gap — but worth noting.

**Why recorded:** These patterns may recur in future workspace or task TUI work. Use these IDs for tracking.
**How to apply:** When touching workspace TUI code, check these issues for regressions.
