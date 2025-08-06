# Autocomplete System - SuggestionManager

## Overview

The `suggestions.Manager` is a reusable component that provides interactive autocomplete functionality for TUI forms. It handles suggestion filtering, circular navigation, visual rendering, and selection application with optimized performance.

## Architecture

### Main Component: `Manager`

**Location:** `pkg/tui/suggestions/manager.go`

```go
type Manager struct {
    allSuggestions      []string    // All available suggestions
    filteredSuggestions []string    // Current filtered suggestions
    selectedIndex       int         // Currently selected suggestion index
    maxVisible          int         // Maximum suggestions to show
    filterFunc          FilterFunc  // Function to filter suggestions
    lastInput           string      // Last input used for filtering (optimization)
}
```

### Filter Function Type

```go
type FilterFunc func(suggestion, input string) bool
```

The filter function determines which suggestions match the user's input. A default filter is used when nil is passed to the constructor.

## Public API

### Constructor

```go
func NewManager(suggestions []string, maxVisible int, filterFunc FilterFunc) *Manager
```

- `suggestions`: Complete list of available options
- `maxVisible`: Maximum number of suggestions to display (e.g., 3)
- `filterFunc`: Filtering function (uses StartsWithFilter if nil)

### State Methods

```go
func (sm *Manager) SetSuggestions(suggestions []string)
func (sm *Manager) UpdateFilter(input string)
func (sm *Manager) Reset()
```

### Navigation Methods

```go
func (sm *Manager) Next()      // Next suggestion (circular)
func (sm *Manager) Previous()  // Previous suggestion (circular)
```

### Query Methods

```go
func (sm *Manager) GetSelected() string           // Selected suggestion
func (sm *Manager) GetVisible() []string          // Visible suggestions
func (sm *Manager) ShouldShow(input string) bool  // Should show suggestions?
```

### Interaction Methods

```go
func (sm *Manager) ApplySelected(input *textinput.Model)  // Apply selection
func (sm *Manager) Render() string                       // Render UI
```

## Predefined Filters

**Location:** `pkg/tui/suggestions/filters.go`

### Available Filters

- **`StartsWithFilter`** (default) - Matches suggestions that start with input
- **`ContainsFilter`** - Matches suggestions containing input anywhere
- **`WordBoundaryFilter`** - For "terminal-bash" finds "bash"
- **`ExactFilter`** - Exact match only
- **`CaseSensitiveStartsWithFilter`** - Case-sensitive starts with
- **`CaseSensitiveContainsFilter`** - Case-sensitive contains

### Usage Examples

```go
// Different filters for different use cases
iconSuggestions := suggestions.NewManager(iconNames, 3, nil) // StartsWithFilter by default
taskSuggestions := suggestions.NewManager(taskNames, 5, suggestions.ContainsFilter)
cmdSuggestions := suggestions.NewManager(commands, 3, suggestions.WordBoundaryFilter)
```

## Form Integration

### 1. Dependency Injection

```go
type TaskModel struct {
    nav                *tui.FormNavigator
    inputs             []textinput.Model
    iconSuggestions    *suggestions.Manager  // For icon field
    colorSuggestions   *suggestions.Manager  // For color field
}
```

### 2. Initialization

```go
func NewModel() tea.Model {
    iconNames := []string{"terminal-bash", "code", "play"}
    colorNames := []string{"ansiRed", "ansiGreen", "ansiBlue"}
    
    model := &TaskModel{
        nav:              tui.NewNavigator(numberOfFields),
        inputs:           make([]textinput.Model, numberOfFields),
        iconSuggestions:  suggestions.NewManager(iconNames, 3, nil),
        colorSuggestions: suggestions.NewManager(colorNames, 3, suggestions.ContainsFilter),
    }
    return model
}
```

### 3. Event Handling

```go
func (t *TaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+n":  // Next suggestion
            if manager := t.getCurrentSuggestionManager(); manager != nil {
                manager.Next()
            }
        case "ctrl+b":  // Previous suggestion
            if manager := t.getCurrentSuggestionManager(); manager != nil {
                manager.Previous()
            }
        case "tab", "enter":  // Apply suggestion
            if manager := t.getCurrentSuggestionManager(); manager != nil {
                if manager.ShouldShow(t.inputs[t.nav.FocusIndex].Value()) {
                    manager.ApplySelected(&t.inputs[t.nav.FocusIndex])
                }
            }
        }
    }
    // ...
}
```

### 4. Filter Updates

