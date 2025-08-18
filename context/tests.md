# VSTR Runner - Testing Guidelines

## Overview

To deliver a robust and bug-free application, comprehensive testing is a core requirement. Our architecture emphasizes dependency injection through dedicated types, making the codebase inherently more testable and maintainable.

## Testing Philosophy

The application follows a **dependency injection pattern** where specialized types provide specific features. This architectural decision enables:

- **Isolated unit testing** - Each component can be tested in isolation
- **Mock-friendly design** - Dependencies can be easily mocked or stubbed
- **Clear separation of concerns** - Business logic is decoupled from external dependencies
- **Improved maintainability** - Changes in one component don't cascade through the system

## Testing Responsibilities

When working with test-related requests, you have two primary responsibilities:

### 1. Testability Assessment

Evaluate whether the code under test follows testable patterns:

- **Dependency injection** - Are external dependencies injected rather than hardcoded?
- **Single responsibility** - Does each function/method have a clear, focused purpose?
- **Pure functions** - Are functions free from side effects where possible?
- **Interface segregation** - Are dependencies abstracted behind interfaces?

If the code doesn't meet these criteria, suggest refactoring improvements to enhance testability.

### 2. Test Implementation

Write comprehensive tests that cover:

- **Happy path scenarios** - Normal operation flows
- **Edge cases** - Boundary conditions and unusual inputs
- **Error handling** - Failure scenarios and error propagation
- **Integration points** - Interactions between components

## Testing Standards

### Test Structure
- Use **table-driven tests** for multiple scenarios
- Follow **AAA pattern** (Arrange, Act, Assert)
- Include **descriptive test names** that explain the scenario

### Mock Strategy
- Mock **external dependencies** (file system, network, processes)
- Use **real implementations** for internal logic when possible
- Prefer **interface mocks** over concrete type mocks

### Coverage Goals
- Aim for **high code coverage** without compromising test quality
- Focus on **critical paths** and **business logic**
- Ensure **error paths** are adequately tested

## Example Testing Approach

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

By following these guidelines, we ensure that our codebase remains maintainable, reliable, and easy to test.