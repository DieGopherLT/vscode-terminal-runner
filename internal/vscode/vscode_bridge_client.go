// internal/vscode/vscode_bridge_client.go
package vscode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
)

// BridgeClient handles communication with the VSCode extension bridge
type BridgeClient struct {
	baseURL string
	client  *http.Client
}

// NewBridgeClient creates a new bridge client for the given port
func NewBridgeClient(port int) *BridgeClient {
	return &BridgeClient{
		baseURL: fmt.Sprintf("http://localhost:%d", port),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Ping checks if the bridge is alive and responding
func (bc *BridgeClient) Ping() error {
	resp, err := bc.client.Get(bc.baseURL + "/ping")
	if err != nil {
		return fmt.Errorf("bridge not responding: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bridge returned status %d", resp.StatusCode)
	}

	return nil
}

// ExecuteTask sends a single task to be executed in VSCode
func (bc *BridgeClient) ExecuteTask(task models.Task) error {
	payload := taskToPayload(task)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	resp, err := bc.client.Post(
		bc.baseURL+"/task",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to send task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("bridge error: %s", errResp.Error)
	}

	return nil
}

// ExecuteWorkspace sends a workspace configuration to be executed
func (bc *BridgeClient) ExecuteWorkspace(workspace models.Workspace) error {
	payload := map[string]interface{}{
		"name":  workspace.Name,
		"tasks": tasksToPayload(workspace.Tasks),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace: %w", err)
	}

	resp, err := bc.client.Post(
		bc.baseURL+"/workspace",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to send workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("bridge error: %s", errResp.Error)
	}

	// Parse results
	var result struct {
		Success bool `json:"success"`
		Results []struct {
			Task    string `json:"task"`
			Success bool   `json:"success"`
			Error   string `json:"error,omitempty"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for failed tasks
	var failures []string
	for _, r := range result.Results {
		if !r.Success {
			failures = append(failures, fmt.Sprintf("%s: %s", r.Task, r.Error))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("some tasks failed: %v", failures)
	}

	return nil
}

// taskToPayload converts a Task model to the bridge API format
func taskToPayload(task models.Task) map[string]interface{} {
	return map[string]interface{}{
		"name":      task.Name,
		"path":      task.Path,
		"cmds":      task.Cmds,
		"icon":      task.Icon,
		"iconColor": task.IconColor,
	}
}

// tasksToPayload converts multiple tasks to payload format
func tasksToPayload(tasks []models.Task) []map[string]interface{} {
	payloads := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		payloads[i] = taskToPayload(task)
	}
	return payloads
}
