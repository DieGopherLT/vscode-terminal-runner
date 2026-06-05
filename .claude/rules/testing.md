---
paths: ["**/*_test.go"]
---

# Testing

## How to Run Tests
- All tests: `go test ./...`
- Single package: `go test ./internal/repository/...`
- Single test: `go test ./internal/client/ -run TestClient_Ping`
- Coverage (per package, with threshold check by eye against 0.8): `go test -cover ./...`
- Coverage profile + report: `go test -coverprofile=cover.out ./... && go tool cover -func=cover.out`
- No Makefile and no watch mode in this repo; invoke `go test` directly.

## Test Organization
- Co-located with production code: `_test.go` lives beside the file under test in the same package directory.
- File naming: `<source>_test.go`. When one source file warrants many tests, split by concern (e.g. `repository_tasks_test.go`, `repository_workspaces_test.go`).
- Structure: table-driven subtests with `t.Run(name, ...)` over a `[]struct{ name string; ... }` slice. Prefer one table per behavior cluster.
- White-box tests (same `package foo`, not `foo_test`) when a test must reach unexported seams — used in `internal/repository` (save-file redirect seams) and `internal/client` (`clientForServer`). Black-box otherwise.

## Test Utilities
Shared helpers live in `pkg/testutils` (import `github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils`). Reuse before creating: check this package first; only add a new helper when no existing one fits, and extend the matching file (`builders.go`, `bridge_server.go`, `helpers.go`) rather than introducing a parallel one.

Available:
- `TaskBuilder` (builder) — `testutils.NewTask().WithName("build").WithCmds("make").Build()`
- `WorkspaceBuilder` (builder) — `testutils.NewWorkspace().WithName("my-project").WithTasks(task1, task2).Build()`
- `NewBridgeTestServer` (fake bridge) — `server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{})`. Override one handler for error paths:
  ```go
  server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
      PingHandler: func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(http.StatusUnauthorized)
      },
  })
  ```
- `ValidTestToken` (const) — 32-byte token satisfying `security.MinTokenLength`; use across all auth-path tests: `c.LoadAuthFromToken(testutils.ValidTestToken)`.
- `CreateTestJSONFile` / `CreateTempFileWithPermissions` / `CreateTempDirWithPermissions` (helpers) — e.g. `testutils.CreateTempFileWithPermissions(0600)` for security/permission tests.

White-box seams (NOT in `pkg/testutils` — they need package-private access, kept in their own `_test.go`):
- `clientForServer(t, server)` (package `client`) — extracts the port from `httptest.Server.URL` and calls `NewClient(port)`; avoids an import cycle.
- `redirectTasksSaveFile(t)` / `redirectWorkspacesSaveFile(t)` (package `repository`) — `defer redirectTasksSaveFile(t)()` redirects persistence to a temp file. Consume these; do not recreate.

## Coverage
- Target threshold: 0.8 per module under test.
- Report command: `go test -coverprofile=cover.out ./... && go tool cover -func=cover.out` (add `-html=cover.out` for a browser view).
- Achieved: `internal/repository` 0.825, `internal/client` 0.933.
- Exclusions: generated code, plain DTO/model structs with no logic (`internal/models`), and `main`/Cobra command bootstrapping (`cmd/`) — wiring, not behavior.

## Patterns Established
- Builder pattern for domain fixtures (`TaskBuilder`, `WorkspaceBuilder`) with fluent `With*` setters and a terminal `Build()`.
- Fake HTTP bridge via `httptest.Server` behind `NewBridgeTestServer` + `BridgeHandlerConfig`, with per-handler overrides for error-path coverage.
- Seam injection for filesystem persistence: `defer redirect*SaveFile(t)()` points writes at a temp file so repository tests stay hermetic.
- Shared 32-byte `ValidTestToken` constant so every auth-path test uses one token that satisfies `security.MinTokenLength`.
- Table-driven subtests as the default shape.
- Bug-documenting tests: a known production bug is captured as a skipped test named with a `BUG:` prefix (`TestSaveFromFile_BUG_shouldSucceedOnFreshSystem`, `t.Skip`) so the expected-correct behavior is recorded without breaking the suite.

## What to Test
- Repository persistence round-trips: save then load, empty-state, and overwrite paths.
- Client request/response behavior against the fake bridge: success, auth failure, and non-2xx handling.
- Auth loading and token validation through `LoadAuthFromToken` with `ValidTestToken`.
- Error branches and boundary inputs at package boundaries (empty files, missing files, bad permissions).

## What NOT to Test
- Plain DTO/model structs in `internal/models` with no behavior.
- `cmd/` Cobra wiring and `main` bootstrapping.
- Generated code.
- Third-party libraries (Cobra, Bubbletea, lipgloss) — test our usage, not their internals.

## Known Bug (do not "fix" by changing the test)
`repository.SaveFromFile` always fails on a fresh install: it reads the destination `tasks.json` and returns "Provided file is empty" when that file is empty, but `ensureTasksSaveFile()` creates an empty `tasks.json` on a fresh system — so import can never succeed on a new machine. The message also names the wrong file (the destination, not the caller-supplied batch path). Captured by the skipped `TestSaveFromFile_BUG_shouldSucceedOnFreshSystem` in `internal/repository/repository_tasks_test.go`. Intended behavior: a valid batch file should import into an empty/non-existent `tasks.json`, treating zero bytes as zero existing tasks.
