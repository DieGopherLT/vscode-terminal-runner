# 🚀 VSCode Terminal Runner - CLI

> **Automate your development workflow** - Launch multiple development projects with a single command

[![Go Version](https://img.shields.io/badge/Go-1.24.5-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](CONTRIBUTING.md)

## 🎯 What is VSCode Terminal Runner?

VSCode Terminal Runner is a powerful CLI tool that **eliminates the pain** of manually setting up your development environment. With configurable tasks and workspaces, you can launch all your project terminals and commands through VSCode.

Perfect for developers working with **microservices**, **full-stack applications**, or any **multi-project setup**.

## ❌ The Problem

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
vstr workspace run my-project  # 🎉 Everything launches automatically in VSCode
```

## ✅ What VSCode Terminal Runner Solves

- ❌ **Manual terminal management** → ✅ **Automated workspace setup**
- ❌ **Repetitive daily commands** → ✅ **One-command project launch**
- ❌ **Forgetting service dependencies** → ✅ **Configured execution order**
- ❌ **Context switching overhead** → ✅ **Instant development environment**

## 🎬 Quick Demo

```bash
# 1. Install VSTR-Bridge VSCode Extension
# Visit: https://github.com/DieGopherLT/VSTR-Bridge

# 2. Install VSCode Terminal Runner CLI
go install github.com/DieGopherLT/vscode-terminal-runner@latest

# 3. Create a task
vstr task create

# 4. Create a workspace with multiple tasks
vstr workspace create

# 5. Launch workspace in VSCode
vstr workspace run my-fullstack-app
```

## 🛠️ Technology Stack

**Built with Modern Go:**
- 🖥️  **[Cobra](https://cobra.dev/)** - Powerful CLI framework with auto-completion
- 🎨  **[Bubbletea](https://github.com/charmbracelet/bubbletea)** - Elegant TUI components  
- 🎭  **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Beautiful styling
- ⚡  **[samber/lo](https://github.com/samber/lo)** - Functional programming utilities
- 🔍  **[gopsutil](https://github.com/shirou/gopsutil)** - Cross-platform process detection

## 📁 Project Structure

**This repository contains the CLI component. The VSCode extension is in a separate repository: [VSTR-Bridge](https://github.com/DieGopherLT/VSTR-Bridge)**

```
vscode-terminal-runner/ (CLI Component)
├── cmd/                    # CLI commands and entry points
├── internal/
│   ├── models/            # Data structures
│   ├── repository/        # Data persistence
│   ├── launcher/          # VSCode integration
│   └── tui/               # Terminal UI components
├── pkg/                   # Reusable packages
├── docs/                  # Documentation
└── main.go               # Application entry point
```

## 🚀 Getting Started

### Installation

```bash
# Install from source
go install github.com/DieGopherLT/vscode-terminal-runner@latest

# Or download binary from releases
curl -L https://github.com/DieGopherLT/vscode-terminal-runner/releases/latest/download/vstr-linux-amd64 -o vstr
chmod +x vstr && sudo mv vstr /usr/local/bin/
```

### Prerequisites

**VSCode Extension Required:**

This CLI requires the **VSTR-Bridge** VSCode extension to function properly. The extension handles the communication between the CLI and VSCode terminals.

**Install the extension:**
- Repository: [VSTR-Bridge Extension](https://github.com/DieGopherLT/VSTR-Bridge)
- **Future enhancement:** A `vstr setup` command will be available to automatically install the extension with user consent

### Quick Setup

1. **Create your first task:**
   ```bash
   vstr task create
   ```

2. **Create a workspace:**
   ```bash
   vstr workspace create
   ```

3. **Run a task:**
   ```bash
   vstr task run my-task
   ```

4. **Run a workspace:**
   ```bash
   vstr workspace run my-workspace
   ```

### Available Commands

#### Task Management
```bash
vstr task create           # Interactive form to create a new task
vstr task list            # List all tasks
vstr task list --only-names  # List task names only
vstr task edit <name>     # Edit an existing task
vstr task run <name>      # Run a specific task
vstr task delete <name>   # Delete a task
```

#### Workspace Management
```bash
vstr workspace create     # Interactive form to create a new workspace
vstr workspace list      # List all workspaces
vstr workspace run <name> # Run all tasks in a workspace
```

## 📖 Use Cases

- **🔧 Full-stack Development**: Launch frontend, backend, and database simultaneously
- **🏗️ Microservices Architecture**: Start all services in correct dependency order  
- **🧪 Testing Environments**: Set up complex test scenarios with multiple components
- **🚀 DevOps Workflows**: Automate local development environment setup

## 🤝 Contributing

We welcome contributions! This project follows clean code principles and modern Go practices.

- **Development Guidelines**: See [CLAUDE.md](CLAUDE.md)
- **Code Standards**: Functional programming with guard clauses
- **Testing**: Comprehensive test coverage with dependency injection
- **Documentation**: Every function documented

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Show Your Support

Give a ⭐ if VSCode Terminal Runner helps streamline your development workflow!

---

**Made with ❤️ by developers, for developers**

---

## 🎯 VSCode Integration

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

### Future Enhancements
A `vstr setup` command is planned to automatically install and configure the VSTR-Bridge extension with user consent, making the initial setup even simpler.