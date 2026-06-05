# internal/vscode

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Handles bridge discovery and authenticated HTTP communication with the VSTR-Bridge VSCode extension. There is a single execution path: `Runner` (this package) orchestrates discovery and delegates I/O to `client.Client` (`internal/client`). All bridge traffic is authenticated; the bridge must run in secure mode.

## Entry Points

- `vscode_bridge_discovery.go::DiscoverBridge` - Layered discovery, every candidate validated for security: VSTR env var -> parent process tree -> directory scan -> user prompt
- `vscode_bridge_discovery.go::discoverFromEnv` - Resolves the bridge from `VSTR`/`VSTR_TOKEN`; prefers the validated bridge file, falls back to the window-scoped `VSTR_TOKEN`
- `vscode_bridge_discovery.go::discoverFromParentProcess` - Walks the process tree to the parent VSCode window, matches a validated bridge by workspace path
- `vscode_bridge_discovery.go::discoverFromScan` - Validates every bridge file; single -> use, multiple -> `selectBridge` prompt
- `vscode_bridge_discovery.go::scanValidBridges` - Returns every bridge file passing `validateBridgeFile`; stale files are left for the extension to clean up (no consumer-side deletion)
- `vscode_bridge_discovery.go::validateBridgeFile` - Validates one file: permissions, structure, `Secure` flag, token length
- `vscode_runner.go::NewRunner` - Discovers bridge, creates `client.Client`, loads auth from the discovered token, tests connection
- `vscode_runner.go::Runner.RunTask` - Looks up task by name, displays info, calls `client.ExecuteTask`
- `vscode_runner.go::Runner.RunWorkspace` - Looks up workspace, calls `client.ExecuteWorkspace`
- `security_errors.go::handleBridgeError` - Pattern-matches error messages, returns user-friendly hints

## Key Files

- **vscode_bridge_discovery.go**: `BridgeInfo` struct; all discovery strategies; per-file validation; interactive selection; process-tree detection
- **vscode_runner.go**: `Runner`; orchestrates discovery + `client.Client` + repository + display
- **security_errors.go**: Error message mapping to user-friendly messages with recovery hints

**Note**: Symbol references use LSP-optimized format (`file::Symbol`) for:

- `goToDefinition`: Jump directly to symbol location
- `findReferences`: Find all real usages (zero false positives)
- `hover`: Get type info and documentation instantly
- `documentSymbol`: Navigate file structure without reading full content

## Business Logic

**Bridge discovery order** (every candidate routed through `validateBridgeFile`, so the result always carries a validated token):

1. `VSTR` env var — the per-window signal the extension injects into its terminals. The validated `bridge-<port>.json` is the source of truth; the window-scoped `VSTR_TOKEN` is a resilient fallback when that file is missing/unreadable.
2. Parent process tree scan — walks up 10 levels looking for `code`/`code-insiders`/`electron`, then matches a validated bridge by workspace path
3. Directory scan of `/tmp/vstr-bridge/bridge-*.json` — validates each; if 0 found: error; if 1: return; if 2+: `selectBridge` prompts via stdin

**Why VSTR matters:** `VSTR`/`VSTR_TOKEN` are window-scoped (set per VSCode extension host), so when running inside an integrated terminal they unambiguously identify the bridge for *that* window — the only precise signal when multiple VSCode windows are open.

**Security validation:** directory permissions <= 0700, per-file permissions via `AuthManager.ValidateFilePermissions`, token length >= 32 bytes, `Secure: true` flag.

**Execution flow:**

1. `DiscoverBridge` -> `NewRunner` creates `client.Client`, loads auth via `LoadAuthFromToken(bridgeInfo.AuthToken)`, runs `TestConnection`
2. `RunTask(name)` -> `repository.FindTaskByName` -> display info -> `client.ExecuteTask`
3. Client converts `Task` to a payload map and POSTs to `http://localhost:<port>/task`
4. Non-200 -> parse error JSON -> `handleBridgeError`

**Workspace differs from task:** payload includes workspace name + array of task payloads; response includes per-task results; any `success: false` result aggregates into the returned error.

