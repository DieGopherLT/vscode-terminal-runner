package suggestions

import (
	"strings"
)

// Predefined filter functions for common use cases

// StartsWithFilter filters suggestions that start with the input (case-insensitive)
// This is the default filter used by NewManager when nil is passed
var StartsWithFilter FilterFunc = func(suggestion, input string) bool {
	return strings.HasPrefix(strings.ToLower(suggestion), strings.ToLower(input))
}

// ContainsFilter filters suggestions that contain the input anywhere (case-insensitive)
var ContainsFilter FilterFunc = func(suggestion, input string) bool {
	return strings.Contains(strings.ToLower(suggestion), strings.ToLower(input))
}

// ExactFilter filters suggestions that exactly match the input (case-insensitive)
var ExactFilter FilterFunc = func(suggestion, input string) bool {
	return strings.EqualFold(suggestion, input)
}

// WordBoundaryFilter filters suggestions where any word starts with the input
// Useful for multi-word suggestions like "terminal-bash" matching "bash"
var WordBoundaryFilter FilterFunc = func(suggestion, input string) bool {
	// Split by common separators: space, dash, underscore
	separators := []string{" ", "-", "_", "."}
	words := []string{suggestion}
	
	// Split by each separator
	for _, sep := range separators {
		var newWords []string
		for _, word := range words {
			newWords = append(newWords, strings.Split(word, sep)...)
		}
		words = newWords
	}
	
	// Check if any word starts with input
	inputLower := strings.ToLower(input)
	for _, word := range words {
		if strings.HasPrefix(strings.ToLower(word), inputLower) {
			return true
		}
	}
	return false
}

// CaseSensitiveStartsWithFilter filters suggestions that start with the input (case-sensitive)
var CaseSensitiveStartsWithFilter FilterFunc = func(suggestion, input string) bool {
	return strings.HasPrefix(suggestion, input)
}

// CaseSensitiveContainsFilter filters suggestions that contain the input (case-sensitive)
var CaseSensitiveContainsFilter FilterFunc = func(suggestion, input string) bool {
	return strings.Contains(suggestion, input)
}