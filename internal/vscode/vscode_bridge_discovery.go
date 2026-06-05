package vscode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/security"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/samber/lo"
	"github.com/shirou/gopsutil/v3/process"
)

type BridgeInfo struct {
	Port          int       `json:"port"`
	PID           int       `json:"pid"`
	InstanceID    int64     `json:"instance_id"`
	WorkspacePath string    `json:"workspace_path"`
	WorkspaceName string    `json:"workspace_name"`
	Timestamp     time.Time `json:"timestamp"`
	AuthToken     string    `json:"auth_token"`
	Secure        bool      `json:"secure"`
}

// DiscoverBridge finds the correct bridge instance, preferring the most precise
// signal available: the per-window VSTR env var, then the parent VSCode process,
// then a scan of the bridge directory. The env path trusts the window-scoped
// VSTR_TOKEN directly when present; the process-tree and scan paths validate each
// candidate file via validateBridgeFile.
func DiscoverBridge() (*BridgeInfo, error) {
	bridgeDir := getBridgeDirectory()

	if _, err := os.Stat(bridgeDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("bridge directory not found: %s", bridgeDir)
	}

	if !validateDirectoryPermissions(bridgeDir) {
		return nil, fmt.Errorf("bridge directory has insecure permissions")
	}

	// 1. VSTR env var: the per-window signal the extension injects into its terminals.
	if bridge, err := discoverFromEnv(bridgeDir); err == nil {
		return bridge, nil
	}

	// 2. Parent VSCode process: match the running window by its workspace path.
	if bridge, err := discoverFromParentProcess(bridgeDir); err == nil {
		return bridge, nil
	}

	// 3. Scan all valid bridges; auto-select when one, prompt when several.
	return discoverFromScan(bridgeDir)
}

// discoverFromEnv resolves the bridge from the VSTR/VSTR_TOKEN env vars the extension
// injects into its integrated terminals. Both are window-scoped, so they unambiguously
// identify the bridge for the current terminal. The window-scoped VSTR_TOKEN is the
// primary credential and authenticates without touching /tmp; the bridge file is read
// best-effort only to enrich display metadata, and serves as the fallback when no token
// is present in the environment.
func discoverFromEnv(bridgeDir string) (*BridgeInfo, error) {
	portStr := os.Getenv("VSTR")
	if portStr == "" {
		return nil, fmt.Errorf("VSTR env var not set")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid VSTR port %q: %w", portStr, err)
	}

	filePath := filepath.Join(bridgeDir, fmt.Sprintf("bridge-%d.json", port))

	if token := os.Getenv("VSTR_TOKEN"); len(token) >= security.MinTokenLength {
		bridge := &BridgeInfo{Port: port, AuthToken: token, Secure: true}
		enrichWorkspaceMetadata(bridge, filePath)
		return bridge, nil
	}

	return validateBridgeFile(filePath)
}

// enrichWorkspaceMetadata fills display-only fields from the bridge file when it is
// readable. Discovery already succeeded via the env token, so any failure is ignored.
func enrichWorkspaceMetadata(bridge *BridgeInfo, filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var fromFile BridgeInfo
	if err := json.Unmarshal(data, &fromFile); err != nil {
		return
	}

	bridge.PID = fromFile.PID
	bridge.InstanceID = fromFile.InstanceID
	bridge.WorkspacePath = fromFile.WorkspacePath
	bridge.WorkspaceName = fromFile.WorkspaceName
}

// discoverFromParentProcess walks the process tree to the parent VSCode window and
// matches a validated bridge by its workspace path.
func discoverFromParentProcess(bridgeDir string) (*BridgeInfo, error) {
	instance, err := detectParentVSCode()
	if err != nil {
		return nil, err
	}

	return findValidBridgeByWorkspace(bridgeDir, instance.WorkspacePath)
}

// discoverFromScan validates every bridge file in the directory and resolves a single
// target: the only valid one, or a user choice when several are present.
func discoverFromScan(bridgeDir string) (*BridgeInfo, error) {
	bridges, err := scanValidBridges(bridgeDir)
	if err != nil {
		return nil, err
	}

	if len(bridges) == 0 {
		return nil, fmt.Errorf("no valid bridge found")
	}

	if len(bridges) == 1 {
		return &bridges[0], nil
	}

	return selectBridge(bridges)
}

// scanValidBridges returns every bridge file that passes security validation. Invalid
// files are reported and skipped; stale files are left for the extension to clean up.
func scanValidBridges(bridgeDir string) ([]BridgeInfo, error) {
	files, err := os.ReadDir(bridgeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read bridge directory: %w", err)
	}

	var bridges []BridgeInfo
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "bridge-") || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(bridgeDir, file.Name())
		bridgeInfo, err := validateBridgeFile(filePath)
		if err != nil {
			styles.PrintError(fmt.Sprintf("Skipping invalid bridge file %s: %v", file.Name(), err))
			continue
		}

		bridges = append(bridges, *bridgeInfo)
	}

	return bridges, nil
}

// findValidBridgeByWorkspace returns the validated bridge whose workspace path matches.
func findValidBridgeByWorkspace(bridgeDir, path string) (*BridgeInfo, error) {
	bridges, err := scanValidBridges(bridgeDir)
	if err != nil {
		return nil, err
	}

	bridge, found := lo.Find(bridges, func(b BridgeInfo) bool {
		return b.WorkspacePath == path || strings.Contains(b.WorkspacePath, path)
	})

	if !found {
		return nil, fmt.Errorf("no bridge found for workspace %s", path)
	}

	return &bridge, nil
}

