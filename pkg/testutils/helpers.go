// pkg/testutils/helpers.go
package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ContainsString checks if a string contains a substring
func ContainsString(s, substr string) bool {
	return strings.Contains(s, substr)
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
