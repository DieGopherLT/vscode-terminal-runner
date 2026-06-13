---
title: vstr --version
version: 1.0
date_created: 2026-06-12
last_updated: 2026-06-12
owner: DieGopherLT
tags: [tool, cli, versioning]
---

# Introduction

This specification defines version reporting for the `vstr` CLI via `vstr --version`. The
version is derived at runtime from the Go module build metadata, requiring no manual version
constant or build-time `-ldflags`.

## 1. Purpose & Scope

Expose the installed version of `vstr` through the conventional `--version` flag (and the
`vstr version` form Cobra derives from `rootCmd.Version`). In scope: wiring `rootCmd.Version`
to a value resolved from `runtime/debug.ReadBuildInfo`. Out of scope: a changelog command,
update checks, or build-time injection.

## 2. Definitions

- **Build info**: the module metadata embedded by the Go toolchain, read via
  `runtime/debug.ReadBuildInfo()`.
- **Main.Version**: the module version string in build info. It carries the VCS tag (e.g.
  `v1.2.0`) when installed via `go install module@tag`/`@latest`, and `(devel)` for a local
  `go build` without a tag.

## 3. Requirements, Constraints & Guidelines

- **REQ-001**: `vstr --version` SHALL print the resolved version string.
- **REQ-002**: The version SHALL be resolved from `runtime/debug.ReadBuildInfo().Main.Version`.
- **REQ-003**: When build info is unavailable, or `Main.Version` is empty or `(devel)`, the
  resolved version SHALL fall back to `"dev"`.
- **REQ-004**: `rootCmd.Version` SHALL be set in `cmd/root.go` so Cobra renders both
  `--version` and the implicit `version` subcommand. It is currently unset.
- **CON-001**: No manual `var version` literal and no `-ldflags -X` injection SHALL be required;
  resolution is runtime-only. This matches the distribution method in `install.sh`
  (`go install github.com/DieGopherLT/vscode-terminal-runner@latest`), which records the tag.
- **GUD-001**: Keep resolution in a small unexported helper (e.g. `resolveVersion() string` in
  `cmd/root.go`) called from `init()` when assigning `rootCmd.Version`.

## 4. Interfaces & Data Contracts

### Resolution helper (to implement, `cmd/root.go`)

```go
import "runtime/debug"

func resolveVersion() string {
    info, ok := debug.ReadBuildInfo()
    if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
        return "dev"
    }
    return info.Main.Version
}
```

### Wiring (existing site, `cmd/root.go`)

`rootCmd` is declared at `cmd/root.go:14`; `init()` at `cmd/root.go:46` currently only calls
`rootCmd.AddCommand(cfg.SetupCMD)`. Set the version there or at declaration:

```go
var rootCmd = &cobra.Command{
    Use:     "vstr",
    Version: resolveVersion(),
    // ... existing Short/Long/Run
}
```

### CLI surface

```
vstr --version
vstr -v            # only if no other -v shorthand conflicts; otherwise omit
vstr version
```

## 5. Acceptance Criteria

- **AC-001**: Given `vstr` installed via `go install ...@v1.2.0`, When `vstr --version` runs,
  Then it prints a string containing `v1.2.0`.
- **AC-002**: Given a local `go build` binary, When `vstr --version` runs, Then it prints `dev`.
- **AC-003**: Given `runtime/debug.ReadBuildInfo` returns not-ok, When version resolves, Then
  the value is `dev` (no panic).
- **AC-004**: `vstr version` (subcommand form) SHALL print the same value as `vstr --version`.

## 6. Test Automation Strategy

- **Test Levels**: Unit for `resolveVersion` fallback branches (empty / `(devel)` / not-ok).
- **Frameworks**: Go standard `testing`. Build info cannot be forced in a unit test directly;
  factor the pure decision into a helper, e.g.
  `resolveVersionFrom(info *debug.BuildInfo, ok bool) string`, and unit-test that.
- **CI/CD Integration**: `go test ./...` must pass.

## 7. Rationale & Context

`install.sh` distributes `vstr` via `go install ...@latest`, so the Go toolchain already
stamps the VCS tag into build info. Reading it at runtime is the idiomatic approach for this
distribution method: it needs no release-time ldflags ceremony and stays correct as tags
advance. The only gap is local builds without a tag, which report `(devel)`; mapping that to
`dev` yields a clean, predictable string.

## 8. Dependencies & External Integrations

### Technology Platform Dependencies

- **PLT-001**: Go 1.24+ standard library (`runtime/debug`), `spf13/cobra` (`rootCmd.Version`).

## 9. Examples & Edge Cases

```
$ vstr --version
vstr version v1.2.0

$ vstr --version        # local go build
vstr version dev
```

Edge cases:

- Empty `Main.Version` (older toolchains / odd build modes) -> `dev`.
- `(devel)` from untagged build -> `dev`.
- `ReadBuildInfo` not-ok (binary stripped of build info) -> `dev`, never panic.

## 10. Validation Criteria

- `rootCmd.Version` is set; `vstr --version` and `vstr version` both print the resolved value.
- Untagged local builds print `dev`; tagged installs print the tag.

## 11. Related Specifications / Further Reading

- `spec-tool-json-import.md`
- `spec-tool-status-command.md`
- Go docs: `runtime/debug.ReadBuildInfo`; Cobra `Command.Version`.
