---
name: project-task-tui-patterns
description: Established patterns in the task and workspace TUI forms as of June 2026 audit
metadata:
  type: project
---

Audited internal/task/ and supporting packages on 2026-06-12.

**Why:** Baseline audit to identify design violations before further feature development.

**How to apply:** Reference these as the known pattern baseline when auditing future changes or proposing new forms.

## Confirmed patterns in use

- `FormNavigator` from `pkg/tui` handles Tab/Shift+Tab/Up/Down field cycling. `HandleNavigation` takes raw `msg.String()` — this works because the switch matches literal strings "up", "down", "tab", "shift+tab".
- All key matching uses `key.Matches(msg, tui.DefaultKeys.X)` — no raw string comparisons.
- Async I/O (path validation + repository save) runs inside `submitTaskCmd` tea.Cmd — correctly off the UI goroutine.
- Success messages added to `MessageManager` before `tea.Quit` — messages persist for caller to read.
- `pkg/messages.MessageManager` used for all error/success feedback in TUI.
- `WindowSizeMsg` handled in both task and workspace forms, resizing inputs dynamically.

## Known issues found in this audit

- `PathManager.UpdateFilter` calls `os.ReadDir` synchronously from within `HandleInput` which is called from `Update`. This is blocking filesystem I/O on the UI goroutine (T-001 critical).
- `DeleteCmd` body in `task_commands.go:70` prints "Deleting task:" but does NOT call `DeleteTask` or `repository.DeleteTask` — the delete is a no-op stub (T-002 critical).
- All Lipgloss style widths in `pkg/styles` (forms.go, messages.go) are hard-coded integers (58, 70, 86, 90) not responsive to terminal size (T-003 high).
- No spinner shown during async submit (`submitTaskCmd`) — user gets no feedback between pressing Enter and the form quitting (T-004 high).
- `strings.Split(cmdsField.Value(), ",")` in `handleTaskCreation` does not trim whitespace from individual commands — "cmd1, cmd2" produces ["cmd1", " cmd2"] with a leading space (T-005 medium).
- `centerText` in `pkg/styles/forms.go` uses `len(text)` (byte count) not `lipgloss.Width` — breaks for multi-byte UTF-8 titles (T-006 medium).
- `DeleteCmd` uses raw `fmt.Println` for status, not `styles.PrintError`/`styles.PrintSuccess` (T-007 low — consistent with ListCmd which also uses fmt.Println for non-TUI output).
- `os.Getenv("HOME")` called twice per task in `listAllTasks` loop without caching (T-008 low).
