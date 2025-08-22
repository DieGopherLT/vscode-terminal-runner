package security

import (
	"os"
	"testing"

	"github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils"
)

func TestAuthManager_LoadTokenFromBridge(t *testing.T) {
	tests := []struct {
		name        string
		bridgeInfo  BridgeInfo
		permissions os.FileMode
		wantErr     bool
		errContains string
	}{
		{
			name: "valid secure bridge",
			bridgeInfo: BridgeInfo{
				Port:          8080,
				PID:           1234,
				InstanceID:    12345,
				WorkspacePath: "/test/workspace",
				WorkspaceName: "test-workspace",
				Timestamp:     "2023-01-01T00:00:00Z",
				AuthToken:     "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				Secure:        true,
			},
			permissions: 0600,
			wantErr:     false,
		},
		{
			name: "insecure permissions",
			bridgeInfo: BridgeInfo{
				Port:          8080,
				PID:           1234,
				InstanceID:    12345,
				WorkspacePath: "/test/workspace",
				WorkspaceName: "test-workspace",
				Timestamp:     "2023-01-01T00:00:00Z",
				AuthToken:     "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				Secure:        true,
			},
			permissions: 0644,
			wantErr:     true,
			errContains: "insecure permissions",
		},
		{
			name: "bridge not in secure mode",
			bridgeInfo: BridgeInfo{
				Port:          8080,
				PID:           1234,
				InstanceID:    12345,
				WorkspacePath: "/test/workspace",
				WorkspaceName: "test-workspace",
				Timestamp:     "2023-01-01T00:00:00Z",
				AuthToken:     "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				Secure:        false,
			},
			permissions: 0600,
			wantErr:     true,
			errContains: "not running in secure mode",
		},
		{
			name: "invalid token length",
			bridgeInfo: BridgeInfo{
				Port:          8080,
				PID:           1234,
				InstanceID:    12345,
				WorkspacePath: "/test/workspace",
				WorkspaceName: "test-workspace",
				Timestamp:     "2023-01-01T00:00:00Z",
				AuthToken:     "short",
				Secure:        true,
			},
			permissions: 0600,
			wantErr:     true,
			errContains: "invalid auth token length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary bridge file
			tempFile, err := testutils.CreateTestJSONFile(tt.bridgeInfo, tt.permissions)
			if err != nil {
				t.Fatalf("Failed to create test bridge file: %v", err)
			}
			defer os.Remove(tempFile)

			// Test auth manager
			am := NewAuthManager()
			err = am.LoadTokenFromBridge(tempFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadTokenFromBridge() expected error but got none")
					return
				}
				if tt.errContains != "" && !testutils.ContainsString(err.Error(), tt.errContains) {
					t.Errorf("LoadTokenFromBridge() error = %v, should contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("LoadTokenFromBridge() unexpected error = %v", err)
					return
				}
				if am.token != tt.bridgeInfo.AuthToken {
					t.Errorf("LoadTokenFromBridge() token = %v, want %v", am.token, tt.bridgeInfo.AuthToken)
				}
			}
		})
	}
}

func TestAuthManager_GetAuthHeaders(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantNil   bool
		wantToken string
	}{
		{
			name:      "valid token",
			token:     "valid-token-123",
			wantNil:   false,
			wantToken: "Bearer valid-token-123",
		},
		{
			name:    "empty token",
			token:   "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am := &AuthManager{token: tt.token}
			headers := am.GetAuthHeaders()

			if tt.wantNil {
				if headers != nil {
					t.Errorf("GetAuthHeaders() = %v, want nil", headers)
				}
				return
			}

			if headers == nil {
				t.Errorf("GetAuthHeaders() = nil, want headers")
				return
			}

			if auth := headers["Authorization"]; auth != tt.wantToken {
				t.Errorf("GetAuthHeaders() Authorization = %v, want %v", auth, tt.wantToken)
			}

			if userAgent := headers["User-Agent"]; userAgent != "VSTR-CLI/1.0" {
				t.Errorf("GetAuthHeaders() User-Agent = %v, want VSTR-CLI/1.0", userAgent)
			}
		})
	}
}

func TestAuthManager_ValidateFilePermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions os.FileMode
		want        bool
	}{
		{"secure permissions 0600", 0600, true},
		{"secure permissions 0700", 0700, true},
		{"insecure permissions 0644", 0644, false},
		{"insecure permissions 0755", 0755, false},
		{"insecure permissions 0666", 0666, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with specific permissions
			tempFile, err := testutils.CreateTempFileWithPermissions(tt.permissions)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile)

			am := NewAuthManager()
			got := am.ValidateFilePermissions(tempFile)

			if got != tt.want {
				t.Errorf("ValidateFilePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}