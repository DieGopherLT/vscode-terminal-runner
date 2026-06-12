# Pattern: ANSI capture required for selection state

**Problem**: After navigating a suggestion dropdown with `C-n` several times, plain
`tmux capture-pane -t SESSION -p` output showed all dropdown items without any visual
distinction between them. There was no way to tell which item was currently selected.

**Solution**: Use the `-e` flag to capture ANSI escape codes:

```bash
tmux capture-pane -t SESSION -p -e
```

Then look for the VSCode blue highlight applied to the selected item:

```
[1m[38;2;0;121;204m• activate-breakpoints[0m
```

The sequence `[1m[38;2;0;121;204m` (bold + RGB 0,121,204) is the VSCode blue used by vstr's
Lipgloss styles to mark the active/focused item.

**Why**: `capture-pane -p` strips all ANSI escape sequences, producing clean text but losing
all color and highlighting information. The selection cursor in Bubbletea list components is
communicated purely through ANSI styling — there is no plain-text equivalent.

**Applies when**: Any test that needs to assert which dropdown item is currently highlighted
before pressing `Tab` to apply it. For verifying the final field value after applying, plain
capture is sufficient.

**Example**:

```bash
# Check which item is selected before applying
ANSI=$(tmux capture-pane -t SESSION -p -e)
echo "$ANSI" | grep "38;2;0;121;204m"  # shows the highlighted line

# Verify final field value after Tab — plain capture is enough
PLAIN=$(tmux capture-pane -t SESSION -p)
echo "$PLAIN" | grep "Icon:"
```
