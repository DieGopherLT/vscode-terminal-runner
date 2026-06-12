// internal/vscode/vscode_test.go
//
// White-box characterization and behavior tests for package vscode.
// All tests in this file are in package vscode so they can access unexported
// functions and use the fakes defined in fakes_test.go.
package vscode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils"
)

// ---------------------------------------------------------------------------
// validateBridgeStructure
// ---------------------------------------------------------------------------

func TestValidateBridgeStructure(t *testing.T) {
	tests := []struct {
		name    string
		info    BridgeInfo
		wantErr string
	}{
		{
			name:    "valid bridge passes all checks",
			info:    newValidBridgeInfo().Build(),
			wantErr: "",
		},
		{
			name:    "port zero is rejected",
			info:    newValidBridgeInfo().WithPort(0).Build(),
			wantErr: "invalid port number",
		},
		{
			name:    "port negative is rejected",
			info:    newValidBridgeInfo().WithPort(-1).Build(),
			wantErr: "invalid port number",
		},
		{
			name:    "port above max is rejected",
			info:    newValidBridgeInfo().WithPort(65536).Build(),
			wantErr: "invalid port number",
		},
		{
			name:    "port at max boundary 65535 is accepted",
			info:    newValidBridgeInfo().WithPort(65535).Build(),
			wantErr: "",
		},
		{
			name:    "port at min boundary 1 is accepted",
			info:    newValidBridgeInfo().WithPort(1).Build(),
			wantErr: "",
		},
		{
			name:    "pid zero is rejected",
			info:    newValidBridgeInfo().WithPID(0).Build(),
			wantErr: "invalid PID",
		},
		{
			name:    "pid negative is rejected",
			info:    newValidBridgeInfo().WithPID(-100).Build(),
			wantErr: "invalid PID",
		},
		{
			name:    "token shorter than minimum is rejected",
			info:    newValidBridgeInfo().WithShortToken().Build(),
			wantErr: "invalid auth token length",
		},
		{
			name:    "token exactly at minimum length is accepted",
			info:    newValidBridgeInfo().WithAuthToken(minValidToken()).Build(),
			wantErr: "",
		},
		{
			name:    "token empty is rejected",
			info:    newValidBridgeInfo().WithAuthToken("").Build(),
			wantErr: "invalid auth token length",
		},
		{
			name:    "secure false is rejected",
			info:    newValidBridgeInfo().WithSecure(false).Build(),
			wantErr: "not in secure mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := tt.info
			err := validateBridgeStructure(&info)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isVSCodeProcess
// ---------------------------------------------------------------------------

func TestIsVSCodeProcess(t *testing.T) {
	tests := []struct {
		name      string
		procName  string
		wantMatch bool
	}{
		// Degenerate
		{"empty name does not match", "", false},

		// Simple: exact VSCode process names
		{"lowercase code matches", "code", true},
		{"mixed-case Code matches", "Code", true},
		{"electron matches", "electron", true},

		// General: insiders and various platforms
		{"code-insiders matches", "code-insiders", true},
		{"ELECTRON uppercase matches", "ELECTRON", true},
		{"code helper process matches", "code helper (renderer)", true},

		// Edge: substring matches — isVSCodeProcess uses strings.Contains so any name
		// that embeds "code" or "electron" matches, even non-VSCode processes.
		// This is the current (characterization) behavior; see the labeled expected-failure
		// test below for the intended stricter behavior.
		{"barcode contains 'code' and matches (characterization)", "barcode", true},
		{"vscode contains 'code' and matches", "vscode", true},
		{"decode contains 'code' and matches", "decode", true},

		// Non-VSCode processes
		{"bash does not match", "bash", false},
		{"node does not match", "node", false},
		{"python does not match", "python", false},
		{"vim does not match", "vim", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVSCodeProcess(tt.procName)
			if got != tt.wantMatch {
				t.Errorf("isVSCodeProcess(%q) = %v, want %v", tt.procName, got, tt.wantMatch)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractWorkspacePath
// ---------------------------------------------------------------------------

func TestExtractWorkspacePath(t *testing.T) {
	// Create a real directory so the direct-path branch can confirm via os.Stat
	realDir := t.TempDir()

	tests := []struct {
		name    string
		cmdline string
		want    string
	}{
		{
			name:    "empty cmdline returns empty",
			cmdline: "",
			want:    "",
		},
		{
			name:    "no workspace flags returns empty",
			cmdline: "/usr/share/code/code --no-sandbox",
			want:    "",
		},
		{
			name:    "folder-uri with file:// prefix strips prefix",
			cmdline: "/usr/share/code/code --folder-uri file:///home/user/projects/myapp",
			want:    "/home/user/projects/myapp",
		},
		{
			name:    "folder-uri without file:// prefix is returned as-is",
			cmdline: "/usr/share/code/code --folder-uri /home/user/projects/myapp",
			want:    "/home/user/projects/myapp",
		},
		{
			name:    "folder-uri flag at end with no value returns empty",
			cmdline: "/usr/share/code/code --folder-uri",
			want:    "",
		},
		{
			name:    "direct path argument that is a real directory is returned",
			cmdline: fmt.Sprintf("/usr/share/code/code %s", realDir),
			want:    realDir,
		},
		{
			name:    "direct path that does not exist is not returned",
			cmdline: "/usr/share/code/code /no/such/directory/xyz123",
			want:    "",
		},
		{
			name:    "folder-uri takes precedence over positional path",
			cmdline: fmt.Sprintf("/usr/share/code/code --folder-uri file:///preferred %s", realDir),
			want:    "/preferred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractWorkspacePath(tt.cmdline)
			if got != tt.want {
				t.Errorf("extractWorkspacePath(%q) = %q, want %q", tt.cmdline, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// handleBridgeError
// ---------------------------------------------------------------------------

func TestHandleBridgeError(t *testing.T) {
	tests := []struct {
		name        string
		inputErrMsg string
		wantErrMsg  string
	}{
		{
			name:        "authentication failed maps to shortened message",
			inputErrMsg: "authentication failed: invalid credentials",
			wantErrMsg:  "authentication failed",
		},
		{
			name:        "rate limit exceeded maps to shortened message",
			inputErrMsg: "rate limit exceeded for client",
			wantErrMsg:  "rate limit exceeded",
		},
		{
			name:        "command blocked maps to policy message",
			inputErrMsg: "command blocked by filter",
			wantErrMsg:  "command blocked by security policy",
		},
		{
			name:        "insecure permissions maps to shortened message",
			inputErrMsg: "insecure permissions on file",
			wantErrMsg:  "insecure file permissions",
		},
		{
			name:        "not in secure mode maps to bridge message",
			inputErrMsg: "bridge not in secure mode",
			wantErrMsg:  "bridge not in secure mode",
		},
		{
			name:        "bridge directory not found maps to shortened message",
			inputErrMsg: "bridge directory not found: /tmp/vstr-bridge",
			wantErrMsg:  "bridge directory not found",
		},
		{
			name:        "no valid bridge found maps to shortened message",
			inputErrMsg: "no valid bridge found in directory",
			wantErrMsg:  "no bridge found",
		},
		{
			name:        "invalid auth token maps to shortened message",
			inputErrMsg: "invalid auth token length",
			wantErrMsg:  "invalid authentication token",
		},
		{
			name:        "connection failed maps to shortened message",
			inputErrMsg: "connection failed: dial tcp refused",
			wantErrMsg:  "connection failed",
		},
		{
			name:        "unknown error is passed through unchanged",
			inputErrMsg: "some completely unknown error occurred",
			wantErrMsg:  "some completely unknown error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputErr := errors.New(tt.inputErrMsg)
			got := handleBridgeError(inputErr)
			if got == nil {
				t.Fatalf("handleBridgeError returned nil, want error containing %q", tt.wantErrMsg)
			}
			if !strings.Contains(got.Error(), tt.wantErrMsg) {
				t.Errorf("handleBridgeError(%q).Error() = %q, want to contain %q",
					tt.inputErrMsg, got.Error(), tt.wantErrMsg)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// selectBridge
// ---------------------------------------------------------------------------

func TestSelectBridge(t *testing.T) {
	buildBridges := func(count int) []BridgeInfo {
		bridges := make([]BridgeInfo, count)
		for i := 0; i < count; i++ {
			bridges[i] = newValidBridgeInfo().
				WithPort(8000 + i).
				WithWorkspaceName(fmt.Sprintf("workspace-%d", i)).
				WithWorkspacePath(fmt.Sprintf("/home/user/project-%d", i)).
				Build()
		}
		return bridges
	}

	tests := []struct {
		name    string
		bridges []BridgeInfo
		input   string
		wantIdx int
		wantErr string
	}{
		// Degenerate: single bridge, input "1"
		{
			name:    "single bridge choice 1 returns index 0",
			bridges: buildBridges(1),
			input:   "1\n",
			wantIdx: 0,
			wantErr: "",
		},
		// Simple: two bridges, select first
		{
			name:    "two bridges choice 1 returns first",
			bridges: buildBridges(2),
			input:   "1\n",
			wantIdx: 0,
			wantErr: "",
		},
		// General: three bridges, select last
		{
			name:    "three bridges choice 3 returns last",
			bridges: buildBridges(3),
			input:   "3\n",
			wantIdx: 2,
			wantErr: "",
		},
		// Edge: boundary choices
		{
			name:    "choice equals len returns last element",
			bridges: buildBridges(2),
			input:   "2\n",
			wantIdx: 1,
			wantErr: "",
		},
		// Error: choice 0 is below range
		{
			name:    "choice 0 is out of range",
			bridges: buildBridges(2),
			input:   "0\n",
			wantErr: "invalid choice",
		},
		// Error: choice above len
		{
			name:    "choice above len is out of range",
			bridges: buildBridges(2),
			input:   "3\n",
			wantErr: "invalid choice",
		},
		// Error: non-numeric input
		{
			name:    "non-numeric input returns invalid input",
			bridges: buildBridges(2),
			input:   "abc\n",
			wantErr: "invalid input",
		},
		// Error: empty input (Fscanln returns error on EOF)
		{
			name:    "empty reader returns invalid input",
			bridges: buildBridges(2),
			input:   "",
			wantErr: "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := selectBridge(tt.bridges, reader)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wantPort := tt.bridges[tt.wantIdx].Port
			if got.Port != wantPort {
				t.Errorf("selectBridge returned port %d, want %d", got.Port, wantPort)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// detectParentVSCode
// ---------------------------------------------------------------------------

func TestDetectParentVSCode(t *testing.T) {
	tests := []struct {
		name           string
		buildInspector func() *fakeProcessInspector
		wantFound      bool
		wantPID        int32
		wantPath       string
		wantErr        string
	}{
		// Degenerate: ppid not in tree at all
		{
			name: "ppid not found in tree returns error",
			buildInspector: func() *fakeProcessInspector {
				return &fakeProcessInspector{
					ppid: 999,
					tree: map[int32]ProcessNode{},
				}
			},
			wantErr: "failed to get process",
		},
		// Simple: VSCode is the direct parent
		{
			name: "vscode is direct parent",
			buildInspector: func() *fakeProcessInspector {
				tree := buildProcessChain(
					fakeNodeSpec{pid: 200, name: "code", cmdline: "--folder-uri file:///workspace/myapp"},
				)
				return &fakeProcessInspector{ppid: 200, tree: tree}
			},
			wantFound: true,
			wantPID:   200,
			wantPath:  "/workspace/myapp",
		},
		// General: VSCode found after walking up two levels
		{
			name: "vscode found after two levels",
			buildInspector: func() *fakeProcessInspector {
				tree := buildProcessChain(
					fakeNodeSpec{pid: 100, name: "bash"},
					fakeNodeSpec{pid: 200, name: "code", cmdline: "--folder-uri file:///workspace/deeper"},
				)
				return &fakeProcessInspector{ppid: 100, tree: tree}
			},
			wantFound: true,
			wantPID:   200,
			wantPath:  "/workspace/deeper",
		},
		// General: electron variant is detected
		{
			name: "electron process is detected as vscode",
			buildInspector: func() *fakeProcessInspector {
				tree := buildProcessChain(
					fakeNodeSpec{pid: 100, name: "bash"},
					fakeNodeSpec{pid: 300, name: "electron", cmdline: "--folder-uri file:///workspace/electron-app"},
				)
				return &fakeProcessInspector{ppid: 100, tree: tree}
			},
			wantFound: true,
			wantPID:   300,
			wantPath:  "/workspace/electron-app",
		},
		// General: code-insiders variant
		{
			name: "code-insiders process is detected",
			buildInspector: func() *fakeProcessInspector {
				tree := buildProcessChain(
					fakeNodeSpec{pid: 100, name: "bash"},
					fakeNodeSpec{pid: 400, name: "code-insiders", cmdline: "--folder-uri file:///workspace/insiders"},
				)
				return &fakeProcessInspector{ppid: 100, tree: tree}
			},
			wantFound: true,
			wantPID:   400,
			wantPath:  "/workspace/insiders",
		},
		// Edge: no VSCode in tree, reaches root (nil parent)
		{
			name: "tree has no vscode process returns not found error",
			buildInspector: func() *fakeProcessInspector {
				tree := buildProcessChain(
					fakeNodeSpec{pid: 100, name: "bash"},
					fakeNodeSpec{pid: 200, name: "systemd"},
					fakeNodeSpec{pid: 300, name: "kernel"},
				)
				return &fakeProcessInspector{ppid: 100, tree: tree}
			},
			wantErr: "VSCode parent process not found",
		},
		// Edge: VSCode at exactly 10th ancestor (should still be found)
		{
			name: "vscode at 9th level is found (within 10-step limit)",
			buildInspector: func() *fakeProcessInspector {
				// Build a chain of 9 non-VSCode nodes + VSCode at level 9
				nodes := []fakeNodeSpec{
					{pid: 1, name: "bash"},
					{pid: 2, name: "sh"},
					{pid: 3, name: "sh"},
					{pid: 4, name: "sh"},
					{pid: 5, name: "sh"},
					{pid: 6, name: "sh"},
					{pid: 7, name: "sh"},
					{pid: 8, name: "sh"},
					{pid: 9, name: "sh"},
					{pid: 10, name: "code", cmdline: "--folder-uri file:///deep/workspace"},
				}
				tree := buildProcessChain(nodes...)
				return &fakeProcessInspector{ppid: 1, tree: tree}
			},
			wantFound: true,
			wantPID:   10,
			wantPath:  "/deep/workspace",
		},
		// Edge: VSCode at exactly 11th ancestor (exceeds 10-step limit)
		{
			name: "vscode at 11th level is not found (exceeds 10-step limit)",
			buildInspector: func() *fakeProcessInspector {
				nodes := []fakeNodeSpec{
					{pid: 1, name: "bash"},
					{pid: 2, name: "sh"},
					{pid: 3, name: "sh"},
					{pid: 4, name: "sh"},
					{pid: 5, name: "sh"},
					{pid: 6, name: "sh"},
					{pid: 7, name: "sh"},
					{pid: 8, name: "sh"},
					{pid: 9, name: "sh"},
					{pid: 10, name: "sh"},
					{pid: 11, name: "code", cmdline: "--folder-uri file:///too/deep"},
				}
				tree := buildProcessChain(nodes...)
				return &fakeProcessInspector{ppid: 1, tree: tree}
			},
			wantErr: "VSCode parent process not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inspector := tt.buildInspector()
			got, err := detectParentVSCode(inspector)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.PID != tt.wantPID {
				t.Errorf("PID = %d, want %d", got.PID, tt.wantPID)
			}

			if got.WorkspacePath != tt.wantPath {
				t.Errorf("WorkspacePath = %q, want %q", got.WorkspacePath, tt.wantPath)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getBridgeDirectory
// ---------------------------------------------------------------------------

func TestGetBridgeDirectory(t *testing.T) {
	t.Run("uses TMPDIR env var when set", func(t *testing.T) {
		customTmp := t.TempDir()
		t.Setenv("TMPDIR", customTmp)

		got := getBridgeDirectory()
		want := filepath.Join(customTmp, "vstr-bridge")
		if got != want {
			t.Errorf("getBridgeDirectory() = %q, want %q", got, want)
		}
	})

	t.Run("falls back to /tmp when TMPDIR is not set", func(t *testing.T) {
		t.Setenv("TMPDIR", "")

		got := getBridgeDirectory()
		want := "/tmp/vstr-bridge"
		if got != want {
			t.Errorf("getBridgeDirectory() = %q, want %q", got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// validateDirectoryPermissions
// ---------------------------------------------------------------------------

func TestValidateDirectoryPermissions(t *testing.T) {
	t.Run("secure directory with 0700 passes", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		if !validateDirectoryPermissions(dir) {
			t.Error("expected validateDirectoryPermissions to return true for 0700 dir")
		}
	})

	t.Run("insecure directory with 0755 fails", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o755); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		if validateDirectoryPermissions(dir) {
			t.Error("expected validateDirectoryPermissions to return false for 0755 dir")
		}
	})

	t.Run("insecure directory with 0777 fails", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o777); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		if validateDirectoryPermissions(dir) {
			t.Error("expected validateDirectoryPermissions to return false for 0777 dir")
		}
	})

	t.Run("nonexistent path returns false", func(t *testing.T) {
		if validateDirectoryPermissions("/no/such/path/xyz999") {
			t.Error("expected validateDirectoryPermissions to return false for nonexistent path")
		}
	})
}

// ---------------------------------------------------------------------------
// validateBridgeFile
// ---------------------------------------------------------------------------

func writeBridgeFile(t *testing.T, dir string, info BridgeInfo, perms os.FileMode) string {
	t.Helper()

	fileName := fmt.Sprintf("bridge-%d.json", info.Port)
	filePath := filepath.Join(dir, fileName)

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	if err := os.Chmod(filePath, perms); err != nil {
		t.Fatalf("os.Chmod: %v", err)
	}

	return filePath
}

func TestValidateBridgeFile(t *testing.T) {
	t.Run("valid file with correct permissions returns bridge info", func(t *testing.T) {
		dir := t.TempDir()
		info := newValidBridgeInfo().Build()
		filePath := writeBridgeFile(t, dir, info, 0o600)

		got, err := validateBridgeFile(filePath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Port != info.Port {
			t.Errorf("port = %d, want %d", got.Port, info.Port)
		}
	})

	t.Run("insecure file permissions returns error", func(t *testing.T) {
		dir := t.TempDir()
		info := newValidBridgeInfo().Build()
		filePath := writeBridgeFile(t, dir, info, 0o644)

		_, err := validateBridgeFile(filePath)
		if err == nil {
			t.Fatal("expected error for insecure permissions, got nil")
		}
		if !strings.Contains(err.Error(), "insecure file permissions") {
			t.Errorf("error %q does not contain 'insecure file permissions'", err.Error())
		}
	})

	t.Run("nonexistent file returns insecure permissions error", func(t *testing.T) {
		// ValidateFilePermissions runs first and returns false for a nonexistent file
		_, err := validateBridgeFile("/no/such/file.json")
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
		// Characterization: permission check runs before read, so the error is about permissions
		if !strings.Contains(err.Error(), "insecure file permissions") {
			t.Errorf("error %q does not contain 'insecure file permissions'", err.Error())
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "bridge-9999.json")
		if err := os.WriteFile(filePath, []byte("not valid json"), 0o600); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}

		_, err := validateBridgeFile(filePath)
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
		if !strings.Contains(err.Error(), "invalid JSON") {
			t.Errorf("error %q does not contain 'invalid JSON'", err.Error())
		}
	})

	t.Run("bridge with invalid structure returns validation error", func(t *testing.T) {
		dir := t.TempDir()
		info := newValidBridgeInfo().WithPort(0).Build()
		filePath := writeBridgeFile(t, dir, info, 0o600)

		_, err := validateBridgeFile(filePath)
		if err == nil {
			t.Fatal("expected error for invalid port, got nil")
		}
		if !strings.Contains(err.Error(), "invalid port number") {
			t.Errorf("error %q does not contain 'invalid port number'", err.Error())
		}
	})
}

// ---------------------------------------------------------------------------
// enrichWorkspaceMetadata
// ---------------------------------------------------------------------------

func TestEnrichWorkspaceMetadata(t *testing.T) {
	t.Run("enriches workspace fields from valid file", func(t *testing.T) {
		dir := t.TempDir()
		fileInfo := newValidBridgeInfo().
			WithWorkspaceName("enriched-ws").
			WithWorkspacePath("/enriched/path").
			Build()
		filePath := writeBridgeFile(t, dir, fileInfo, 0o600)

		bridge := &BridgeInfo{Port: fileInfo.Port, AuthToken: minValidToken(), Secure: true}
		enrichWorkspaceMetadata(bridge, filePath)

		if bridge.WorkspaceName != "enriched-ws" {
			t.Errorf("WorkspaceName = %q, want %q", bridge.WorkspaceName, "enriched-ws")
		}
		if bridge.WorkspacePath != "/enriched/path" {
			t.Errorf("WorkspacePath = %q, want %q", bridge.WorkspacePath, "/enriched/path")
		}
	})

	t.Run("silently ignores nonexistent file", func(t *testing.T) {
		bridge := &BridgeInfo{Port: 8765, AuthToken: minValidToken(), Secure: true}
		// Should not panic or return an error
		enrichWorkspaceMetadata(bridge, "/no/such/file.json")

		if bridge.WorkspaceName != "" {
			t.Errorf("WorkspaceName should remain empty, got %q", bridge.WorkspaceName)
		}
	})

	t.Run("silently ignores invalid JSON in file", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "bridge-bad.json")
		if err := os.WriteFile(filePath, []byte("not json"), 0o600); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}

		bridge := &BridgeInfo{Port: 8765, AuthToken: minValidToken(), Secure: true}
		enrichWorkspaceMetadata(bridge, filePath) // should not panic

		if bridge.WorkspaceName != "" {
			t.Errorf("WorkspaceName should remain empty, got %q", bridge.WorkspaceName)
		}
	})
}

// ---------------------------------------------------------------------------
// scanValidBridges
// ---------------------------------------------------------------------------

func TestScanValidBridges(t *testing.T) {
	t.Run("empty directory returns empty slice", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		bridges, err := scanValidBridges(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bridges) != 0 {
			t.Errorf("expected 0 bridges, got %d", len(bridges))
		}
	})

	t.Run("single valid bridge file is returned", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		info := newValidBridgeInfo().Build()
		writeBridgeFile(t, dir, info, 0o600)

		bridges, err := scanValidBridges(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bridges) != 1 {
			t.Fatalf("expected 1 bridge, got %d", len(bridges))
		}
		if bridges[0].Port != info.Port {
			t.Errorf("port = %d, want %d", bridges[0].Port, info.Port)
		}
	})

	t.Run("multiple valid bridge files all returned", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		info1 := newValidBridgeInfo().WithPort(8001).Build()
		info2 := newValidBridgeInfo().WithPort(8002).Build()
		writeBridgeFile(t, dir, info1, 0o600)
		writeBridgeFile(t, dir, info2, 0o600)

		bridges, err := scanValidBridges(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bridges) != 2 {
			t.Errorf("expected 2 bridges, got %d", len(bridges))
		}
	})

	t.Run("non-bridge files are ignored", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		// Write a file that does not match bridge-*.json pattern
		otherPath := filepath.Join(dir, "other-file.json")
		if err := os.WriteFile(otherPath, []byte("{}"), 0o600); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}

		bridges, err := scanValidBridges(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bridges) != 0 {
			t.Errorf("expected 0 bridges, got %d", len(bridges))
		}
	})

	t.Run("invalid bridge files are skipped", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		// Write an invalid bridge file (bad port)
		info := newValidBridgeInfo().WithPort(0).Build()
		writeBridgeFile(t, dir, info, 0o600)

		bridges, err := scanValidBridges(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bridges) != 0 {
			t.Errorf("expected 0 bridges after skipping invalid ones, got %d", len(bridges))
		}
	})

	t.Run("mix of valid and invalid files returns only valid ones", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		valid := newValidBridgeInfo().WithPort(8003).Build()
		invalid := newValidBridgeInfo().WithPort(8004).WithSecure(false).Build()
		writeBridgeFile(t, dir, valid, 0o600)
		writeBridgeFile(t, dir, invalid, 0o600)

		bridges, err := scanValidBridges(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bridges) != 1 {
			t.Fatalf("expected 1 valid bridge, got %d", len(bridges))
		}
		if bridges[0].Port != valid.Port {
			t.Errorf("port = %d, want %d", bridges[0].Port, valid.Port)
		}
	})
}

// ---------------------------------------------------------------------------
// findValidBridgeByWorkspace
// ---------------------------------------------------------------------------

func TestFindValidBridgeByWorkspace(t *testing.T) {
	t.Run("finds bridge by exact workspace path", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		info := newValidBridgeInfo().WithWorkspacePath("/home/user/myapp").Build()
		writeBridgeFile(t, dir, info, 0o600)

		got, err := findValidBridgeByWorkspace(dir, "/home/user/myapp")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Port != info.Port {
			t.Errorf("port = %d, want %d", got.Port, info.Port)
		}
	})

	t.Run("finds bridge by partial workspace path match", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		info := newValidBridgeInfo().WithWorkspacePath("/home/user/projects/myapp").Build()
		writeBridgeFile(t, dir, info, 0o600)

		got, err := findValidBridgeByWorkspace(dir, "myapp")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Port != info.Port {
			t.Errorf("port = %d, want %d", got.Port, info.Port)
		}
	})

	t.Run("no matching bridge returns error", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		info := newValidBridgeInfo().WithWorkspacePath("/home/user/myapp").Build()
		writeBridgeFile(t, dir, info, 0o600)

		_, err := findValidBridgeByWorkspace(dir, "/different/workspace")
		if err == nil {
			t.Fatal("expected error for no matching bridge, got nil")
		}
		if !strings.Contains(err.Error(), "no bridge found for workspace") {
			t.Errorf("error %q does not contain 'no bridge found for workspace'", err.Error())
		}
	})

	t.Run("empty directory returns error", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		_, err := findValidBridgeByWorkspace(dir, "/any/path")
		if err == nil {
			t.Fatal("expected error for empty directory, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// discoverFromEnv
// ---------------------------------------------------------------------------

func TestDiscoverFromEnv(t *testing.T) {
	// All tests manipulate env vars so they must remain serial (no t.Parallel)

	t.Run("VSTR not set returns error", func(t *testing.T) {
		t.Setenv("VSTR", "")

		_, err := discoverFromEnv("/any/dir")
		if err == nil {
			t.Fatal("expected error when VSTR is not set")
		}
	})

	t.Run("VSTR with invalid port returns error", func(t *testing.T) {
		t.Setenv("VSTR", "not-a-port")

		_, err := discoverFromEnv("/any/dir")
		if err == nil {
			t.Fatal("expected error for invalid VSTR value")
		}
		if !strings.Contains(err.Error(), "invalid VSTR port") {
			t.Errorf("error %q does not contain 'invalid VSTR port'", err.Error())
		}
	})

	t.Run("valid VSTR and VSTR_TOKEN returns bridge without file read", func(t *testing.T) {
		t.Setenv("VSTR", "8765")
		t.Setenv("VSTR_TOKEN", minValidToken())

		got, err := discoverFromEnv("/nonexistent/dir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Port != 8765 {
			t.Errorf("port = %d, want 8765", got.Port)
		}
		if got.AuthToken != minValidToken() {
			t.Errorf("AuthToken = %q, want minValidToken()", got.AuthToken)
		}
		if !got.Secure {
			t.Error("expected Secure = true when using VSTR_TOKEN")
		}
	})

	t.Run("VSTR set but VSTR_TOKEN too short falls back to file validation", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		info := newValidBridgeInfo().WithPort(8770).Build()
		writeBridgeFile(t, dir, info, 0o600)

		t.Setenv("VSTR", "8770")
		t.Setenv("VSTR_TOKEN", "short")

		got, err := discoverFromEnv(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Port != 8770 {
			t.Errorf("port = %d, want 8770", got.Port)
		}
	})

	t.Run("VSTR set, no VSTR_TOKEN, and file does not exist returns error", func(t *testing.T) {
		t.Setenv("VSTR", "9999")
		t.Setenv("VSTR_TOKEN", "")

		_, err := discoverFromEnv("/nonexistent/dir")
		if err == nil {
			t.Fatal("expected error when bridge file does not exist")
		}
	})

	t.Run("VSTR_TOKEN enriches workspace metadata from bridge file when present", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		info := newValidBridgeInfo().
			WithPort(8780).
			WithWorkspaceName("test-workspace").
			WithWorkspacePath("/test/path").
			Build()
		writeBridgeFile(t, dir, info, 0o600)

		t.Setenv("VSTR", "8780")
		t.Setenv("VSTR_TOKEN", minValidToken())

		got, err := discoverFromEnv(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Workspace metadata should be enriched from the file
		if got.WorkspaceName != "test-workspace" {
			t.Errorf("WorkspaceName = %q, want 'test-workspace'", got.WorkspaceName)
		}
	})
}

// ---------------------------------------------------------------------------
// discoverFromScan
// ---------------------------------------------------------------------------

func TestDiscoverFromScan(t *testing.T) {
	t.Run("empty directory returns no valid bridge error", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}

		_, err := discoverFromScan(dir)
		if err == nil {
			t.Fatal("expected error for empty bridge directory")
		}
		if !strings.Contains(err.Error(), "no valid bridge found") {
			t.Errorf("error %q does not contain 'no valid bridge found'", err.Error())
		}
	})

	t.Run("single valid bridge is returned without prompting", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		info := newValidBridgeInfo().WithPort(8890).Build()
		writeBridgeFile(t, dir, info, 0o600)

		got, err := discoverFromScan(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Port != 8890 {
			t.Errorf("port = %d, want 8890", got.Port)
		}
	})
}

// ---------------------------------------------------------------------------
// Runner.RunTask
// ---------------------------------------------------------------------------

func TestRunnerRunTask(t *testing.T) {
	tests := []struct {
		name         string
		taskName     string
		taskRepo     *fakeTaskRepository
		client       *fakeBridgeClient
		wantErr      string
		wantLastTask bool
	}{
		// Degenerate: task not found in repo
		{
			name:     "task not found returns wrapped error",
			taskName: "missing-task",
			taskRepo: &fakeTaskRepository{},
			client:   &fakeBridgeClient{},
			wantErr:  "task not found",
		},
		// Simple: task found, execution succeeds
		{
			name:     "found task is executed successfully",
			taskName: "build",
			taskRepo: &fakeTaskRepository{
				task: func() *models.Task {
					t := testutils.NewTask().WithName("build").Build()
					return &t
				}(),
			},
			client:       &fakeBridgeClient{},
			wantLastTask: true,
		},
		// Error: repo returns error
		{
			name:     "repo error is wrapped",
			taskName: "any",
			taskRepo: &fakeTaskRepository{returnErr: errors.New("disk read failed")},
			client:   &fakeBridgeClient{},
			wantErr:  "task not found",
		},
		// Error: bridge client returns error
		{
			name:     "bridge execution error is handled",
			taskName: "build",
			taskRepo: &fakeTaskRepository{
				task: func() *models.Task {
					t := testutils.NewTask().WithName("build").Build()
					return &t
				}(),
			},
			client:  &fakeBridgeClient{executeTaskErr: errors.New("connection failed: refused")},
			wantErr: "connection failed",
		},
		// General: multi-command task exercises the display loop
		{
			name:     "task with multiple commands is executed",
			taskName: "deploy",
			taskRepo: &fakeTaskRepository{
				task: func() *models.Task {
					t := testutils.NewTask().WithName("deploy").WithCmds("go build", "go test", "docker push").Build()
					return &t
				}(),
			},
			client:       &fakeBridgeClient{},
			wantLastTask: true,
		},
		// Edge: task with empty commands slice
		{
			name:     "task with no commands is executed",
			taskName: "empty",
			taskRepo: &fakeTaskRepository{
				task: func() *models.Task {
					t := testutils.NewTask().WithName("empty").WithCmds().Build()
					return &t
				}(),
			},
			client:       &fakeBridgeClient{},
			wantLastTask: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewRunnerWithDeps(RunnerDeps{
				Client:     tt.client,
				Tasks:      tt.taskRepo,
				Workspaces: &fakeWorkspaceRepository{},
			})

			err := runner.RunTask(tt.taskName)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantLastTask && tt.client.lastTask == nil {
				t.Error("expected client.ExecuteTask to be called, but lastTask is nil")
			}

			if tt.wantLastTask && tt.client.lastTask.Name != tt.taskName {
				t.Errorf("lastTask.Name = %q, want %q", tt.client.lastTask.Name, tt.taskName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Runner.RunWorkspace
// ---------------------------------------------------------------------------

func TestRunnerRunWorkspace(t *testing.T) {
	tests := []struct {
		name              string
		workspaceName     string
		workspaceRepo     *fakeWorkspaceRepository
		client            *fakeBridgeClient
		wantErr           string
		wantLastWorkspace bool
	}{
		// Degenerate: workspace not found in repo
		{
			name:          "workspace not found returns error",
			workspaceName: "missing",
			workspaceRepo: &fakeWorkspaceRepository{},
			client:        &fakeBridgeClient{},
			wantErr:       "workspace not found",
		},
		// Degenerate: workspace found but has no tasks
		{
			name:          "workspace with no tasks returns error",
			workspaceName: "empty-ws",
			workspaceRepo: &fakeWorkspaceRepository{
				workspace: func() *models.Workspace {
					w := testutils.NewWorkspace().WithName("empty-ws").WithNoTasks().Build()
					return &w
				}(),
			},
			client:  &fakeBridgeClient{},
			wantErr: "no tasks found in workspace",
		},
		// Simple: workspace found, execution succeeds
		{
			name:          "workspace with tasks is executed successfully",
			workspaceName: "dev",
			workspaceRepo: &fakeWorkspaceRepository{
				workspace: func() *models.Workspace {
					task := testutils.NewTask().WithName("server").Build()
					w := testutils.NewWorkspace().WithName("dev").WithTasks(task).Build()
					return &w
				}(),
			},
			client:            &fakeBridgeClient{},
			wantLastWorkspace: true,
		},
		// General: workspace with multiple tasks
		{
			name:          "workspace with multiple tasks is executed",
			workspaceName: "production",
			workspaceRepo: &fakeWorkspaceRepository{
				workspace: func() *models.Workspace {
					t1 := testutils.NewTask().WithName("api").Build()
					t2 := testutils.NewTask().WithName("worker").Build()
					t3 := testutils.NewTask().WithName("scheduler").Build()
					w := testutils.NewWorkspace().WithName("production").WithTasks(t1, t2, t3).Build()
					return &w
				}(),
			},
			client:            &fakeBridgeClient{},
			wantLastWorkspace: true,
		},
		// Error: repo returns error
		{
			name:          "repo error is wrapped",
			workspaceName: "any",
			workspaceRepo: &fakeWorkspaceRepository{returnErr: errors.New("disk error")},
			client:        &fakeBridgeClient{},
			wantErr:       "workspace not found",
		},
		// Error: bridge client returns error
		{
			name:          "bridge execution error is handled",
			workspaceName: "dev",
			workspaceRepo: &fakeWorkspaceRepository{
				workspace: func() *models.Workspace {
					task := testutils.NewTask().WithName("server").Build()
					w := testutils.NewWorkspace().WithName("dev").WithTasks(task).Build()
					return &w
				}(),
			},
			client:  &fakeBridgeClient{executeWorkspaceErr: errors.New("authentication failed: expired")},
			wantErr: "authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewRunnerWithDeps(RunnerDeps{
				Client:     tt.client,
				Tasks:      &fakeTaskRepository{},
				Workspaces: tt.workspaceRepo,
			})

			err := runner.RunWorkspace(tt.workspaceName)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantLastWorkspace && tt.client.lastWorkspace == nil {
				t.Error("expected client.ExecuteWorkspace to be called, but lastWorkspace is nil")
			}

			if tt.wantLastWorkspace && tt.client.lastWorkspace.Name != tt.workspaceName {
				t.Errorf("lastWorkspace.Name = %q, want %q", tt.client.lastWorkspace.Name, tt.workspaceName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DiscoverBridge — env path only (no process-tree or file-scan I/O)
// ---------------------------------------------------------------------------

func TestDiscoverBridge_EnvPath(t *testing.T) {
	// This test exercises DiscoverBridge entirely via the env path:
	// VSTR + VSTR_TOKEN -> discoverFromEnv succeeds before any process-tree or scan I/O.
	// We point TMPDIR at a controlled dir to satisfy the bridge directory existence and permission checks.

	t.Run("env path succeeds when bridge dir exists with correct permissions", func(t *testing.T) {
		bridgeDir := filepath.Join(t.TempDir(), "vstr-bridge")
		if err := os.MkdirAll(bridgeDir, 0o700); err != nil {
			t.Fatalf("os.MkdirAll: %v", err)
		}

		parentDir := filepath.Dir(bridgeDir)
		t.Setenv("TMPDIR", parentDir)
		t.Setenv("VSTR", "8765")
		t.Setenv("VSTR_TOKEN", minValidToken())

		got, err := DiscoverBridge()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Port != 8765 {
			t.Errorf("port = %d, want 8765", got.Port)
		}
	})

	t.Run("missing bridge dir returns error", func(t *testing.T) {
		parent := t.TempDir()
		// Do not create the vstr-bridge subdir
		t.Setenv("TMPDIR", parent)
		t.Setenv("VSTR", "8765")
		t.Setenv("VSTR_TOKEN", minValidToken())

		_, err := DiscoverBridge()
		if err == nil {
			t.Fatal("expected error when bridge directory does not exist")
		}
		if !strings.Contains(err.Error(), "bridge directory not found") {
			t.Errorf("error %q does not contain 'bridge directory not found'", err.Error())
		}
	})

	t.Run("insecure bridge dir permissions return error", func(t *testing.T) {
		bridgeDir := filepath.Join(t.TempDir(), "vstr-bridge")
		if err := os.MkdirAll(bridgeDir, 0o755); err != nil {
			t.Fatalf("os.MkdirAll: %v", err)
		}

		parentDir := filepath.Dir(bridgeDir)
		t.Setenv("TMPDIR", parentDir)
		t.Setenv("VSTR", "8765")
		t.Setenv("VSTR_TOKEN", minValidToken())

		_, err := DiscoverBridge()
		if err == nil {
			t.Fatal("expected error for insecure bridge directory permissions")
		}
		if !strings.Contains(err.Error(), "insecure permissions") {
			t.Errorf("error %q does not contain 'insecure permissions'", err.Error())
		}
	})
}
