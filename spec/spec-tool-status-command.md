---
title: vstr status command
version: 1.0
date_created: 2026-06-12
last_updated: 2026-06-12
owner: DieGopherLT
tags: [tool, cli, diagnostics, bridge]
---

# Introduction

This specification defines a `vstr status` diagnostic command that reports whether the
VSTR-Bridge VSCode extension is installed and whether the CLI can reach the bridge running
inside VSCode. Its purpose is to turn the most common first-run failure ("I ran a task and
nothing happened") into an actionable diagnosis.

## 1. Purpose & Scope

Provide a single read-only command that checks the two things the whole tool depends on:

1. The VSTR-Bridge extension is installed in VSCode.
2. A secure bridge connection can be established (discovery + authenticated `/ping`).

Out of scope: installing or updating anything, modifying config, running tasks/workspaces, or
listing bridges interactively. `status` only reports.

## 2. Definitions

- **Extension**: the VSTR-Bridge VSCode extension, id `diegopherlt.vstr-bridge`.
- **Bridge**: the local HTTP server the extension runs; discovered by `vscode.DiscoverBridge`.
- **Secure mode**: the bridge `/ping` response field `secure: true` (see `client.TestConnection`).

## 3. Requirements, Constraints & Guidelines

- **REQ-001**: `vstr status` SHALL report whether the extension `diegopherlt.vstr-bridge` is
  installed, using `code --list-extensions` (case-insensitive substring match).
- **REQ-002**: `vstr status` SHALL report whether a secure bridge connection succeeds, by
  running `vscode.DiscoverBridge()` then `client.TestConnection(ctx)` against `GET /ping`.
- **REQ-003**: When the bridge is reachable, the report SHALL include the discovered port and
  workspace name (`BridgeInfo.Port`, `BridgeInfo.WorkspaceName`).
- **REQ-004**: Each check SHALL render a clear OK / NOT-OK line; a failing check SHALL include
  the underlying reason (e.g. "extension not installed", discovery error, auth failure).
- **REQ-005**: `vstr status` SHALL exit zero when both checks pass and non-zero when any check
  fails, so it is usable in scripts.
- **CON-001**: `isExtensionInstalled` (`internal/cfg/config_setup.go:152`) is unexported.
  Export it (e.g. `cfg.IsExtensionInstalled`) or add a thin exported wrapper; do not duplicate
  the `exec.Command("code", "--list-extensions")` logic.
- **CON-002**: The command SHALL NOT print the auth token or any secret.
- **CON-003**: The connection check MUST use a bounded `context` timeout (mirror the 30s used
  in `NewRunner`, `internal/vscode/vscode_runner.go:48`).
- **GUD-001**: Reuse `pkg/styles` printers (`PrintSuccess`, `PrintError`, `PrintInfo`,
  `PrintProgress`) for consistent output; do not introduce a new rendering style.
- **PAT-001**: Register the command in `cmd/root.go::init` via `rootCmd.AddCommand(...)`, the
  same way `cfg.SetupCMD` is registered (`cmd/root.go:47`).

## 4. Interfaces & Data Contracts

### Existing functions to reuse

- `cfg.isExtensionInstalled() bool` (`internal/cfg/config_setup.go:152`) — to be exported.
- `vscode.DiscoverBridge() (*BridgeInfo, error)` (`internal/vscode/vscode_bridge_discovery.go:79`).
- `client.NewClient(port int) *Client` (`internal/client/client.go:25`).
- `(*Client).LoadAuthFromToken(token string) error` (`internal/client/client.go:43`).
- `(*Client).TestConnection(ctx context.Context) error` (`internal/client/client.go:48`).

### BridgeInfo (existing, `internal/vscode/vscode_bridge_discovery.go:20`)

```go
type BridgeInfo struct {
    Port          int
    AuthToken     string
    Secure        bool
    WorkspaceName string
    // ... other discovery fields
}
```

### Connection check shape (to implement, mirrors NewRunner steps 1-4)

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

bridge, err := vscode.DiscoverBridge()        // step 1
// on err: report NOT-OK with err, skip remaining steps
c := client.NewClient(bridge.Port)            // step 2
c.LoadAuthFromToken(bridge.AuthToken)         // step 3
err = c.TestConnection(ctx)                   // step 4 -> OK / NOT-OK
```

### CLI surface

```
vstr status
```

### Example output (illustrative)

```
VSTR status
  [OK]      Extension 'diegopherlt.vstr-bridge' installed
  [OK]      Bridge connection (port 51234, workspace "my-project")
```

```
VSTR status
  [FAIL]    Extension 'diegopherlt.vstr-bridge' not installed  (run 'vstr setup')
  [FAIL]    Bridge connection: VSCode bridge not found
```

## 5. Acceptance Criteria

- **AC-001**: Given the extension is installed and VSCode is running with the bridge active,
  When `vstr status` runs, Then both checks report OK, the port and workspace are shown, and
  the exit code is 0.
- **AC-002**: Given the extension is not installed, When `vstr status` runs, Then the extension
  check reports NOT-OK with a hint to run `vstr setup`, and the exit code is non-zero.
- **AC-003**: Given the extension is installed but no bridge is discoverable (VSCode closed),
  When `vstr status` runs, Then the connection check reports NOT-OK with the discovery error
  and the exit code is non-zero.
- **AC-004**: Given a bridge is discoverable but authentication fails, When `vstr status` runs,
  Then the connection check reports NOT-OK with the auth failure reason.
- **AC-005**: The command SHALL never print the auth token.

## 6. Test Automation Strategy

- **Test Levels**: Unit for the report-assembly logic with injected check results; the raw
  `code --list-extensions` and network calls are environment-dependent and may be left to
  manual/integration verification.
- **Frameworks**: Go standard `testing`.
- **Test Data Management**: Inject fake check outcomes; do not depend on a real VSCode install
  in unit tests.
- **CI/CD Integration**: `go test ./...` must pass.

## 7. Rationale & Context

The bridge is an external, fragile dependency: the extension may be missing, VSCode may be
closed, or the token may be stale. A dedicated diagnostic command is the highest-value
addition for a first impression because it converts an opaque "nothing happened" into a
precise cause. Both checks already have working plumbing (`isExtensionInstalled`,
`DiscoverBridge`, `TestConnection`); `status` only composes and reports them.

## 8. Dependencies & External Integrations

### External Systems

- **EXT-001**: VSCode `code` CLI — required for the extension check (`code --list-extensions`).
- **EXT-002**: VSTR-Bridge extension HTTP server — target of the `/ping` connection check.

### Technology Platform Dependencies

- **PLT-001**: Go 1.24+, `spf13/cobra`.

## 9. Examples & Edge Cases

- `code` CLI not on PATH -> `isExtensionInstalled` returns false; report NOT-OK and note that
  the `code` command was not found if distinguishable.
- Run from a terminal inside VSCode (env `VSTR`/`VSTR_TOKEN` set) -> discovery uses the env
  path and should succeed fast.
- Run from a plain terminal outside VSCode -> discovery falls back to process-tree / `/tmp`
  scan; may legitimately report NOT-OK if no instance is found.

## 10. Validation Criteria

- Both checks render an OK/NOT-OK line with a reason on failure.
- Exit code reflects combined pass/fail.
- No secret is ever printed.

## 11. Related Specifications / Further Reading

- `spec-tool-json-import.md`
- `spec-tool-version-command.md`
- Project `CLAUDE.md` — "Bridge Discovery" section (discovery order and validation).
