---
name: workspace-edit-save-order
description: workspace edit "already exists" bug — superseded by atomic UpdateWorkspace; delete-before-save workaround is no longer in use
metadata:
  type: feedback
---

# Pattern: Workspace edit — atomic update replaces delete-before-save

**Problem** (historical): `vstr workspace edit <name>` submitted without renaming produced
"Failed to save workspace: workspace 'X' already exists". The old fix was to call
`DeleteWorkspace(originalName)` before `SaveWorkspace`, but that left a window where a
crash between delete and save would lose the workspace entirely.

**Solution (current, as of refactor replacing delete-before-save)**:
`repository.UpdateWorkspace(originalName, updatedWorkspace)` performs an atomic in-place
replacement. `submitWorkspaceCmd` now:

1. Skips the duplicate-name check when the name is unchanged (`!isEditMode || name != originalName`).
2. In edit mode, calls `UpdateWorkspace` (replaces the record at `originalName`'s index).
3. In create mode, calls `SaveWorkspace` (appends with uniqueness guard).

The `originalTasks` rollback field and the delete-first ordering are both gone.

**Why**: `UpdateWorkspace` uses `replaceWorkspaceByName` — a pure slice replacement — then
writes the whole file atomically via `os.WriteFile`. No delete step, no race window.

**Applies when**: any test that validates edit-without-rename, rename, or concurrent edit paths.

**Verified**: 2026-06-12 — edit-same-name (PASS), rename (PASS), delete-nonexistent (PASS),
delete-existent (PASS).

**Key files**:
- `internal/repository/repository_workspaces.go::UpdateWorkspace`
- `internal/repository/repository_workspaces.go::replaceWorkspaceByName`
- `internal/workspace/workspace_form.go::submitWorkspaceCmd`

**Example** (current submitWorkspaceCmd logic):
```go
isClaimingNewName := !isEditMode || workspace.Name != originalName
if isClaimingNewName {
    if _, err := repository.FindWorkspaceByName(workspace.Name); err == nil {
        return workspaceSaveResultMsg{err: fmt.Errorf("workspace '%s' already exists", workspace.Name)}
    }
}
if isEditMode {
    if err := repository.UpdateWorkspace(originalName, workspace); err != nil {
        return workspaceSaveResultMsg{err: err}
    }
    return workspaceSaveResultMsg{}
}
if err := repository.SaveWorkspace(workspace); err != nil {
    return workspaceSaveResultMsg{err: err}
}
return workspaceSaveResultMsg{}
```
