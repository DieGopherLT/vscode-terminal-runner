---
name: path-trailing-slash-exact-match
description: Path suggestions do not appear for ~/  or /home/  because contractPath converts the single result to match the input exactly, triggering ShouldShow's exact-match suppression
metadata:
  type: project
---

# Pattern: Path dropdown suppressed by exact-match rule on trailing-slash inputs

**Problem**: Typing "~/" or "/home/" in the Path field shows no dropdown even after the async scan
completes. This looks like a bug but is intentional behavior.

**Solution**: The scan generates exactly one suggestion that, after `contractPath` conversion, equals
the input string. `ShouldShow` suppresses the dropdown when `len(visible)==1 && visible[0]==input`.
To trigger the dropdown, type a partial name after the separator, e.g. "~/D" or "/home/d".

**Why**: With input="~/":
- `expandPath("~/")` returns "/home/diego" (no trailing slash — `filepath.Join` strips it)
- `generatePathSuggestions` uses Dir("/home/diego")="/home" as searchDir, prefix="diego"
- Only "/home/diego" matches → `contractPath` → "~" → add "/" → suggestion is "~/"
- StartsWithFilter("~/", "~/") = true, but ShouldShow sees `visible[0]=="~/"` == input → suppressed

With "/home/" the same happens: scan returns "/home/" (exact match with input) → suppressed.
With "/tmp/" it works because there are multiple subdirectories, none are an exact match.

**Applies when**: Testing the Path field with trailing-slash inputs pointing to the home directory
or any directory that has exactly one matching subdirectory that aliases back to the input.

**Example**:
```bash
# Does NOT show dropdown (exact match suppression)
tmux send-keys -t SESSION -l "~/"
sleep 1.0
# No dropdown visible

# DOES show dropdown (partial prefix filters to multiple results)
tmux send-keys -t SESSION -l "~/D"
sleep 1.0
# Dropdown shows ~/Desktop/, ~/Documents/, ~/Downloads/

# DOES show dropdown outside home (multiple results, no exact match)
tmux send-keys -t SESSION -l "/tmp/"
sleep 1.0
# Dropdown shows /tmp/claude-1000/, /tmp/gitkraken/, etc.
```
