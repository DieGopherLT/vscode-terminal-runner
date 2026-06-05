---
paths: ["**/*_test.go"]
---

# Testing

## How to Run Tests

- All tests: `go test ./...`
- Single package: `go test ./internal/vscode/...`
- Single test: `go test ./internal/vscode/ -run TestValidateBridgeStructure`
- Verbose: add `-v`. No watch mode is configured; re-run the command on change.

## Test Organization

- Co-located with production code: `*_test.go` lives in the same directory as the code it covers.
- File naming: `<package>_test.go` for the test cases (e.g. `internal/vscode/vscode_test.go`); seam fakes/builders for a package go in `fakes_test.go`.
- Structure: table-driven subtests. Each test defines a `tests := []struct{ name string; ...; wantErr string }` slice and loops with `t.Run(tt.name, ...)`. One behavior per table row; assert on `wantErr` substrings via `strings.Contains`.
- White-box by default: tests use `package vscode` (not `vscode_test`) so they reach unexported functions (`validateBridgeStructure`, `detectParentVSCode`, `selectBridge`) and package-local fakes directly.

## Test Utilities

Reuse-before-create: check both locations below before writing a new builder or fake. Only add a new one when no existing utility fits the seam.

Two homes, chosen by import-cycle constraints:

- `pkg/testutils` — cross-cutting builders importable from any package (`import "github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils"`):
  - `TaskBuilder` (builder): `testutils.NewTask().WithName("build").WithCmds("go build ./...").Build()`
  - `WorkspaceBuilder` (builder): `testutils.NewWorkspace().WithName("dev").WithTasks(testutils.NewTask().Build()).Build()`
- `internal/vscode` (`fakes_test.go`, `package vscode`) — seam fakes that MUST live in-package to avoid import cycles; usable only from white-box tests:
  - `fakeBridgeClient` (fake) — satisfies `BridgeClient`. Inject via `NewRunnerWithDeps(RunnerDeps{Client: client, ...})`; assert on `client.lastTask` after `RunTask()`. Drive errors with `executeTaskErr`.
  - `fakeTaskRepository` (fake) — satisfies `TaskRepository`. `&fakeTaskRepository{task: &models.Task{...}}` or `{returnErr: errors.New("not found")}` for the error path.
  - `fakeWorkspaceRepository` (fake) — satisfies `WorkspaceRepository`. `&fakeWorkspaceRepository{workspace: &models.Workspace{...}}`.
  - `fakeProcessInspector` + `fakeProcessNode` + `buildProcessChain` (fakes) — satisfy `ProcessInspector`/`ProcessNode`. Build a tree with `buildProcessChain(fakeNodeSpec{pid:100,name:"bash"}, fakeNodeSpec{pid:200,name:"code",cmdline:"--folder-uri file:///workspace/myapp"})`, then `detectParentVSCode(&fakeProcessInspector{ppid:100, tree:tree})`.
  - `bridgeInfoBuilder` via `newValidBridgeInfo()` (builder) — produces a `BridgeInfo` that passes `validateBridgeStructure`; override one field per rejection test: `.WithPort(0)`, `.WithShortToken()`, `.WithSecure(false)`.
  - `minValidToken()` (helper) — returns a token of exactly `security.MinTokenLength` bytes; use instead of hand-counted literals.

## Coverage

- Target threshold: 80% (`internal/vscode` is at 80.6%).
- Report: `go test ./internal/vscode/ -cover` for the summary; `go test ./internal/vscode/ -coverprofile=cover.out && go tool cover -func=cover.out` for per-function detail (or `-html=cover.out` for an annotated view).
- Exclusions: production seam adapters that only wrap third-party I/O (`realProcessInspector`/`gopsutilProcessNode` over gopsutil, `productionTaskRepository`/`productionWorkspaceRepository` over `repository`) are thin pass-throughs exercised end-to-end, not unit-targeted; pure data structs in `internal/models` carry no logic to cover.

## Patterns Established

- Constructor-injection seams: business logic depends on small interfaces (`BridgeClient`, `TaskRepository`, `WorkspaceRepository`, `ProcessInspector`/`ProcessNode`); `NewRunnerWithDeps(RunnerDeps{...})` is the test entry point, `NewRunner` is the production wiring.
- Direct unexported-function tests for seams that hardwire production adapters: call `detectParentVSCode(fakeInspector)` and `selectBridge(bridges, strings.NewReader("1\n"))` directly rather than going through `discoverFromParentProcess`/`discoverFromScan`.
- One-field-at-a-time validation: a "valid baseline" builder (`newValidBridgeInfo()`) plus single-field overrides drives each rejection path of `validateBridgeStructure` independently.
- Table-driven subtests with `wantErr` substring matching as the uniform assertion shape.
- Constant-derived test data (`minValidToken()` ties to `security.MinTokenLength`) so tests stay correct if the constant changes.

## What to Test

- Validation and rejection logic: every branch of `validateBridgeStructure` (port bounds, token length, `Secure` flag).
- Discovery seam behavior: parent-process matching by workspace path, interactive bridge selection from stdin.
- Runner orchestration: task/workspace lookup, payload handoff to the client, and error propagation from repository and client failures.
- Both happy and error paths — error paths driven by fake `returnErr`/`executeTaskErr` fields.

## What NOT to Test

- Thin production adapters wrapping third-party libraries (gopsutil process traversal, `repository` file I/O) — covered through integration, not unit assertions.
- Pure data structs (`models.Task`, `models.Workspace`, `BridgeInfo` field storage) with no behavior.
- The `gopsutil`/`samber/lo`/cobra dependencies themselves — trust the library; test only our use of it via the seam interfaces.
- Real filesystem `/tmp/vstr-bridge` scanning and real VSCode process trees — replaced by fakes; do not reach for live system state in unit tests.
