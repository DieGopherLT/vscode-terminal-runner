# Pattern: Suggestions workflow — Ctrl+N/B + Tab

**Problem**: The suggestion dropdown for Icon and Icon Color fields was not obvious to drive
automatically. Initial attempts either did not open the dropdown, typed key names literally,
or closed the entire form.

**Solution**: Three-phase workflow:

```bash
# 1. Open dropdown and place cursor on first item
tmux send-keys -t SESSION C-n
sleep 0.3

# 2. Navigate to the target item
#    Each C-n advances one position; each C-b goes back one
tmux send-keys -t SESSION C-n   # item 2
sleep 0.2
tmux send-keys -t SESSION C-n   # item 3
sleep 0.2
tmux send-keys -t SESSION C-b   # back to item 2
sleep 0.2

# 3. Apply selected item — fills the field AND closes the dropdown
tmux send-keys -t SESSION Tab
sleep 0.3

# 4. Continue to next field
tmux send-keys -t SESSION Down
sleep 0.2
```

**Why**: The first `C-n` both opens the list AND selects item 1. Subsequent `C-n` calls advance
the selection. `Tab` commits the current selection to the field text and dismisses the dropdown.
This is the only way to close the dropdown while staying in the form.

**Applies when**: Any field that shows "ctrl+b/n suggestions" in the help line at the bottom
of the vstr form. Currently: Icon and Icon Color fields on both task create and workspace create.

**Counting tip**: If you need item N, send exactly N `C-n` presses (first opens + selects item 1,
each subsequent advances by one). Use ANSI capture to verify the selection before applying if
the exact item matters.
