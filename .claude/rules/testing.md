---
paths: ["**/*_test.go"]
---

# Testing

## How to Run Tests
- All tests: `go test ./...`
- Single package: `go test ./internal/repository/...` (replace path for any package)
- Single test: `go test ./internal/client/ -run TestClient_Ping`
- Coverage (quick summary): `go test -cover ./...`
- Coverage profile + report: `go test -coverprofile=cover.out ./... && go tool cover -func=cover.out`
- Verbose: add `-v`. No watch mode; re-run the command on change.

## Test Organization
- Co-located with production code: `_test.go` lives beside the file under test in the same package directory.
- File naming: `<source>_test.go`. When one source file warrants many tests, split by concern (e.g. `repository_tasks_test.go`, `repository_workspaces_test.go`). Seam fakes/builders for a package go in `fakes_test.go`.
- Structure: table-driven subtests with `t.Run(name, ...)` over a `[]struct{ name string; ...; wantErr string }` slice. Prefer one table per behavior cluster; assert on `wantErr` substrings via `strings.Contains`.
- White-box tests (same `package foo`, not `foo_test`) when a test must reach unexported seams — used in `internal/repository` (save-file redirect seams), `internal/client` (`clientForServer`), and `internal/vscode` (discovery helpers, validation). Black-box otherwise.

## Test Utilities

Reuse-before-create: check both locations before writing a new builder or fake.

### `pkg/testutils` — cross-cutting helpers (import `github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils`)
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

### `internal/vscode` (`fakes_test.go`, `package vscode`) — seam fakes that must stay in-package to avoid import cycles
- `fakeBridgeClient` (fake) — satisfies `BridgeClient`. Inject via `NewRunnerWithDeps(RunnerDeps{Client: client, ...})`; assert on `client.lastTask` after `RunTask()`. Drive errors with `executeTaskErr`.
- `fakeTaskRepository` (fake) — satisfies `TaskRepository`. `&fakeTaskRepository{task: &models.Task{...}}` or `{returnErr: errors.New("not found")}` for the error path.
- `fakeWorkspaceRepository` (fake) — satisfies `WorkspaceRepository`. `&fakeWorkspaceRepository{workspace: &models.Workspace{...}}`.
- `fakeProcessInspector` + `fakeProcessNode` + `buildProcessChain` (fakes) — satisfy `ProcessInspector`/`ProcessNode`. Build a tree with `buildProcessChain(fakeNodeSpec{pid:100,name:"bash"}, fakeNodeSpec{pid:200,name:"code",cmdline:"--folder-uri file:///workspace/myapp"})`, then `detectParentVSCode(&fakeProcessInspector{ppid:100, tree:tree})`.
- `bridgeInfoBuilder` via `newValidBridgeInfo()` (builder) — produces a `BridgeInfo` that passes `validateBridgeStructure`; override one field per rejection test: `.WithPort(0)`, `.WithShortToken()`, `.WithSecure(false)`.
- `minValidToken()` (helper) — returns a token of exactly `security.MinTokenLength` bytes; use instead of hand-counted literals.

### White-box seams in their own `_test.go` files
- `clientForServer(t, server)` (package `client`) — extracts port from `httptest.Server.URL` and calls `NewClient(port)`; avoids import cycles.
- `redirectTasksSaveFile(t)` / `redirectWorkspacesSaveFile(t)` (package `repository`) — `defer redirectTasksSaveFile(t)()` redirects persistence to a temp file. Consume these; do not recreate.

## Coverage
- Target threshold: 0.8 per module under test.
- Report command: `go test -coverprofile=cover.out ./... && go tool cover -func=cover.out` (add `-html=cover.out` for a browser view).
- Achieved: `internal/repository` 0.825, `internal/client` 0.933, `internal/vscode` 0.806.
- Exclusions: generated code; plain DTO/model structs with no logic (`internal/models`); `main`/Cobra command bootstrapping (`cmd/`); thin production adapters that only wrap third-party I/O (`realProcessInspector`/`gopsutilProcessNode` over gopsutil, `productionTaskRepository`/`productionWorkspaceRepository` over `repository`) — covered through integration, not unit assertions.

## Patterns Established
- Builder pattern for domain fixtures (`TaskBuilder`, `WorkspaceBuilder`) with fluent `With*` setters and a terminal `Build()`.
- Fake HTTP bridge via `httptest.Server` behind `NewBridgeTestServer` + `BridgeHandlerConfig`, with per-handler overrides for error-path coverage.
- Constructor-injection seams: business logic depends on small interfaces (`BridgeClient`, `TaskRepository`, `WorkspaceRepository`, `ProcessInspector`/`ProcessNode`); `NewRunnerWithDeps(RunnerDeps{...})` is the test entry point, `NewRunner` is the production wiring.
- Direct unexported-function tests for seams that hardwire production adapters: call `detectParentVSCode(fakeInspector)` and `selectBridge(bridges, strings.NewReader("1\n"))` directly rather than going through `discoverFromParentProcess`/`discoverFromScan`.
- One-field-at-a-time validation: a "valid baseline" builder (`newValidBridgeInfo()`) plus single-field overrides drives each rejection path of `validateBridgeStructure` independently.
- Seam injection for filesystem persistence: `defer redirect*SaveFile(t)()` points writes at a temp file so repository tests stay hermetic.
- Shared 32-byte `ValidTestToken` constant and `minValidToken()` helper so every auth-path test uses a token that satisfies `security.MinTokenLength`.
- Table-driven subtests with `wantErr` substring matching as the uniform assertion shape.
- Bug-documenting tests: capture unfixable production bugs as skipped tests with a `BUG:` prefix and `t.Skip`; once fixed, un-skip, rename to a positive assertion, and keep as a regression guard.

## What to Test
- Repository persistence round-trips: save then load, empty-state, and overwrite paths.
- Client request/response behavior against the fake bridge: success, auth failure, and non-2xx handling.
- Auth loading and token validation through `LoadAuthFromToken` with `ValidTestToken`.
- Validation and rejection logic: every branch of `validateBridgeStructure` (port bounds, token length, `Secure` flag).
- Discovery seam behavior: parent-process matching by workspace path, interactive bridge selection from stdin.
- Runner orchestration: task/workspace lookup, payload handoff to the client, and error propagation from repository and client failures.
- Error branches and boundary inputs at package boundaries (empty files, missing files, bad permissions).

## What NOT to Test
- Plain DTO/model structs in `internal/models` with no behavior.
- `cmd/` Cobra wiring and `main` bootstrapping.
- Generated code.
- Thin production adapters wrapping third-party libraries (gopsutil process traversal, `repository` file I/O).
- Third-party libraries (Cobra, Bubbletea, lipgloss, gopsutil) — test our usage, not their internals.
- Real filesystem `/tmp/vstr-bridge` scanning and real VSCode process trees — replaced by fakes; do not reach for live system state in unit tests.

## Fixed Bugs (kept as regression guards)
- `repository.SaveFromFile` previously failed on a fresh install: it read the destination `tasks.json` and returned "Provided file is empty" whenever that file was empty, but `ensureTasksSaveFile()` creates an empty `tasks.json` on a fresh system, so import could never succeed on a new machine. Fixed to treat an empty destination as zero existing tasks. Guarded by `TestSaveFromFile_importsIntoFreshSystem` in `internal/repository/repository_tasks_test.go` — do not weaken it.
