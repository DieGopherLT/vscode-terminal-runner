# internal/vscode

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Handles bridge discovery and HTTP communication with the VSTR-Bridge VSCode extension. Provides two execution paths: plain (`Runner`/`BridgeClient`) and authenticated (`SecureRunner`/`SecureClient`).

## Entry Points

- `vscode_bridge_discovery.go::DiscoverBridge` - Layered bridge discovery: env var -> process tree -> /tmp scan -> user prompt
- `vscode_bridge_discovery.go::DiscoverSecureBridge` - Secure bridge discovery with file permission validation and token length check
- `vscode_bridge_discovery.go::ListAvailableBridges` - Scans `/tmp/vstr-bridge/*.json`, pings each, returns active bridges
- `vscode_bridge_discovery.go::IsBridgeOperative` - Single HTTP GET `/ping` with 1s timeout
- `vscode_runner.go::NewRunner` - Creates plain Runner: discovers bridge, creates BridgeClient, pings
- `vscode_runner.go::Runner.RunTask` - Looks up task by name, displays info, calls `client.ExecuteTask`
- `vscode_runner.go::Runner.RunWorkspace` - Looks up workspace, calls `client.ExecuteWorkspace`
- `vscode_bridge_client.go::NewBridgeClient` - Plain HTTP client constructor given port
- `vscode_bridge_client.go::BridgeClient.ExecuteTask` - POST JSON to `/task` endpoint
- `vscode_bridge_client.go::BridgeClient.ExecuteWorkspace` - POST JSON to `/workspace`; aggregates per-task failures
- `secure_runner.go::NewSecureRunner` - Creates SecureRunner: discovers secure bridge, loads token, tests connection
- `secure_runner.go::SecureRunner.RunTask` - Same flow as `Runner.RunTask` but uses `SecureClient`
- `secure_runner.go::SecureRunner.RunWorkspace` - Same flow as `Runner.RunWorkspace` but uses `SecureClient`
- `security_errors.go::handleSecureError` - Pattern-matches error messages, returns user-friendly hints

## Key Files

- **vscode_bridge_discovery.go**: `BridgeInfo` struct; all discovery strategies; bridge validation; interactive selection
- **vscode_bridge_client.go**: `BridgeClient`; plain HTTP POST; `taskToPayload`/`tasksToPayload` converters
- **vscode_runner.go**: `Runner`; orchestrates discovery + client + repository + display for plain mode
- **secure_runner.go**: `SecureRunner`; orchestrates discovery + `SecureClient` + auth for secure mode
- **security_errors.go**: Error code mapping to user-friendly messages with recovery hints

**Note**: Symbol references use LSP-optimized format (`file::Symbol`) for:

- `goToDefinition`: Jump directly to symbol location
- `findReferences`: Find all real usages (zero false positives)
- `hover`: Get type info and documentation instantly
- `documentSymbol`: Navigate file structure without reading full content

## Business Logic

**Bridge discovery order (plain):**

1. `VSTR` env var — fastest; extension sets this in its spawned terminals
2. Parent process tree scan — walks up 10 levels looking for `code`/`code-insiders`/`electron`
3. `/tmp/vstr-bridge/*.json` scan — finds all active instances; pings each; removes stale files
4. If 0 found: error; if 1: return; if 2+: `selectBridge` prompts user via stdin

**Secure discovery extras:** validates directory permissions <= 0700, per-file permissions via `AuthManager.ValidateFilePermissions`, token length >= 32 bytes, `Secure: true` flag. Returns bridge with highest `InstanceID`.

**Execution flow (both modes):**

1. Discover bridge -> create client (plain or secure)
2. `RunTask(name)` -> `repository.FindTaskByName` -> display info -> `client.ExecuteTask`
3. Client converts `Task` to `map[string]interface{}` payload via `taskToPayload`
4. POST to `http://localhost:<port>/task` with `Content-Type: application/json`
5. Decode response; non-200 -> parse error JSON -> `handleSecureError` (secure path only)

**Workspace differs from task:** payload includes workspace name + array of task payloads; response includes per-task results; any `success: false` result aggregates into returned error.

**SecureClient** adds `Authorization: Bearer <token>` and `User-Agent: VSTR-CLI/1.0` headers. Uses `context.WithContext` for cancellation (RunTask: 60s, RunWorkspace: 120s).

## Dependencies

**Internal:**

- `internal/models`: `Task`, `Workspace` structs
- `internal/repository`: `FindTaskByName`, `FindWorkspaceByName`
- `internal/security`: `AuthManager` — file permission validation, auth header generation
- `internal/client`: `SecureClient` — authenticated HTTP
- `pkg/styles`: `PrintInfo`, `PrintWarning`, `PrintError`, `PrintSuccess`, `RunnerTaskNameStyle`

**External:**

- `github.com/shirou/gopsutil/v3/process`: parent process tree traversal
- `github.com/samber/lo`: `lo.Find` for bridge matching

**Environment Variables:**

- `VSTR`: Port of the active bridge (set by VSTR-Bridge extension in its terminals)
- `TMPDIR`/`TEMP`/`TMP`: System temp dir used to locate `/tmp/vstr-bridge` (or Windows equivalent)

## Architecture

- **Two symmetric paths**: `Runner+BridgeClient` (plain) and `SecureRunner+SecureClient` (auth). Both expose identical `RunTask`/`RunWorkspace` APIs — callers switch with one line.
- **BridgeInfo as config carrier**: The JSON file written by the extension carries port, PID, InstanceID, workspace path/name, auth token, and secure flag. No separate config files.
- **Payload decoupling**: `taskToPayload` converts `models.Task` to `map[string]interface{}` to avoid tight coupling with bridge API schema.
- **Local HTTP only**: no TLS; relies on loopback binding + filesystem permissions + auth token. Acceptable for localhost-only communication.

## Modification Guide

### Adding Features

- **New discovery strategy**: add `func discover<Method>() (*BridgeInfo, error)` and call it in `DiscoverBridge` at appropriate priority; return early on success, fall through on error
- **New execution endpoint**: add method to `BridgeClient` and `SecureClient` following `ExecuteTask` pattern; POST to new `/path`; parse response
- **New error type from bridge**: add pattern match in `handleSecureError`; provide user-friendly message + `styles.PrintInfo` hint

### Removing Code

- Removing a discovery strategy: ensure remaining strategies still cover all user scenarios
- Removing `BridgeClient`: also update `vscode_runner.go` and any callers; verify `SecureClient` path still complete

### Common Pitfalls

- `DiscoverBridge` not finding a running bridge -> check `VSTR` env var, verify `/tmp/vstr-bridge/*.json` files exist, check extension is running
- `SecureRunner` fails with "invalid auth token length" -> bridge file has token < 32 bytes or file permissions > 0700; check extension secure mode settings
- `selectBridge` hangs -> stdin not a TTY (script context); set `VSTR` env var to bypass interactive selection
- Process tree scan misses VSCode -> process renamed or reparented (tmux/SSH/nohup); use `VSTR` env var as fallback

## Usage Examples

```go
// Plain mode
runner, err := vscode.NewRunner()
if err != nil {
    return err
}
return runner.RunTask("build")

// Secure mode
runner, err := vscode.NewSecureRunner()
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

Last verified: 2026-02-18