// selectBridge presents a selection menu for multiple bridges
func selectBridge(bridges []BridgeInfo) (*BridgeInfo, error) {
	styles.PrintInfo("\nMultiple VSCode instances detected")
	fmt.Println()

	for i, bridge := range bridges {
		fmt.Printf("%d. %s (PID %d)\n",
			i+1,
			styles.RunnerTaskNameStyle.Render(bridge.WorkspaceName),
			bridge.PID)
		fmt.Printf("   Path: %s\n", bridge.WorkspacePath)
	}

	fmt.Printf("\nSelect instance (1-%d): ", len(bridges))

	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		return nil, fmt.Errorf("invalid input")
	}

	if choice < 1 || choice > len(bridges) {
		return nil, fmt.Errorf("invalid choice")
	}

	return &bridges[choice-1], nil
}

// VSCodeInstance represents a running VSCode process (minimal version)
type VSCodeInstance struct {
	PID           int32
	Name          string
	WorkspacePath string
}

// detectParentVSCode tries to detect if we're running inside a VSCode terminal
func detectParentVSCode() (*VSCodeInstance, error) {
	// Get parent process ID
	ppid := int32(os.Getppid())

	// Walk up the process tree (max 10 levels)
	currentPID := ppid
	for i := 0; i < 10; i++ {
		proc, err := process.NewProcess(currentPID)
		if err != nil {
			return nil, fmt.Errorf("failed to get process %d: %w", currentPID, err)
		}

		name, err := proc.Name()
		if err != nil {
			return nil, fmt.Errorf("failed to get process name: %w", err)
		}

		// Check if this is a VSCode process
		if isVSCodeProcess(name) {
			cmdline, _ := proc.Cmdline()
			return &VSCodeInstance{
				PID:           currentPID,
				Name:          name,
				WorkspacePath: extractWorkspacePath(cmdline),
			}, nil
		}

		// Get parent process
		parent, err := proc.Parent()
		if err != nil || parent == nil {
			break
		}
		currentPID = parent.Pid
	}

	return nil, fmt.Errorf("VSCode parent process not found")
}

// isVSCodeProcess checks if a process name matches VSCode
func isVSCodeProcess(name string) bool {
	lowerName := strings.ToLower(name)
	return strings.Contains(lowerName, "code") ||
		strings.Contains(lowerName, "code-insiders") ||
		strings.Contains(lowerName, "electron") // VSCode uses Electron
}

// extractWorkspacePath tries to extract workspace path from command line
func extractWorkspacePath(cmdline string) string {
	// Look for --folder-uri flag
	parts := strings.Split(cmdline, " ")
	for i, part := range parts {
		if part == "--folder-uri" && i+1 < len(parts) {
			uri := parts[i+1]
			// Remove file:// prefix if present
			return strings.TrimPrefix(uri, "file://")
		}
		// Also check for direct path arguments
		if strings.HasPrefix(part, "/") || strings.HasPrefix(part, "~/") {
			// This might be a workspace path
			if info, err := os.Stat(part); err == nil && info.IsDir() {
				return part
			}
		}
	}
	return ""
}

// getBridgeDirectory returns the platform-specific bridge directory
func getBridgeDirectory() string {
	var tmpDir string

	switch runtime.GOOS {
	case "windows":
		tmpDir = os.Getenv("TEMP")
		if tmpDir == "" {
			tmpDir = os.Getenv("TMP")
		}
		if tmpDir == "" {
			tmpDir = "C:\\Windows\\Temp"
		}
	default:
		tmpDir = "/tmp"
		if envTmp := os.Getenv("TMPDIR"); envTmp != "" {
			tmpDir = envTmp
		}
	}

	return filepath.Join(tmpDir, "vstr-bridge")
}

// validateDirectoryPermissions checks directory has secure permissions
func validateDirectoryPermissions(dirPath string) bool {
	if runtime.GOOS == "windows" {
		// Limited validation on Windows
		return true
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}

	mode := info.Mode().Perm()
	// Check that only owner has permissions (max 0700)
	return mode&0o077 == 0
}

// validateBridgeFile validates a single bridge file for security compliance
func validateBridgeFile(filePath string) (*BridgeInfo, error) {
	// 1. Validate file permissions
	if !security.ValidateFilePermissions(filePath) {
		return nil, fmt.Errorf("insecure file permissions")
	}

	// 2. Read and parse content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var bridgeInfo BridgeInfo
	if err := json.Unmarshal(data, &bridgeInfo); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// 3. Validate structure and required fields
	if err := validateBridgeStructure(&bridgeInfo); err != nil {
		return nil, err
	}

	return &bridgeInfo, nil
}

// validateBridgeStructure validates bridge information structure and values
func validateBridgeStructure(info *BridgeInfo) error {
	if info.Port <= 0 || info.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", info.Port)
	}

	if info.PID <= 0 {
		return fmt.Errorf("invalid PID: %d", info.PID)
	}

	if len(info.AuthToken) < security.MinTokenLength {
		return fmt.Errorf("invalid auth token length")
	}

	if !info.Secure {
		return fmt.Errorf("bridge is not in secure mode")
	}

	return nil
}
