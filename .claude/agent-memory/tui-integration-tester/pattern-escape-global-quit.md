# Pattern: Escape quits the entire form

**Problem**: After opening a suggestion dropdown with `C-n`, pressing `Escape` expecting it to
close only the dropdown ended up exiting the entire TUI form and returning to the shell.

**Solution**: Never press `Escape` mid-test unless the intent is to abort the form entirely.
To dismiss a suggestion dropdown, use `Tab` (apply + close) or navigate to a different field
with `Down`. There is no "close dropdown without applying" shortcut in vstr TUI forms.

**Why**: vstr's Bubbletea models bind `Escape` as a global quit action on the top-level `Update`
function. The dropdown does not intercept it. The key event propagates to the form root, which
calls `tea.Quit`.

**Applies when**: Any test flow that opens a suggestion dropdown and then wants to continue
filling other fields. Use `Tab` to accept the current selection, or `Down` to move away
(which implicitly closes the dropdown by defocusing the field).

**Example**:

```bash
# Wrong: kills the form
tmux send-keys -t SESSION C-n   # open dropdown
tmux send-keys -t SESSION Escape  # QUITS the entire form

# Correct: apply and continue
tmux send-keys -t SESSION C-n   # open dropdown
tmux send-keys -t SESSION C-n   # advance to desired item
tmux send-keys -t SESSION Tab   # apply + close dropdown
tmux send-keys -t SESSION Down  # move to next field
```
