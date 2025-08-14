# VSTR Runner - Project Context

## VSTR Runner

**Acronym:** VSCode Terminal Runner

## What is it?

A CLI tool that automates the management and execution of multiple development projects in VSCode terminals through configurable workspaces.

## What problem does it solve?

Eliminates the manual and repetitive process of:
- Opening multiple terminals manually
- Executing startup commands one by one
- Managing interdependent projects (microservices, APIs + Frontend)
- Remembering the correct execution order of services
- Time loss on repetitive daily setup tasks

## Technology Stack

**Language:** Go 1.24.5

**Main Libraries:**
- **Cobra** - CLI command handling and auto-completion
- **Bubbletea** - TUI engine for interactive interfaces
- **Bubbles** - UI components (textinput, spinners)
- **Lipgloss** - Styling and visual presentation
- **gopsutil** - VSCode process detection
- **samber/lo** - Functional utilities for Go

## Architecture

```
vscode-terminal-runner/
├── cmd/           # Main CLI commands (Cobra)
├── internal/      # Domain business logic
├── pkg/           # Reusable components
├── docs/          # Technical documentation
├── main.go        # Entry point
└── go.mod         # Dependency management
```