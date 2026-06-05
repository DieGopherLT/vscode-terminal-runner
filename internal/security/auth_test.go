package security

import (
	"os"
	"testing"

	"github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils"
)

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

func TestValidateFilePermissions(t *testing.T) {
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

			got := ValidateFilePermissions(tempFile)

			if got != tt.want {
				t.Errorf("ValidateFilePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}
