package task

import (
	"os"
	"path/filepath"
	"strings"
)

// Path Suggestions for Task Creation
//
// This file contains all path-specific autocomplete logic for task creation.
// This functionality is placed here (rather than in the generic suggestions package)
// to maintain better cohesion - this is task-specific logic that deals with filesystem
// operations and isn't a general-purpose filter that would be useful elsewhere.

// updatePathSuggestions dynamically updates path suggestions based on filesystem
func (t *TaskModel) updatePathSuggestions(input string) {
	// Only regenerate suggestions when the directory context changes
	// This prevents resetting the selection index on every keystroke
	currentDir := t.getDirectoryContext(input)

	// Check if we need to regenerate suggestions (directory changed)
	if t.shouldRegeneratePaths(currentDir) {
		dirSuggestions := getDirectorySuggestions(input)
		t.pathSuggestions.SetSuggestions(dirSuggestions)
		t.lastPathDirectory = currentDir
	}

	// Always update filter for real-time filtering
	t.pathSuggestions.UpdateFilter(input)
}

// getDirectoryContext extracts the directory portion of the input path
func (t *TaskModel) getDirectoryContext(input string) string {
	if input == "" {
		return "."
	}

	// Expand ~ to home directory for comparison
	expanded := input
	if strings.HasPrefix(input, "~/") || input == "~" {
		home, _ := os.UserHomeDir()

		if input == "~" {
			expanded = home
		} else {
			expanded = filepath.Join(home, input[2:])
		}
	}

	// If ends with separator or is current/parent dir, use as-is
	if strings.HasSuffix(expanded, string(filepath.Separator)) || expanded == "." || expanded == ".." {
		return expanded
	}

	// Otherwise return the directory part
	return filepath.Dir(expanded)
}

// shouldRegeneratePaths determines if we need to regenerate the suggestions list
func (t *TaskModel) shouldRegeneratePaths(currentDir string) bool {
	// Always regenerate on first run
	if t.lastPathDirectory == "" {
		return true
	}

	// Regenerate if directory context changed
	return t.lastPathDirectory != currentDir
}

// getDirectorySuggestions returns directory suggestions for the given input path
func getDirectorySuggestions(input string) []string {
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
	if !strings.HasPrefix(path, "~/") && path != "~" {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}

	return filepath.Join(home, path[2:])
}

// contractPath converts absolute path back to ~ format if it's under home directory
func contractPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if !strings.HasPrefix(path, home) {
		return path
	}

	relPath, err := filepath.Rel(home, path)
	if err != nil {
		return path
	}

	if relPath == "." {
		return "~"
	}

	return filepath.Join("~", relPath)
}
