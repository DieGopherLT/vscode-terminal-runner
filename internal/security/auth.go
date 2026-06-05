// internal/security/auth.go
package security

import (
	"fmt"
	"os"
	"runtime"
)

// MinTokenLength is the minimum byte length required for a valid bridge auth token.
const MinTokenLength = 32

// AuthManager handles secure token management for bridge communication
type AuthManager struct {
	token string
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{}
}

// LoadTokenFromString loads and validates an auth token provided directly,
// such as the VSTR_TOKEN env var injected by the extension into its terminals.
func (am *AuthManager) LoadTokenFromString(token string) error {
	if len(token) < MinTokenLength {
		return fmt.Errorf("invalid auth token length")
	}

	am.token = token
	return nil
}

// GetAuthHeaders returns authentication headers for HTTP requests
func (am *AuthManager) GetAuthHeaders() map[string]string {
	if am.token == "" {
		return nil
	}

	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", am.token),
		"User-Agent":    "VSTR-CLI/1.0",
	}
}

// ValidateFilePermissions checks that a bridge file has secure permissions
// (owner-only access on Unix). It is stateless, hence a package-level function.
func ValidateFilePermissions(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// On Unix systems, verify only owner has access
	if runtime.GOOS != "windows" {
		mode := info.Mode().Perm()
		// Check that only owner has read/write permissions (max 0700)
		if mode&0o077 != 0 {
			return false
		}
	}

	return true
}
