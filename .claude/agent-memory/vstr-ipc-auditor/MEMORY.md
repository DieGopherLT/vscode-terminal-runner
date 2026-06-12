# VSTR IPC Auditor Memory

## Key Architectural Facts

- Both `task run` and `workspace run` exclusively use `SecureRunner` (plain `Runner` exists but is never called by commands — dead code path).
- The extension (SecureBridgeServer) DOES implement full auth: writes `auth_token` (64-hex-char, crypto.randomBytes(32)) and `secure: true` to bridge JSON.
- Bridge files are written with explicit `mode: 0o600` via SecureFileManager.writeBridgeInfo — matches CLI's <= 0700 requirement.
- The extension sets the `VSTR` env var via `context.environmentVariableCollection.replace('VSTR', port)` — affects new terminals opened after activation.
- CORS: validateOrigin returns true when `Origin` header is absent (CLI sends no Origin) — CLI requests are allowed through.

## Known Open Issues (as of 2026-06-04)

### Warning

- `IsBridgeOperative` and `validateBridge` (vscode_bridge_discovery.go) send unauthenticated GET /ping — extension requires auth on ALL non-OPTIONS requests. These helpers return false for any secure bridge, causing `ListAvailableBridges` to delete live bridge files via `os.Remove`. Impact: plain discovery path marks all secure-mode bridges as dead and deletes them. However, both `task run` and `workspace run` use `DiscoverSecureBridge` (not the plain `DiscoverBridge` / `ListAvailableBridges`), so the deletion side-effect is not triggered by normal command usage.
- `selectBridge` reads from stdin unconditionally — hangs in non-TTY/script contexts.
- `DiscoverBridge` step 2 (process tree) falls through silently if workspace path doesn't match any bridge file — no warning to user.
- Extension uses `Date.now()` as `instance_id` — two instances started within 1ms would have identical IDs.

### Cosmetic

- `BridgeClient.ExecuteWorkspace` has a double-decode bug (response body already drained on non-200 path before re-decode on success). Plain `Runner` is dead code so this is never triggered in practice.

## Previously Wrong Memory (corrected)

Previous memory claimed: extension has no auth, world-readable files, SecureRunner always fails. All false as of current code.

## File Locations

- CLI bridge discovery: `cli/internal/vscode/vscode_bridge_discovery.go`
- CLI bridge client (plain): `cli/internal/vscode/vscode_bridge_client.go`
- CLI secure runner: `cli/internal/vscode/secure_runner.go`
- CLI secure client: `cli/internal/client/secure_client.go`
- CLI auth manager: `cli/internal/security/auth.go`
- Extension entry: `vscode-extension/src/extension.ts`
- Extension server: `vscode-extension/src/secure-bridge-server.ts`
- Extension auth: `vscode-extension/src/security/auth-manager/index.ts`
- Extension file manager: `vscode-extension/src/security/file-manager/index.ts`
- Extension CORS: `vscode-extension/src/security/cors-manager/index.ts`
- Extension middleware: `vscode-extension/src/security/security-middleware/index.ts`

## Protocol Summary

- Endpoints: GET /ping, POST /task, POST /workspace, GET /security/status
- Payload: JSON. Task fields: name, path, cmds[], icon, iconColor. Workspace: name, tasks[].
- Auth: Bearer token in Authorization header + User-Agent: VSTR-CLI/1.0 (SecureClient only)
- /ping response (secure): {status, version, workspace, port, secure:true, security_features:[]}
- /task response (200): {success:true, message}; error: {success:false, error}
- /workspace response (200): {success:true, results:[{task, success, error?}]}; error: {success:false, error}
- Auth token: crypto.randomBytes(32).toString('hex') = 64 hex chars >= 32 byte minimum

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
