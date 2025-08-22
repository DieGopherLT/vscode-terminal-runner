// pkg/testutils/helpers.go
package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ContainsString checks if a string contains a substring
func ContainsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 findSubstring(s, substr))))
}

// findSubstring searches for a substring within a string
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CreateTestJSONFile creates a temporary JSON file with the given data and permissions
func CreateTestJSONFile(data interface{}, permissions os.FileMode) (string, error) {
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "test-file.json")

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(tempFile, jsonData, permissions); err != nil {
		return "", err
	}

	return tempFile, nil
}

// CreateTempFileWithPermissions creates a temporary file with specific permissions
func CreateTempFileWithPermissions(permissions os.FileMode) (string, error) {
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "test-file")

	if err := os.WriteFile(tempFile, []byte("test"), permissions); err != nil {
		return "", err
	}

	return tempFile, nil
}

// CreateTempDirWithPermissions creates a temporary directory with specific permissions
func CreateTempDirWithPermissions(permissions os.FileMode) (string, error) {
	tempDir := filepath.Join(os.TempDir(), "test-dir")
	
	if err := os.MkdirAll(tempDir, permissions); err != nil {
		return "", err
	}

	return tempDir, nil
}