```go
func (t *TaskModel) HandleInput(msg tea.Msg) tea.Cmd {
    for i := range t.inputs {
        t.inputs[i], _ = t.inputs[i].Update(msg)
        
        // Update filters based on active field
        if i == iconField && i == t.nav.FocusIndex {
            t.iconSuggestions.UpdateFilter(t.inputs[i].Value())
        }
        if i == colorField && i == t.nav.FocusIndex {
            t.colorSuggestions.UpdateFilter(t.inputs[i].Value())
        }
    }
}
```

### 5. Rendering

```go
func (t *TaskModel) View() string {
    // For each field with autocomplete
    if t.nav.FocusIndex == i {
        if manager := t.getCurrentSuggestionManager(); manager != nil {
            if manager.ShouldShow(t.inputs[i].Value()) {
                suggestionBox := manager.Render()
                if suggestionBox != "" {
                    // Combine field + suggestions
                    fieldContent = lipgloss.JoinVertical(
                        lipgloss.Left,
                        fieldContent,
                        suggestionBox,
                    )
                }
            }
        }
    }
}
```

## User Behavior

### Keyboard Shortcuts

- **Ctrl+N**: Navigate to next suggestion
- **Ctrl+B**: Navigate to previous suggestion  
- **Tab/Enter**: Apply selected suggestion
- **↑/↓**: Navigate between form fields

### Visual Behavior

1. **Real-time filtering**: Suggestions filter as user types
2. **Circular navigation**: Reaching the end returns to beginning
3. **Visual limit**: Only shows first N suggestions (configurable)
4. **Highlighting**: Selected suggestion is visually highlighted
5. **Auto-hide**: Hides when there's a unique exact match

### Display Logic

Suggestions are shown when:
- There are filtered suggestions available
- NOT a unique exact match
- The field is focused

## Performance Optimizations

### Input Change Detection

The system now includes performance optimizations:

```go
// Only updates when input actually changes
func (sm *Manager) UpdateFilter(input string) {
    if input == sm.lastInput {
        return  // Skip processing if input hasn't changed
    }
    // ... rest of filtering logic
}
```

This prevents:
- Unnecessary reprocessing on every keystroke
- Index resets during navigation
- Performance degradation with large suggestion lists

## Use Cases

### 1. Icon Autocomplete (TaskModel)
```go
iconSuggestions := suggestions.NewManager(
    []string{"terminal-bash", "code", "play", "terminal-cmd"},
    3,  // Show max 3
    nil // Use default StartsWithFilter
)
```

### 2. Task Autocomplete (WorkspaceModel)
```go
taskSuggestions := suggestions.NewManager(
    []string{"frontend", "backend", "database", "worker"},
    5,  // Show max 5
    suggestions.ContainsFilter
)
```

### 3. Custom Filter
```go
taskSuggestions := suggestions.NewManager(
    commands,
    3,
    suggestions.WordBoundaryFilter, // For multi-word matching
)
```

## Design Benefits

### 1. **Reusability**
- Single component for all autocompletions
- Consistent UX across different forms
- Reduces code duplication

### 2. **Flexibility**
- Customizable filter functions
- Configurable number of visible suggestions
- Dependency injection enables testing

### 3. **Simplicity**
- Clear and direct API
- Encapsulated state management
- Separation of concerns

### 4. **Maintainability**
- Bugs fixed in one place
- New features benefit all uses
- Easier unit testing

### 5. **Performance**
- Input change detection prevents unnecessary processing
- Optimized filtering and rendering
- Memory efficient suggestion management

## Testing

```go
func TestSuggestionManager(t *testing.T) {
    sm := suggestions.NewManager(
        []string{"apple", "application", "apply"},
        2,
        nil,
    )
    
    sm.UpdateFilter("app")
    visible := sm.GetVisible()
    
    assert.Len(t, visible, 2)  // Max 2 visible
    assert.Contains(t, visible, "apple")
    assert.Contains(t, visible, "application")
    
    sm.Next()
    selected := sm.GetSelected()
    assert.Equal(t, "application", selected)
    
    // Test input change optimization
    sm.UpdateFilter("app") // Same input
    assert.Equal(t, "application", sm.GetSelected()) // Index preserved
}
```

## Future Extensions

1. **Fuzzy matching**: More intelligent search
2. **Async loading**: Asynchronously loaded suggestions
3. **Categorization**: Grouped suggestions by category
4. **History**: Remember previous selections
5. **Caching**: Cache filtered results for better performance