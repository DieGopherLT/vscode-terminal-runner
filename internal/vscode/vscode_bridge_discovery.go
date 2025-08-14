package vscode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/samber/lo"
)

type BridgeInfo struct {
	Port          int       `json:"port"`
	PID           int       `json:"pid"`
	InstanceID    int       `json:"instance_id"`
	WorkspacePath string    `json:"workspace_path"`
	WorkspaceName string    `json:"workspace_name"`
	Timestamp     time.Time `json:"timestamp"`
}

// DiscoverBridge finds the correct bridge instance for the current VSCode
func DiscoverBridge() (*BridgeInfo, error) {
	if port := os.Getenv("VSCT"); port != "" {
		info, err := validateBridge(port)
		if err == nil {
			return info, nil
		}
	}

	instance, err := DetectParentVSCode()
	if err == nil {
		return findBridgeByWorkspace(instance.GetWorkspacePath())
	}

	bridges, err := ListAvailableBridges()
	if err != nil {
		return nil, err
	}

	if len(bridges) == 0 {
		return nil, fmt.Errorf("no VSCode bridge instances found")
	}

	if len(bridges) == 1 {
		return &bridges[0], nil
	}

	return selectBridge(bridges)
}

func ListAvailableBridges() ([]BridgeInfo, error) {
	tmpDir := filepath.Join(os.TempDir(), "vscr-bridge")

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

		if isBridgeAlive(info.Port) {
			bridges = append(bridges, info)
		} else {
			os.Remove(path)
		}
	}

	return bridges, nil
}

func findBridgeByWorkspace(path string) (*BridgeInfo, error) {

	bridges, err := ListAvailableBridges()
	if err != nil {
		return nil, err
	}

	bridge, found := lo.Find(bridges, func(b BridgeInfo) bool {
		return b.WorkspacePath == path
	})

	if !found {
		return nil, fmt.Errorf("No bridge found for workspace %s", path)
	}

	return &bridge, nil
}

func selectBridge(bridges []BridgeInfo) (*BridgeInfo, error) {

	styles.PrintInfo("\nMultiple VSCode instances detected")
	fmt.Println()

	for i, bridge := range bridges {
		fmt.Printf("%d, %s (PID %d)\n", i+1, styles.RenderInfoMessage(bridge.WorkspaceName), bridge.PID)
		fmt.Printf("\tPath: %s\n", styles.RenderInfoMessage(bridge.WorkspacePath))
	}

	fmt.Printf("\nSelect instance (1-%d):  ", len(bridges))
	var choice int

	fmt.Scanln(&choice)
	if choice < 1 || choice > len(bridges) {
		return nil, fmt.Errorf("invalid choice")
	}

	return &bridges[choice-1], nil
}

func isBridgeAlive(port int) bool {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func validateBridge(port string) (*BridgeInfo, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/ping", port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bridge not responding on port %s", port)
	}

	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	return &BridgeInfo{Port: parsedPort}, nil
}
