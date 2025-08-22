// internal/security/auth.go
package security

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
)

// AuthManager handles secure token management for bridge communication
type AuthManager struct {
	token string
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{}
}

// BridgeInfo structure expected from bridge file
type BridgeInfo struct {
	Port          int    `json:"port"`
	PID           int    `json:"pid"`
	InstanceID    int64  `json:"instance_id"`
	WorkspacePath string `json:"workspace_path"`
	WorkspaceName string `json:"workspace_name"`
	Timestamp     string `json:"timestamp"`
	AuthToken     string `json:"auth_token"`
	Secure        bool   `json:"secure"`
}

// LoadTokenFromBridge loads and validates authentication token from bridge file
func (am *AuthManager) LoadTokenFromBridge(bridgeFilePath string) error {
	// 1. Validate file permissions are secure
	if !am.ValidateFilePermissions(bridgeFilePath) {
		return fmt.Errorf("bridge info file has insecure permissions")
	}
	
	// 2. Read and validate content
	bridgeInfo, err := am.readBridgeInfo(bridgeFilePath)
	if err != nil {
		return fmt.Errorf("failed to read bridge info: %w", err)
	}
	
	// 3. Validate bridge is in secure mode
	if !bridgeInfo.Secure {
		return fmt.Errorf("bridge is not running in secure mode")
	}
	
	// 4. Validate token length and format
	if len(bridgeInfo.AuthToken) < 32 {
		return fmt.Errorf("invalid auth token length")
	}
	
	am.token = bridgeInfo.AuthToken
	return nil
}

// ValidateFilePermissions checks that bridge file has secure permissions
func (am *AuthManager) ValidateFilePermissions(filePath string) bool {
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

// readBridgeInfo reads and parses bridge information from file
func (am *AuthManager) readBridgeInfo(filePath string) (*BridgeInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var bridgeInfo BridgeInfo
	if err := json.Unmarshal(data, &bridgeInfo); err != nil {
		return nil, err
	}
	
	return &bridgeInfo, nil
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