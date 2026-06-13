# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-06-13

First public release of `vstr`, a CLI that drives VSCode terminals through the
VSTR-Bridge extension. Define reusable tasks and workspaces once, then launch
your whole development environment with a single command.

### Added

#### Tasks

- `vstr task` (alias `t`) — interactive TUI to create, edit, list, delete, and
  run tasks.
- Task names are enforced unique on create and rename (case-sensitive),
  preventing accidental duplicates.

#### Workspaces

- `vstr workspace` (alias `w`) — interactive TUI to create, edit, list, delete,
  and run workspaces.
- A workspace groups multiple tasks and launches them together across VSCode
  terminals with a single command.

#### Bulk import

- `vstr import --from-json <file|->` — import tasks and workspaces in bulk from
  a JSON file or stdin. Validation is all-or-nothing: if any entry is invalid,
  nothing is imported.

#### Setup and diagnostics

- `vstr setup` — guided wizard that installs the VSTR-Bridge VSCode extension
  and configures the CLI.
- `vstr status` — reports whether the VSTR-Bridge extension is installed and
  whether the CLI can reach the active bridge.
- `vstr version` subcommand and `--version`/`-v` flag — print the installed
  version, resolved from build metadata.

#### Persistence and reliability

- Tasks, workspaces, and config are stored as JSON under the user config
  directory.
- JSON writes are atomic (temp file plus rename) to avoid corrupting your data
  on partial writes.

#### Bridge connectivity

- Automatic bridge discovery via environment variables, the parent process
  tree, and a `/tmp/vstr-bridge` scan, with auth-token and file-permission
  validation before trusting a bridge.
