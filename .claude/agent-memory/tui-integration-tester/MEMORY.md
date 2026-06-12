# TUI Integration Tester — Pattern Memory

Patterns recorded from automated interaction sessions with vstr TUI forms.
Each entry links to a dedicated file with problem, solution, root cause, and example.

## Key dispatch

- [Literal text vs special keys (-l flag rule)](pattern-tmux-literal-vs-special.md) — `-l` flag sends literal chars; omit it for special keys like Down, Enter, Tab, C-n

## Form behavior

- [Escape quits the entire form](pattern-escape-global-quit.md) — Escape is global quit, NOT a "close dropdown" shortcut
- [Suggestions workflow: Ctrl+N/B + Tab](pattern-suggestions-workflow.md) — how to open, navigate, and apply the suggestion dropdown without closing the form

## Screen capture

- [ANSI capture required for selection state](pattern-ansi-for-selection-state.md) — plain capture-pane loses highlights; use -e flag to see which dropdown item is selected
