# VSTR Runner - Technical Documentation & Development Guidelines

## Project Information

**Binary Name:** `vstr`
**Root Command:** `vstr` (not `vsct-runner`)

## Project Overview

For project context and general information, see [README.md](README.md).

## Response Guidelines

### Code Generation Language

All generated code and documentation comments must be written in **English**.

- Variable names: English
- Function names: English  
- Comments: English
- Documentation: English
- Error messages: English

### Response Language

Respond to the user in the **same language** as their message:

- If user writes in English → Respond in English
- If user writes in Spanish → Respond in Spanish  
- If user writes in another language → Respond in that language

The code itself always remains in English, but explanations and conversations should match the user's language preference.

## Code Generation Guidelines

### Clean Code Principles

#### 1. Guard Clauses over Nesting

**❌ Bad - Excessive nesting:**
```go
func (t *TaskModel) validateTask(task models.Task) bool {
    if strings.TrimSpace(task.Name) != "" {
        if len(task.Cmds) > 0 {
            if task.Icon != "" {
                _, iconExists := lo.Find(styles.VSCodeIcons, func(i styles.VSCodeIcon) bool {
                    return i.Name == task.Icon
                })
                if iconExists {
                    return true
                }
            }
        }
    }
    return false
}
```

**✅ Good - Guard clauses:**
```go
// validateTask checks if the task contains all required fields and returns validation result.
func (t *TaskModel) validateTask(task models.Task) bool {
    if strings.TrimSpace(task.Name) == "" {
        t.messages.AddError("Name is required")
        return false
    }
    
    if len(task.Cmds) == 0 {
        t.messages.AddError("At least one command is required")
        return false
    }
    
    if task.Icon == "" {
        t.messages.AddError("Icon is required")
        return false
    }
    
    _, iconExists := lo.Find(styles.VSCodeIcons, func(i styles.VSCodeIcon) bool {
        return i.Name == task.Icon
    })
    if !iconExists {
        t.messages.AddError("Invalid icon")
        return false
    }
    
    return true
}
```

#### 2. Functional Approach over Imperative

**❌ Bad - Imperative loop:**
```go
func getTaskNames(tasks []models.Task) []string {
    var names []string
    for i := 0; i < len(tasks); i++ {
        if tasks[i].Name != "" {
            names = append(names, tasks[i].Name)
        }
    }
    return names
}
```

**✅ Good - Functional transformation:**
```go
// getTaskNames extracts valid task names from a slice of tasks.
func getTaskNames(tasks []models.Task) []string {
    return lo.FilterMap(tasks, func(task models.Task, _ int) (string, bool) {
        hasValidName := task.Name != ""
        return task.Name, hasValidName
    })
}
```

#### 3. Self-Descriptive Variable Names & Documentation

**Note:** Receivers can remain as single/double letters (e.g., `t`, `sm`) as they're understood in context. Focus on descriptive names for value-storing variables.

**❌ Bad - Cryptic variable names:**
```go
func (sm *Manager) u(i string) {
    if i == sm.li {
        return
    }
    sm.li = i
    if i == "" {
        sm.fs = sm.as
    } else {
        sm.fs = make([]string, 0)
        for _, s := range sm.as {
            if sm.ff(s, i) {
                sm.fs = append(sm.fs, s)
            }
        }
    }
    sm.si = 0
}
```

**✅ Good - Descriptive names with documentation:**
```go
// UpdateFilter filters suggestions based on input text and resets selection only if input changed.
func (sm *Manager) UpdateFilter(inputText string) {
    if inputText == sm.lastInput {
        return
    }
    
    sm.lastInput = inputText
    
    if inputText == "" {
        sm.filteredSuggestions = sm.allSuggestions
    } else {
        sm.filteredSuggestions = lo.Filter(sm.allSuggestions, func(suggestion string, _ int) bool {
            return sm.filterFunc(suggestion, inputText)
        })
    }
    
    sm.selectedIndex = 0
}
```

#### 4. Extract Complex Conditions

**❌ Bad - Complex inline condition:**
```go
func (sm *Manager) ShouldShow(input string) bool {
    if (!sm.showOnEmpty && input == "") || len(sm.GetVisible()) == 0 || (len(sm.GetVisible()) == 1 && sm.GetVisible()[0] == input) {
        return false
    }
    return true
}
```

