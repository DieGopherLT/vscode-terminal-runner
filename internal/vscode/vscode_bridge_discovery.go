package vscode

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// DiscoverBridge finds the correct bridge instance for the current VSCode
func DiscoverBridge() (*BridgeInfo, error) {
	// 1. First check environment variable (if running from VSCode terminal)
	if port := os.Getenv("VSTR"); port != "" {
		info, err := validateBridge(port)
		if err == nil {
			return info, nil
		}
		styles.PrintWarning("Environment variable VSTR found but bridge not responding")
	}

	// 2. Try to detect parent VSCode process
	if instance, err := detectParentVSCode(); err == nil {
		if bridge, err := findBridgeByWorkspace(instance.WorkspacePath); err == nil {
			return bridge, nil
		}
	}

	// 3. List all available bridges
	bridges, err := ListAvailableBridges()
	if err != nil {
		return nil, err
	}

	if len(bridges) == 0 {
		return nil, fmt.Errorf("no VSCode bridge instances found")
	}

	// 4. If only one bridge, use it
	if len(bridges) == 1 {
		return &bridges[0], nil
	}

	// 5. Multiple bridges - let user select
	return selectBridge(bridges)
}

// DiscoverSecureBridge finds the correct secure bridge instance for the current VSCode
func DiscoverSecureBridge() (*BridgeInfo, error) {
	authManager := security.NewAuthManager()
	bridgeDir := getBridgeDirectory()
	
	if _, err := os.Stat(bridgeDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("bridge directory not found: %s", bridgeDir)
	}
	
	// Verify directory permissions
	if !validateDirectoryPermissions(bridgeDir) {
		return nil, fmt.Errorf("bridge directory has insecure permissions")
	}
	
	files, err := os.ReadDir(bridgeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read bridge directory: %w", err)
	}
	
	var validBridges []*BridgeInfo
	
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "bridge-") || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		filePath := filepath.Join(bridgeDir, file.Name())
		
		// Validate each bridge file
		bridgeInfo, err := validateSecureBridgeFile(authManager, filePath)
		if err != nil {
			// Log but continue searching other files
			styles.PrintError(fmt.Sprintf("Skipping invalid bridge file %s: %v", file.Name(), err))
			continue
		}
		
		validBridges = append(validBridges, bridgeInfo)
	}
	
	if len(validBridges) == 0 {
		return nil, fmt.Errorf("no valid secure bridge found")
	}
	
	// Return the most recent bridge
	latest := validBridges[0]
	for _, bridge := range validBridges[1:] {
		if bridge.InstanceID > latest.InstanceID {
			latest = bridge
		}
	}
	
	return latest, nil
}

// ListAvailableBridges scans for active bridge instances
func ListAvailableBridges() ([]BridgeInfo, error) {
	tmpDir := filepath.Join(os.TempDir(), "vstr-bridge")

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BridgeInfo{}, nil
		}
		return nil, err
	}

	var bridges []BridgeInfo

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(tmpDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var info BridgeInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}

		if IsBridgeOperative(info.Port) {
			bridges = append(bridges, info)
		} else {
			os.Remove(path)
		}
	}

	return bridges, nil
}

// findBridgeByWorkspace finds a bridge matching the given workspace path
func findBridgeByWorkspace(path string) (*BridgeInfo, error) {
	bridges, err := ListAvailableBridges()
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

// IsBridgeOperative checks if a bridge server is responding
func IsBridgeOperative(port int) bool {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// validateBridge validates a bridge on the given port
func validateBridge(portStr string) (*BridgeInfo, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bridge not responding on port %d", port)
	}

	// Try to get more info from the bridge
	var pingResponse struct {
		Status    string `json:"status"`
		Workspace string `json:"workspace"`
		Port      int    `json:"port"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pingResponse); err == nil {
		return &BridgeInfo{
			Port:          port,
			WorkspaceName: pingResponse.Workspace,
		}, nil
	}

	return &BridgeInfo{Port: port}, nil
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

// validateSecureBridgeFile validates a single bridge file for security compliance
func validateSecureBridgeFile(authManager *security.AuthManager, filePath string) (*BridgeInfo, error) {
	// 1. Validate file permissions
	if !authManager.ValidateFilePermissions(filePath) {
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
	
	if len(info.AuthToken) < 32 {
		return fmt.Errorf("invalid auth token length")
	}
	
	if !info.Secure {
		return fmt.Errorf("bridge is not in secure mode")
	}
	
	return nil
}