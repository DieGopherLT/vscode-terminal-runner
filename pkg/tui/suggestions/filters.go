package suggestions

import (
	"os"
	"path/filepath"
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

// PathDirectoryFilter provides dynamic directory suggestions based on filesystem
// Expands ~ to home directory and filters only directories
var PathDirectoryFilter FilterFunc = func(suggestion, input string) bool {
	// This filter is handled specially by the path suggestions manager
	// The actual filtering logic is in getDirectorySuggestions
	return strings.HasPrefix(strings.ToLower(suggestion), strings.ToLower(input))
}

// GetDirectorySuggestions returns directory suggestions for the given input path
func GetDirectorySuggestions(input string) []string {
	if input == "" {
		input = "."
	}

	// Expand ~ to home directory
	expandedPath := expandPath(input)
	
	// Determine directory to search and prefix for matching
	var searchDir, matchPrefix string
	
	if strings.HasSuffix(expandedPath, string(filepath.Separator)) || expandedPath == "." || expandedPath == ".." {
		// Input ends with separator or is current/parent dir - search inside
		searchDir = expandedPath
		matchPrefix = ""
	} else {
		// Extract directory and filename prefix
		searchDir = filepath.Dir(expandedPath)
		matchPrefix = strings.ToLower(filepath.Base(expandedPath))
	}

	// Read directory contents
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return []string{}
	}

	var suggestions []string
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Skip files, only include directories
		}
		
		entryName := entry.Name()
		
		// Skip hidden directories unless input starts with dot
		if strings.HasPrefix(entryName, ".") && !strings.HasPrefix(matchPrefix, ".") {
			continue
		}
		
		// Filter by prefix if provided
		if matchPrefix != "" && !strings.HasPrefix(strings.ToLower(entryName), matchPrefix) {
			continue
		}
		
		// Build full path suggestion
		fullPath := filepath.Join(searchDir, entryName)
		
		// Convert back to display format (with ~ if applicable)
		displayPath := contractPath(fullPath)
		
		// Add trailing separator for directories
		displayPath += string(filepath.Separator)
		
		suggestions = append(suggestions, displayPath)
	}
	
	return suggestions
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~" {
			return home
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// contractPath converts absolute path back to ~ format if it's under home directory
func contractPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	
	if strings.HasPrefix(path, home) {
		relPath, err := filepath.Rel(home, path)
		if err != nil {
			return path
		}
		if relPath == "." {
			return "~"
		}
		return filepath.Join("~", relPath)
	}
	return path
}