**`client.Client`** adds `Authorization: Bearer <token>` and `User-Agent: VSTR-CLI/1.0` headers. Uses context cancellation (RunTask: 60s, RunWorkspace: 120s).

## Dependencies

**Internal:**

- `internal/models`: `Task`, `Workspace` structs
- `internal/repository`: `FindTaskByName`, `FindWorkspaceByName`
- `internal/security`: `AuthManager` — file permission validation, `LoadTokenFromBridge`/`LoadTokenFromString`, auth header generation
- `internal/client`: `Client` — authenticated HTTP; `taskToPayload` lives here
- `pkg/styles`: `PrintInfo`, `PrintError`, `PrintSuccess`, `RunnerTaskNameStyle`

**External:**

- `github.com/shirou/gopsutil/v3/process`: parent process tree traversal
- `github.com/samber/lo`: `lo.Find` for bridge matching

**Environment Variables:**

- `VSTR`: Port of the active bridge (set by VSTR-Bridge extension in its terminals)
- `VSTR_TOKEN`: Auth token of the active bridge (also window-scoped); resilient fallback for auth when the bridge file is unavailable
- `TMPDIR`/`TEMP`/`TMP`: System temp dir used to locate `/tmp/vstr-bridge` (or Windows equivalent)

## Architecture

- **Single authenticated path**: `Runner` (orchestration, this package) + `client.Client` (HTTP, `internal/client`). Discovery and transport are decoupled.
- **BridgeInfo as config carrier**: The JSON file written by the extension carries port, PID, InstanceID, workspace path/name, auth token, and secure flag. `discoverFromEnv`'s fallback synthesizes a `BridgeInfo` from `VSTR`/`VSTR_TOKEN` directly.
- **Token reuse**: discovery always yields a validated `AuthToken` (from file or `VSTR_TOKEN`), so `NewRunner` authenticates via `LoadAuthFromToken` without re-reading `/tmp`.
- **Local HTTP only**: no TLS; relies on loopback binding + filesystem permissions + auth token. Acceptable for localhost-only communication.

## Modification Guide

### Adding Features

- **New discovery strategy**: add `func discoverFrom<Method>(...) (*BridgeInfo, error)` and call it in `DiscoverBridge` at the right priority; return early on success, fall through on error. Route candidates through `validateBridgeFile`.
- **New execution endpoint**: add a method to `client.Client` following `ExecuteTask`; POST to the new `/path`; parse the response
- **New error type from bridge**: add a pattern match in `handleBridgeError`; provide a user-friendly message + `styles.PrintInfo` hint

### Removing Code

- Removing a discovery strategy: ensure remaining strategies still cover all user scenarios (especially the `VSTR` env path for multi-window correctness)

### Common Pitfalls

- `DiscoverBridge` not finding a running bridge -> check `VSTR` env var, verify `/tmp/vstr-bridge/bridge-*.json` files exist, check extension is running in secure mode
- "invalid auth token length" -> bridge file token < 32 bytes (or `VSTR_TOKEN` too short) or file permissions > 0700; check extension secure mode settings
- `selectBridge` hangs -> stdin not a TTY (script context); set `VSTR` env var to bypass interactive selection
- Process tree scan misses VSCode -> process renamed or reparented (tmux/SSH/nohup); use `VSTR` env var as fallback

## Usage Examples

```go
// Run a single task
runner, err := vscode.NewRunner()
if err != nil {
    return err
}
return runner.RunTask("build")

// Run a whole workspace
runner, err := vscode.NewRunner()
if err != nil {
    return err
}
return runner.RunWorkspace("dev-setup")

// Check bridge availability
info, err := vscode.DiscoverBridge()
if err != nil {
    return fmt.Errorf("VSCode not available: %w", err)
}
fmt.Printf("Bridge on port %d, workspace: %s\n", info.Port, info.WorkspaceName)
```

---

## Claude's Navigation Commitment

This CLAUDE.md is my map for navigating this module. I commit to:

- **Update immediately** after any code modification in this module
- **Verify accuracy** of all symbol references after each change
- **Maintain truth** - outdated documentation is a critical bug
- **Treat this as my compass** - if this map is wrong, I'm lost

Last verified: 2026-06-04
