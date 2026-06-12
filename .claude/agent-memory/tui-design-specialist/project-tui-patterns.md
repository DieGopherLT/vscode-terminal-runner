---
name: project-tui-patterns
description: Established TUI patterns in vstr — component choices, key bindings, layout conventions verified by audit
metadata:
  type: project
---

## Verified Patterns (as of 2026-06-12)

### Key bindings
- All key matching uses `key.Matches(msg, tui.DefaultKeys.X)` — never raw string comparison.
- `tui.DefaultKeys` in `pkg/tui/keymap.go` is the single source of truth.
- `Quit` binding covers both `ctrl+c` and `esc`.

### FormNavigator usage
- `tui.NewNavigator(elementCount)` where elementCount = number of fields (NOT counting the submit button).
- Navigator allows `FocusIndex == elementCount` as the submit-button state.
- `HandleNavigation(msg.String())` is called directly with the raw key string ("tab", "shift+tab", "up", "down").

### Submit button pattern
- Submit button is rendered at `FocusIndex >= elementCount` (focused) vs `< elementCount` (blurred).
- Enter at `FocusIndex == elementCount` triggers submit.

### Async I/O
- All file I/O (repository calls, path validation) is wrapped in a `tea.Cmd` closure, never called inside `Update` directly.
- Result arrives as a typed `*SaveResultMsg` and is handled in the `switch msg := msg.(type)` block.

### MessageManager
- `pkg/messages.MessageManager` is used for all user-facing feedback.
- `Clear()` is called before each re-validation attempt.
- Messages are cleared on user typing (via `clearMessagesOnInput` / `HasMessages` check).
- `fmt.Println` / `log` are only used in CLI entry points (workspace_commands.go), not inside TUI models.

### WindowSizeMsg
- Handled in both task and workspace models to resize inputs (`Width = msg.Width - 10`).

### Style conventions
- `pkg/styles` package-level vars for all Lipgloss styles — no inline `lipgloss.NewStyle()` calls inside `View()`.
- Focused input: `FocusedInputStyle` applied to `PromptStyle` and `TextStyle`.
- Unfocused input: `UnfocusedInputStyle` (blank style) resets them.

### TaskSelector sub-component (workspace-specific)
- `components.TaskSelector` is used as a mutable sub-component; not a `tea.Model`.
- `taskSelector.Update(msg)` returns `tea.Cmd`, not `(tea.Model, tea.Cmd)`.
- Parent routes messages manually based on `IsInSearchMode()` state.

**Why:** These patterns were verified by reading all files in internal/task/ and internal/workspace/.
**How to apply:** New TUI forms should follow these same conventions. Flag deviations as potential issues.
