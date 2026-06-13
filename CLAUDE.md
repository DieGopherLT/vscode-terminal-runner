# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Information

**Binary name:** `vstr`

```shell
# Build
go build -o bin/vstr

# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/tui/...

# Run a single test
go test ./pkg/tui/ -run TestNavigator_HandleNavigation
```

## Architecture Overview

The CLI communicates with the **VSTR-Bridge** VSCode extension via a local HTTP server. The extension creates terminals inside VSCode; the CLI tells it what to run.

```
vstr (CLI) --> Runner --> client.Client --> HTTP --> VSTR-Bridge Extension --> VSCode terminals
```

### Command Structure

`cmd/` registers Cobra commands and delegates to `internal/` packages:

- `vstr setup` -> `internal/cfg` (setup wizard, installs VSCode extension)
- `vstr task <sub>` -> `internal/task` (TUI forms for CRUD + run)
- `vstr workspace <sub>` -> `internal/workspace` (TUI forms for CRUD + run)

### Package Responsibilities

| Package              | Responsibility                                                     |
| -------------------- | ------------------------------------------------------------------ |
| `internal/models`    | Data types: `Task`, `Workspace`, `Config` (each impl `GetName()`)  |
| `internal/repository`| JSON persistence via generic `JSONRepository[T NamedEntity]`        |
| `internal/cfg`       | App config file, setup wizard, extension install                   |
| `internal/vscode`    | Bridge discovery and `Runner` orchestration                        |
| `internal/client`    | `Client` — auth-aware HTTP client for bridge                       |
| `internal/security`  | Auth token management and file permission validation               |
| `internal/task`      | Bubbletea TUI model for task form (create/edit/list/delete/run)    |
| `internal/workspace` | Bubbletea TUI model for workspace form                             |
| `pkg/tui`            | Reusable `FormNavigator` and `suggestions.Manager` for TUI forms   |
| `pkg/messages`       | `MessageManager` — collects error/success messages shown in TUI    |
| `pkg/styles`         | Lipgloss styles, VSCode icon list, ANSI color list                 |
| `pkg/collections`    | Generic `Set[T comparable]` (used by the workspace task selector)  |
| `pkg/testutils`      | Shared test helpers                                                |

### Data Persistence

Tasks and workspaces are stored as JSON in the user config directory:

- `$XDG_CONFIG_HOME/vscode-terminal-runner/tasks.json`
- `$XDG_CONFIG_HOME/vscode-terminal-runner/workspaces.json`
- `$XDG_CONFIG_HOME/vscode-terminal-runner/config.json`

Both `tasks.json` and `workspaces.json` are persisted by a single generic
`JSONRepository[T NamedEntity]` (`internal/repository/repository.go`). The public
funcs (`ReadTasks`, `SaveWorkspace`, …) are thin adapters over package-level
`taskRepo`/`workspaceRepo` singletons wired in `init()`. The save-file path is
resolved lazily (`getSaveFile func() string`) so tests can redirect it after
`init()`. The Task-allows-duplicates / Workspace-rejects-duplicates asymmetry is
the injected `onAppend` strategy. Set `XDG_CONFIG_HOME` to isolate the store.

### Bridge Discovery

When a task or workspace is run, `vscode.DiscoverBridge()` resolves the target VSCode instance in this order:

1. `VSTR`/`VSTR_TOKEN` env vars (set by the extension in its own terminals; window-scoped)
2. Parent process tree scan for a VSCode process
3. Scan `/tmp/vstr-bridge/bridge-*.json` files written by the extension
4. If multiple bridges found, prompt user to select one

The env path trusts the window-scoped `VSTR_TOKEN` directly (the extension injected it into this terminal); the process-tree and directory-scan paths validate each candidate file (permissions, `Secure: true` flag, auth token >= 32 bytes). Discovery yields a token that `NewRunner` reuses to authenticate without re-reading the bridge file.

### Environment Variables

| Variable             | Purpose                                                                 |
| -------------------- | ----------------------------------------------------------------------- |
| `VSTR`               | Port of the active bridge (set automatically by the VSCode extension)   |
| `VSTR_EXTENSION_NAME`| Override extension ID used during `vstr setup` installation             |

---

## TUI Changes

After any change to a file under `internal/task/`, `internal/workspace/`, or `pkg/tui/`,
invoke the `tui-integration-tester` agent to run an end-to-end flow through the affected
form before reporting the task as done. This verifies that navigation, suggestion dropdowns,
and form submission still work correctly in the actual terminal — type checking alone does
not catch Bubbletea UX regressions.

---

## Claude's Navigation Commitment

This CLAUDE.md is my map for navigating this module. I commit to:

- **Update immediately** after any code modification in this module
- **Verify accuracy** of all symbol references after each change
- **Maintain truth** - outdated documentation is a critical bug
- **Treat this as my compass** - if this map is wrong, I'm lost

Last verified: 2026-06-13
