# 🚀 VSTR Runner - VSCode Terminal Runner

> **Automate your development workflow** - Launch multiple development projects with a single command

[![Go Version](https://img.shields.io/badge/Go-1.24.5-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](CONTRIBUTING.md)

## 🎯 What is VSTR Runner?

VSTR Runner is a powerful CLI tool that **eliminates the pain** of manually setting up your development environment. With configurable workspaces, you can launch all your project terminals and commands with a single command.

Perfect for developers working with **microservices**, **full-stack applications**, or any **multi-project setup**.

## ❌ The Problem

**Before VSTR Runner:**
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

**After VSTR Runner:**
```bash
vstr run my-project  # 🎉 Everything launches automatically
```

## ✅ What VSTR Runner Solves

- ❌ **Manual terminal management** → ✅ **Automated workspace setup**
- ❌ **Repetitive daily commands** → ✅ **One-command project launch**
- ❌ **Forgetting service dependencies** → ✅ **Configured execution order**
- ❌ **Context switching overhead** → ✅ **Instant development environment**

## 🎬 Quick Demo

```bash
# Install VSTR Runner
go install github.com/yourusername/vscode-terminal-runner@latest

# Configure your workspace
vstr init my-fullstack-app

# Launch everything at once
vstr run my-fullstack-app
```

## 🛠️ Technology Stack

**Built with Modern Go:**
- 🖥️  **[Cobra](https://cobra.dev/)** - Powerful CLI framework with auto-completion
- 🎨  **[Bubbletea](https://github.com/charmbracelet/bubbletea)** - Elegant TUI components  
- 🎭  **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Beautiful styling
- ⚡  **[samber/lo](https://github.com/samber/lo)** - Functional programming utilities
- 🔍  **[gopsutil](https://github.com/shirou/gopsutil)** - Cross-platform process detection

## 📁 Project Structure

```
vscode-terminal-runner/
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
go install github.com/yourusername/vscode-terminal-runner@latest

# Or download binary from releases
curl -L https://github.com/yourusername/vscode-terminal-runner/releases/latest/download/vstr-linux-amd64 -o vstr
chmod +x vstr && sudo mv vstr /usr/local/bin/
```

### Quick Setup

1. **Initialize your workspace:**
   ```bash
   vstr init my-project
   ```

2. **Configure your tasks:**
   ```bash
   vstr edit my-project
   ```

3. **Launch your development environment:**
   ```bash
   vstr run my-project
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

Give a ⭐ if VSTR Runner helps streamline your development workflow!

---

**Made with ❤️ by developers, for developers**