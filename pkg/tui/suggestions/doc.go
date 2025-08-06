// Package suggestions provides autocomplete functionality for TUI forms.
//
// The main component is Manager, which handles filtering, navigation,
// and rendering of suggestions. It supports predefined filter functions
// for common use cases.
//
// Example usage:
//
//	iconSuggestions := suggestions.NewManager(
//		[]string{"terminal-bash", "code", "play"},
//		3,  // max visible
//		suggestions.StartsWithFilter,
//	)
//
//	// In your Update method:
//	iconSuggestions.Next()  // Navigate to next suggestion
//	iconSuggestions.ApplySelected(&textInput)  // Apply selection
//
//	// In your View method:
//	if iconSuggestions.ShouldShow(input.Value()) {
//		return iconSuggestions.Render()
//	}
package suggestions