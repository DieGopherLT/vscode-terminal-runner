# Pattern: Literal text vs special keys (-l flag rule)

**Problem**: Sending `tmux send-keys -t SESSION "Down"` typed the four characters D-o-w-n into
the active TUI field instead of moving the cursor to the next field.

**Solution**: Two distinct command forms — never mix in the same call:

```bash
# Literal text: use -l flag
tmux send-keys -t SESSION -l "text to type"

# Special keys: omit -l entirely
tmux send-keys -t SESSION Down
tmux send-keys -t SESSION Enter
tmux send-keys -t SESSION Tab
tmux send-keys -t SESSION Escape
tmux send-keys -t SESSION C-n
tmux send-keys -t SESSION C-b
tmux send-keys -t SESSION BSpace
```

**Why**: The `-l` flag tells tmux to treat every argument as a literal string of characters.
Without `-l`, tmux recognizes key names (Down, Enter, C-n, etc.) as special key symbols and
dispatches the corresponding key event to the terminal.

**Applies when**: Any automated tmux interaction that involves both typing text into fields
AND navigating with arrow/control keys. This is every vstr TUI test.
