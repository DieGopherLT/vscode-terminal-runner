// internal/client/secure_client.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/security"
	"github.com/samber/lo"
)

// SecureClient handles secure communication with VSCode bridge
type SecureClient struct {
	httpClient  *http.Client
	authManager *security.AuthManager
	baseURL     string
}

// NewSecureClient creates a new secure client for bridge communication
func NewSecureClient(port int) *SecureClient {
	return &SecureClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives:   true,
				MaxIdleConns:        1,
				IdleConnTimeout:     30 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		authManager: security.NewAuthManager(),
		baseURL:     fmt.Sprintf("http://localhost:%d", port),
	}
}

// LoadAuth loads authentication credentials from bridge file
func (c *SecureClient) LoadAuth(bridgeFilePath string) error {
	return c.authManager.LoadTokenFromBridge(bridgeFilePath)
}

// TestConnection verifies connectivity and authentication with bridge
func (c *SecureClient) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/ping", nil)
	if err != nil {
		return err
	}
	
	// Add authentication headers
	for key, value := range c.authManager.GetAuthHeaders() {
		req.Header.Set(key, value)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 401 {
		return fmt.Errorf("authentication failed - invalid token")
	}
	
	if resp.StatusCode == 429 {
		return fmt.Errorf("rate limit exceeded - too many requests")
	}
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	// Verify bridge responds as secure
	var pingResp struct {
		Status   string   `json:"status"`
		Secure   bool     `json:"secure"`
		Features []string `json:"security_features"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&pingResp); err != nil {
		return fmt.Errorf("invalid ping response: %w", err)
	}
	
	if !pingResp.Secure {
		return fmt.Errorf("bridge is not running in secure mode")
	}
	
	return nil
}

// ExecuteTask sends a task for secure execution
func (c *SecureClient) ExecuteTask(ctx context.Context, task models.Task) error {
	payload := c.taskToPayload(task)
	
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(
		ctx, 
		"POST", 
		c.baseURL+"/task", 
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Add authentication headers
	for key, value := range c.authManager.GetAuthHeaders() {
		req.Header.Set(key, value)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	return c.handleResponse(resp)
}

// ExecuteWorkspace sends a workspace for secure execution
func (c *SecureClient) ExecuteWorkspace(ctx context.Context, workspace models.Workspace) error {
	payload := map[string]interface{}{
		"name":  workspace.Name,
		"tasks": c.tasksToPayload(workspace.Tasks),
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(
		ctx, 
		"POST", 
		c.baseURL+"/workspace", 
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Add authentication headers
	for key, value := range c.authManager.GetAuthHeaders() {
		req.Header.Set(key, value)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	return c.handleResponse(resp)
}

// handleResponse processes HTTP response and handles security-specific errors
func (c *SecureClient) handleResponse(resp *http.Response) error {
	if isSuccessResponse(resp.StatusCode) {
		return nil
	}
	
	apiResp, err := c.parseAPIResponse(resp.Body)
	if err != nil {
		return fmt.Errorf("invalid response format: %w", err)
	}
	
	return c.createErrorFromStatusCode(resp.StatusCode, apiResp)
}

// isSuccessResponse checks if the status code indicates success
func isSuccessResponse(statusCode int) bool {
	return statusCode == 200
}

// parseAPIResponse parses the API response body into a structured format
func (c *SecureClient) parseAPIResponse(body io.ReadCloser) (*apiResponse, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	
	var apiResp apiResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, err
	}
	
	return &apiResp, nil
}

// apiResponse represents the structure of API responses
type apiResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// createErrorFromStatusCode creates appropriate error messages based on status codes
func (c *SecureClient) createErrorFromStatusCode(statusCode int, apiResp *apiResponse) error {
	switch statusCode {
	case 401:
		return fmt.Errorf("authentication failed: %s", apiResp.Error)
	case 403:
		return fmt.Errorf("command blocked by security policy: %s", apiResp.Error)
	case 429:
		return fmt.Errorf("rate limit exceeded: %s", apiResp.Error)
	default:
		return fmt.Errorf("request failed (%d): %s", statusCode, apiResp.Error)
	}
}

// taskToPayload converts a Task model to bridge API format
func (c *SecureClient) taskToPayload(task models.Task) map[string]interface{} {
	return map[string]interface{}{
		"name":      task.Name,
		"path":      task.Path,
		"cmds":      task.Cmds,
		"icon":      task.Icon,
		"iconColor": task.IconColor,
	}
}

// tasksToPayload converts multiple tasks to payload format using functional approach
func (c *SecureClient) tasksToPayload(tasks []models.Task) []map[string]interface{} {
	return lo.Map(tasks, func(task models.Task, _ int) map[string]interface{} {
		return c.taskToPayload(task)
	})
}