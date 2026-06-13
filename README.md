# VSCode Terminal Runner - CLI

> **Automate your development workflow** - Launch multiple development projects with a single command

[![Go Version](https://img.shields.io/badge/Go-1.24.5-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)

## What is VSCode Terminal Runner?

VSCode Terminal Runner is a CLI tool that **eliminates the pain** of manually setting up your development environment. With configurable tasks and workspaces, you can launch all your project terminals and commands through VSCode.

Perfect for developers working with **microservices**, **full-stack applications**, or any **multi-project setup**.

## The Problem

**Before VSCode Terminal Runner:**

```bash
# Every day, manually:
cd frontend && npm run dev
# New terminal
cd backend && npm run server
# New terminal
cd database && docker-compose up
# New terminal
cd api-gateway && go run main.go
# ... and so on
```

**After VSCode Terminal Runner:**

```bash
vstr workspace run my-project
# Under the hood, vstr creates every terminal and runs every command
# in the same order they were configured in your workspace.
```

## What VSCode Terminal Runner Solves

- Manual terminal management -> Automated workspace setup
- Repetitive daily commands -> One-command project launch
- Forgetting service dependencies -> Configured execution order
- Context switching overhead -> Instant development environment

## VSCode Integration

This tool is specifically designed to integrate with VSCode through the **VSTR-Bridge** extension. The architecture works as follows:

### How it Works

- **CLI Component** (`vstr`): Manages tasks, workspaces, and user configuration
- **VSCode Extension** ([VSTR-Bridge](https://github.com/DieGopherLT/VSTR-Bridge)): Handles terminal creation and command execution within VSCode
- **Communication**: The CLI communicates with the extension to automatically open terminals and run commands within your VSCode workspace

### Features

- Seamless integration with VSCode terminal system
- Automatic terminal creation and management
- Commands execute in the context of your VSCode workspace
- No manual terminal switching required

## Installation

### Prerequisites

- [Go 1.24+](https://go.dev/dl/) must be installed on your system.

```sh
curl -sSL https://raw.githubusercontent.com/DieGopherLT/vscode-terminal-runner/main/install.sh | sh
```

Then run `vstr setup` to install the required VSCode extension ([VSTR-Bridge](https://github.com/DieGopherLT/VSTR-Bridge)) directly from GitHub Releases.

### Quick Setup

A **task** is the unit of execution: it runs one or more commands inside a dedicated VSCode terminal. A **workspace** groups several tasks together and launches each one in its own terminal, in the order you defined.

1. **Create your first task:**

   ```bash
   vstr task create   # or: vstr t create
   ```

2. **Create a workspace:**

   ```bash
   vstr workspace create   # or: vstr w create
   ```

3. **Run a task:**

   ```bash
   vstr task run my-task   # or: vstr t run my-task
   ```

4. **Run a workspace:**

   ```bash
   vstr workspace run my-workspace   # or: vstr w run my-workspace
   ```

### Available Commands

#### Task Management

```bash
vstr task create           # Interactive form to create a new task
vstr task list             # List all tasks
vstr task list --only-names  # List task names only
vstr task edit <name>      # Edit an existing task
vstr task run <name>       # Run a specific task
vstr task delete <name>    # Delete a task
```

#### Workspace Management

```bash
vstr workspace create      # Interactive form to create a new workspace
vstr workspace list        # List all workspaces
vstr workspace edit <name> # Edit an existing workspace
vstr workspace run <name>  # Run all tasks in a workspace
vstr workspace delete <name> # Delete a workspace
```

> `task` and `workspace` have short aliases: `t` and `w` respectively.
> For example: `vstr t run my-task` or `vstr w run my-project`.

## Use Cases

- **Full-stack Development**: Launch frontend, backend, and database simultaneously
- **Microservices Architecture**: Start all services in correct dependency order
- **Testing Environments**: Set up complex test scenarios with multiple components
- **DevOps Workflows**: Automate local development environment setup

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Made by DieGopherLT**
