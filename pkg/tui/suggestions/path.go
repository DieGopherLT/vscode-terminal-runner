package suggestions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/samber/lo"
)

// PathManager handles autocomplete suggestions for filesystem paths with dynamic directory scanning
type PathManager struct {
	*Manager
	lastDirectory string // Last directory context to avoid unnecessary rescanning
}

// NewPathManager creates a new path suggestion manager with filesystem-aware autocomplete
func NewPathManager(maxVisible int) *PathManager {
	// Use a custom filter that's optimized for paths but start with empty suggestions
	baseManager := NewManagerWithOptions([]string{}, maxVisible, StartsWithFilter, false)

	return &PathManager{
		Manager:       baseManager,
		lastDirectory: "",
	}
}

// UpdateFilter overrides the base UpdateFilter to provide dynamic path suggestions
func (pm *PathManager) UpdateFilter(inputText string) {
	// Only process if input actually changed
	if inputText == pm.lastInput {
		return
	}

	// Get current directory context from input
	currentDirectory := pm.getDirectoryContext(inputText)

	// Only regenerate suggestions when directory context changes
	if pm.shouldRegeneratePathSuggestions(currentDirectory) {
		pathSuggestions := pm.generatePathSuggestions(inputText)
		pm.SetSuggestions(pathSuggestions)
		pm.lastDirectory = currentDirectory
	}

	// Update the last input and filter the suggestions manually
	pm.lastInput = inputText

	if inputText == "" {
		pm.filteredSuggestions = pm.allSuggestions
	} else {
		pm.filteredSuggestions = lo.Filter(pm.allSuggestions, func(suggestion string, _ int) bool {
			return pm.filterFunc(suggestion, inputText)
		})
	}

	// Reset selection when filter changes
	pm.selectedIndex = 0
}

// getDirectoryContext extracts the directory portion of the input path for context comparison
func (pm *PathManager) getDirectoryContext(inputText string) string {
	if inputText == "" {
		return "."
	}

	// Expand ~ to home directory for consistent comparison
	expandedPath := pm.expandPath(inputText)

	// If path ends with separator or is a special directory, use as-is
	if strings.HasSuffix(expandedPath, string(filepath.Separator)) || expandedPath == "." || expandedPath == ".." {
		return expandedPath
	}

	// Otherwise return the directory part
	return filepath.Dir(expandedPath)
}

// shouldRegeneratePathSuggestions determines if filesystem scanning is needed
func (pm *PathManager) shouldRegeneratePathSuggestions(currentDirectory string) bool {
	// Always regenerate on first run
	if pm.lastDirectory == "" {
		return true
	}

	// Regenerate if directory context changed
	return pm.lastDirectory != currentDirectory
}

// generatePathSuggestions creates directory suggestions based on filesystem contents
func (pm *PathManager) generatePathSuggestions(inputText string) []string {
	if inputText == "" {
		inputText = "."
	}

	// Expand ~ to home directory
	expandedPath := pm.expandPath(inputText)

	// Determine search directory and match prefix
	var searchDirectory, matchPrefix string

	if strings.HasSuffix(expandedPath, string(filepath.Separator)) || expandedPath == "." || expandedPath == ".." {
		// Input ends with separator or is current/parent dir - search inside
		searchDirectory = expandedPath
		matchPrefix = ""
	} else {
		// Extract directory and filename prefix for filtering
		searchDirectory = filepath.Dir(expandedPath)
		matchPrefix = strings.ToLower(filepath.Base(expandedPath))
	}

	// Read directory contents
	directoryEntries, err := os.ReadDir(searchDirectory)
	if err != nil {
		return []string{}
	}

	var pathSuggestions []string

	for _, entry := range directoryEntries {
		if !entry.IsDir() {
			continue // Only include directories
		}

		entryName := entry.Name()

		// Skip hidden directories unless user explicitly types dot
		if strings.HasPrefix(entryName, ".") && !strings.HasPrefix(matchPrefix, ".") {
			continue
		}

		// Filter by prefix if provided
		if matchPrefix != "" && !strings.HasPrefix(strings.ToLower(entryName), matchPrefix) {
			continue
		}

		// Build full path suggestion
		fullSuggestionPath := filepath.Join(searchDirectory, entryName)

		// Convert back to display format (with ~ if applicable)
		displayPath := pm.contractPath(fullSuggestionPath)

		// Add trailing separator to indicate directory
		displayPath += string(filepath.Separator)

		pathSuggestions = append(pathSuggestions, displayPath)
	}

	return pathSuggestions
}

// expandPath expands ~ to home directory for filesystem operations
func (pm *PathManager) expandPath(path string) string {
	if !strings.HasPrefix(path, "~/") && path != "~" {
		return path
	}

	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return homeDirectory
	}

	return filepath.Join(homeDirectory, path[2:])
}

// ApplySelected applies the selected path suggestion to the textinput with path-specific handling
func (pm *PathManager) ApplySelected(input *textinput.Model) {
	selectedPath := pm.GetSelected()
	if selectedPath == "" {
		return
	}

	input.SetValue(selectedPath)
	input.SetCursor(len(selectedPath))
	pm.Reset()

	// After applying a path, update suggestions for the new context
	// This ensures that if user continues typing, suggestions remain relevant
	pm.UpdateFilter(selectedPath)
}

// contractPath converts absolute path back to ~ format if under home directory
func (pm *PathManager) contractPath(path string) string {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if !strings.HasPrefix(path, homeDirectory) {
		return path
	}

	relativePath, err := filepath.Rel(homeDirectory, path)
	if err != nil {
		return path
	}

	if relativePath == "." {
		return "~"
	}

	return filepath.Join("~", relativePath)
}
