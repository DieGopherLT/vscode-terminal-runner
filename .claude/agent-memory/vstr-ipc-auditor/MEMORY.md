# VSTR IPC Auditor Memory

## Key Architectural Facts

- Both `task run` and `workspace run` exclusively use `SecureRunner` (plain `Runner` exists but is never called by commands).
- The extension (vscode-extension/src/extension.ts) does NOT implement auth. It has no `auth_token` field in its `BridgeInfo` interface and no `secure` flag. The CLI's secure mode assumes the extension writes these fields but the actual extension does not.
- The extension writes bridge files with `fs.writeFileSync` using no explicit permissions — files inherit the process umask (typically 0644, world-readable). This is a security gap.
- The extension's `BridgeInfo` interface omits `auth_token` and `secure` fields that the CLI requires for `SecureRunner`.
- The extension does NOT set `VSTR` env var in the standard way — it uses `vscode.ExtensionContext.environmentVariableCollection` which only affects terminals opened AFTER the extension activates.

## Known Audit Findings

### Critical

- Extension never writes `auth_token` or `secure:true` to bridge JSON — SecureRunner will always fail in production unless a patched extension is in use.
- Bridge JSON written with no explicit permissions (world-readable by default umask 0644), but CLI requires <= 0700.
- `BridgeClient.ExecuteWorkspace` reads the response body after it has already been partially read by `handleResponse` path — double-decode bug (body already drained on error path).

### Warning

- `IsBridgeOperative` and `validateBridge` create a new `http.Client` per call — minor, but unnecessary.
- `selectBridge` reads from stdin unconditionally — hangs in non-TTY/script contexts.
- `DiscoverBridge` step 2 (process tree) falls through silently if workspace path doesn't match any bridge file — no warning to user.
- `ListAvailableBridges` calls `os.Remove` on stale files — side effect during a read operation, may surprise callers.
- Extension uses `Date.now()` as `instance_id` — two instances started within 1ms would have identical IDs.

## File Locations

- CLI bridge discovery: `/home/diego/Documents/projects/vscode-terminal-runner/cli/internal/vscode/vscode_bridge_discovery.go`
- CLI bridge client (plain): `/home/diego/Documents/projects/vscode-terminal-runner/cli/internal/vscode/vscode_bridge_client.go`
- CLI secure runner: `/home/diego/Documents/projects/vscode-terminal-runner/cli/internal/vscode/secure_runner.go`
- CLI secure client: `/home/diego/Documents/projects/vscode-terminal-runner/cli/internal/client/secure_client.go`
- CLI auth manager: `/home/diego/Documents/projects/vscode-terminal-runner/cli/internal/security/auth.go`
- VSCode extension: `/home/diego/Documents/projects/vscode-terminal-runner/vscode-extension/src/extension.ts`

## Protocol Summary

- Endpoints: GET /ping, POST /task, POST /workspace
- Payload: JSON. Task fields: name, path, cmds[], icon, iconColor. Workspace: name, tasks[].
- Auth: Bearer token in Authorization header + User-Agent: VSTR-CLI/1.0 (SecureClient only)
- /ping response: {status, version, workspace, port} — plain; {status, secure, security_features} — secure variant
- /task response: {success, message} or {success:false, error}
- /workspace response: {success, results:[{task, success, error?}]} or {success:false, error}

## Timeout Map

| Location | Timeout |
|----------|---------|
| IsBridgeOperative GET /ping | 1s |
| validateBridge GET /ping | 2s |
| BridgeClient (plain) all requests | 10s |
| SecureClient http.Client | 30s |
| NewSecureRunner overall context | 30s |
| SecureRunner.RunTask context | 60s |
| SecureRunner.RunWorkspace context | 120s |

## Links to Detail Files

- See `ipc-protocol.md` for full protocol documentation
- See `audit-findings.md` for detailed audit findings with line numbers
