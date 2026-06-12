---
name: cli-conventions
description: Established naming, flag, and structural patterns across vstr task/workspace commands
metadata:
  type: project
---

## Subcommand verbs (confirmed in codebase)

Verbs used across both `task` and `workspace` groups: `create`, `list`, `edit`, `delete`, `run`.
`delete` is the canonical delete verb — not `remove` or `rm`.

## Positional argument pattern

All commands targeting a named resource use `<name>` as a single positional arg with `cobra.ExactArgs(1)`:
- `task edit <name>`, `task delete <name>`, `task run <name>`
- `workspace edit <name>`, `workspace delete <name>`, `workspace run <name>`

## Error/success output pattern

All non-TUI commands use `styles.PrintError(...)` and `styles.PrintSuccess(...)` (not `fmt.Println`).
Error messages include the resource name: `"Failed to delete workspace '%s': %v"`.

## Repository delete behavior (known design issue)

Both `repository.DeleteTask` and `repository.DeleteWorkspace` are silent-no-op on a missing name:
they filter the slice and rewrite the file, returning nil even when nothing was removed.
The `delete` commands consequently print "deleted successfully!" for nonexistent names.
This is a confirmed honesty violation per the project's error-handling rules (fail-fast at boundaries,
no misleading messages). As of 2026-06-12, this is a known issue in both `task delete` and
`workspace delete`, not yet fixed.

## Flag patterns

- `task list` has `--only-names` / `-n` (BoolP)
- `task create` has `--file` / `-f` (StringP) for batch import
- No flags on delete/edit/run commands — all operate via positional arg only

## Root-level aliases

`workspace` has alias `w`. No alias on `task` group confirmed.

## `task edit` vs `workspace edit` asymmetry

`task edit <name>` calls `repository.FindTaskByName` first and errors if not found — honest behavior.
`workspace edit <name>` passes name directly to `EditWorkspaceCommand`, which presumably calls
`FindWorkspaceByName` internally (via workspace_create.go) — need to verify if it also fails fast.