**✅ Good - Extracted conditions:**
```go
// ShouldShow determines if suggestions should be displayed based on current state and input.
func (sm *Manager) ShouldShow(inputText string) bool {
    shouldHideOnEmptyInput := !sm.showOnEmpty && inputText == ""
    if shouldHideOnEmptyInput {
        return false
    }
    
    visibleSuggestions := sm.GetVisible()
    
    hasNoSuggestions := len(visibleSuggestions) == 0
    if hasNoSuggestions {
        return false
    }
    
    hasExactMatchOnly := len(visibleSuggestions) == 1 && visibleSuggestions[0] == inputText
    if hasExactMatchOnly {
        return false
    }
    
    return true
}
```

### Project-Specific Patterns

#### Error Handling with Guard Clauses
```go
// RunTask executes a single task in a new VSCode terminal.
func (r *Runner) RunTask(taskName string) error {
    if taskName == "" {
        return fmt.Errorf("task name cannot be empty")
    }
    
    task, err := repository.FindTaskByName(taskName)
    if err != nil {
        return fmt.Errorf("task not found: %w", err)
    }
    
    if err := r.launcher.LaunchTerminal(*task); err != nil {
        return fmt.Errorf("failed to launch terminal: %w", err)
    }
    
    return nil
}
```

#### Functional Transformations with lo
```go
// Convert structs to names
iconNames := lo.Map(styles.VSCodeIcons, func(icon styles.VSCodeIcon, _ int) string { 
    return icon.Name 
})

// Combined filtering and mapping
validTasks := lo.FilterMap(tasks, func(task models.Task, _ int) (models.Task, bool) {
    isValid := task.Name != "" && len(task.Cmds) > 0
    return task, isValid
})
```

### Documentation Requirements

- **All functions and methods** must include documentation comments
- Use standard Go documentation format: `// FunctionName does...`
- Explain what the function does, not how it does it
- Include parameter and return value descriptions when not obvious

## Testing Guidelines

### Overview

To deliver a robust and bug-free application, comprehensive testing is a core requirement. Our architecture emphasizes dependency injection through dedicated types, making the codebase inherently more testable and maintainable.

### Testing Philosophy

The application follows a **dependency injection pattern** where specialized types provide specific features. This architectural decision enables:

- **Isolated unit testing** - Each component can be tested in isolation
- **Mock-friendly design** - Dependencies can be easily mocked or stubbed
- **Clear separation of concerns** - Business logic is decoupled from external dependencies
- **Improved maintainability** - Changes in one component don't cascade through the system

### Testing Responsibilities

When working with test-related requests, you have two primary responsibilities:

#### 1. Testability Assessment

Evaluate whether the code under test follows testable patterns:

- **Dependency injection** - Are external dependencies injected rather than hardcoded?
- **Single responsibility** - Does each function/method have a clear, focused purpose?
- **Pure functions** - Are functions free from side effects where possible?
- **Interface segregation** - Are dependencies abstracted behind interfaces?

If the code doesn't meet these criteria, suggest refactoring improvements to enhance testability.

#### 2. Test Implementation

Write comprehensive tests that cover:

- **Happy path scenarios** - Normal operation flows
- **Edge cases** - Boundary conditions and unusual inputs
- **Error handling** - Failure scenarios and error propagation
- **Integration points** - Interactions between components

### Testing Standards

#### Test Structure
- Use **table-driven tests** for multiple scenarios
- Follow **AAA pattern** (Arrange, Act, Assert)
- Include **descriptive test names** that explain the scenario

#### Mock Strategy
- Mock **external dependencies** (file system, network, processes)
- Use **real implementations** for internal logic when possible
- Prefer **interface mocks** over concrete type mocks

#### Coverage Goals
- Aim for **high code coverage** without compromising test quality
- Focus on **critical paths** and **business logic**
- Ensure **error paths** are adequately tested

#### Example Testing Approach

```go
// Testable function with dependency injection
func (r *Runner) ExecuteTask(task models.Task, launcher TerminalLauncher) error {
    // Business logic that can be easily tested
}

// Corresponding test with mocked dependency
func TestRunner_ExecuteTask(t *testing.T) {
    mockLauncher := &MockTerminalLauncher{}
    runner := NewRunner()
    
    // Test implementation
}
```

## Important Reminders

- Do what has been asked; nothing more, nothing less
- **NEVER** create files unless they're absolutely necessary for achieving your goal
- **ALWAYS** prefer editing an existing file to creating a new one
- **NEVER** proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User