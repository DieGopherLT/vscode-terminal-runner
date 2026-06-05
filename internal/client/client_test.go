package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/testutils"
)

// ---- helpers ---------------------------------------------------------------

// deadPortClient creates a Client pointed at a port that is guaranteed to
// refuse connections. It binds a listener, reads its port, immediately closes
// the listener, and returns a Client for that port. This avoids double-closing
// the httptest server (which t.Cleanup already handles) and is safe for
// parallel tests.
func deadPortClient(t *testing.T) *Client {
	t.Helper()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("deadPortClient: could not bind listener: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return NewClient(port)
}

// captureAuthHeader returns a handler that records the Authorization header
// from each request into the provided pointer and then delegates to next.
// This validates that the client actually transmits the auth token.
func captureAuthHeader(captured *string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*captured = r.Header.Get("Authorization")
		next(w, r)
	}
}

// ---- LoadAuthFromToken tests -----------------------------------------------

func TestLoadAuthFromToken(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{
			name:      "empty token is rejected",
			token:     "",
			wantError: true,
		},
		{
			name:      "token shorter than MinTokenLength is rejected",
			token:     "tooshort",
			wantError: true,
		},
		{
			name:      "token of 31 bytes is rejected (boundary off-by-one)",
			token:     strings.Repeat("a", 31),
			wantError: true,
		},
		{
			name:      "token exactly at MinTokenLength is accepted",
			token:     testutils.ValidTestToken, // 32 bytes
			wantError: false,
		},
		{
			name:      "token longer than MinTokenLength is accepted",
			token:     strings.Repeat("x", 64),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(9999)
			err := c.LoadAuthFromToken(tt.token)
			if tt.wantError && err == nil {
				t.Errorf("expected error for token %q but got nil", tt.token)
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error for token %q but got: %v", tt.token, err)
			}
		})
	}
}

// ---- TestConnection tests --------------------------------------------------

func TestClient_TestConnection(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path returns nil", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{})
		c := clientForServer(t, server)

		if err := c.TestConnection(ctx); err != nil {
			t.Errorf("expected nil error but got: %v", err)
		}
	})

	t.Run("auth header is transmitted when token is loaded", func(t *testing.T) {
		var capturedAuth string
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			PingHandler: captureAuthHeader(&capturedAuth, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"status": "ok",
					"secure": true,
				})
			}),
		})
		c := clientForServer(t, server)
		if err := c.LoadAuthFromToken(testutils.ValidTestToken); err != nil {
			t.Fatalf("LoadAuthFromToken: %v", err)
		}

		if err := c.TestConnection(ctx); err != nil {
			t.Fatalf("TestConnection: %v", err)
		}

		want := "Bearer " + testutils.ValidTestToken
		if capturedAuth != want {
			t.Errorf("Authorization header = %q, want %q", capturedAuth, want)
		}
	})

	t.Run("no auth header when token is not loaded", func(t *testing.T) {
		var capturedAuth string
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			PingHandler: captureAuthHeader(&capturedAuth, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"status": "ok",
					"secure": true,
				})
			}),
		})
		c := clientForServer(t, server)

		if err := c.TestConnection(ctx); err != nil {
			t.Fatalf("TestConnection: %v", err)
		}

		if capturedAuth != "" {
			t.Errorf("expected no Authorization header but got %q", capturedAuth)
		}
	})

	t.Run("401 returns authentication error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			PingHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
		})
		c := clientForServer(t, server)

		err := c.TestConnection(ctx)
		if err == nil {
			t.Fatal("expected error for 401 but got nil")
		}
		if !strings.Contains(err.Error(), "authentication failed") {
			t.Errorf("error %q does not contain %q", err.Error(), "authentication failed")
		}
	})

	t.Run("429 returns rate limit error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			PingHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
			},
		})
		c := clientForServer(t, server)

		err := c.TestConnection(ctx)
		if err == nil {
			t.Fatal("expected error for 429 but got nil")
		}
		if !strings.Contains(err.Error(), "rate limit exceeded") {
			t.Errorf("error %q does not contain %q", err.Error(), "rate limit exceeded")
		}
	})

	t.Run("unexpected non-200 returns status code error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			PingHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
		})
		c := clientForServer(t, server)

		err := c.TestConnection(ctx)
		if err == nil {
			t.Fatal("expected error for 500 but got nil")
		}
		if !strings.Contains(err.Error(), "unexpected status code: 500") {
			t.Errorf("error %q does not contain %q", err.Error(), "unexpected status code: 500")
		}
	})

	t.Run("200 with malformed JSON returns invalid ping response error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			PingHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not-json")) //nolint:errcheck
			},
		})
		c := clientForServer(t, server)

		err := c.TestConnection(ctx)
		if err == nil {
			t.Fatal("expected error for malformed JSON but got nil")
		}
		if !strings.Contains(err.Error(), "invalid ping response") {
			t.Errorf("error %q does not contain %q", err.Error(), "invalid ping response")
		}
	})

	t.Run("200 with secure false returns bridge not secure error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			PingHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"status": "ok",
					"secure": false,
				})
			},
		})
		c := clientForServer(t, server)

		err := c.TestConnection(ctx)
		if err == nil {
			t.Fatal("expected error for insecure bridge but got nil")
		}
		if !strings.Contains(err.Error(), "bridge is not running in secure mode") {
			t.Errorf("error %q does not contain %q", err.Error(), "bridge is not running in secure mode")
		}
	})

	t.Run("server down returns connection failed error", func(t *testing.T) {
		c := deadPortClient(t)

		err := c.TestConnection(ctx)
		if err == nil {
			t.Fatal("expected error when server is down but got nil")
		}
		if !strings.Contains(err.Error(), "connection failed") {
			t.Errorf("error %q does not contain %q", err.Error(), "connection failed")
		}
	})
}

// ---- ExecuteTask tests -----------------------------------------------------

func TestClient_ExecuteTask(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path returns nil", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{})
		c := clientForServer(t, server)
		if err := c.LoadAuthFromToken(testutils.ValidTestToken); err != nil {
			t.Fatalf("LoadAuthFromToken: %v", err)
		}
		task := testutils.NewTask().WithName("build").WithCmds("make").Build()

		if err := c.ExecuteTask(ctx, task); err != nil {
			t.Errorf("expected nil error but got: %v", err)
		}
	})

	t.Run("auth header is transmitted when token is loaded", func(t *testing.T) {
		var capturedAuth string
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			TaskHandler: captureAuthHeader(&capturedAuth, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": true,
					"message": "task executed",
				})
			}),
		})
		c := clientForServer(t, server)
		if err := c.LoadAuthFromToken(testutils.ValidTestToken); err != nil {
			t.Fatalf("LoadAuthFromToken: %v", err)
		}
		task := testutils.NewTask().Build()

		if err := c.ExecuteTask(ctx, task); err != nil {
			t.Fatalf("ExecuteTask: %v", err)
		}

		want := "Bearer " + testutils.ValidTestToken
		if capturedAuth != want {
			t.Errorf("Authorization header = %q, want %q", capturedAuth, want)
		}
	})

	// Characterization test: ExecuteTask ignores the body's success field on 200.
	// handleResponse only checks the HTTP status code; the JSON body is not read.
	// This is asymmetric with ExecuteWorkspace which does inspect per-task success.
	// Pinned as current behavior. The bridge contract is silent on whether
	// body-success is authoritative for /task, so this is an observation, not a bug.
	t.Run("200 with success false in body returns nil (characterization)", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			TaskHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": false,
					"error":   "task failed internally",
				})
			},
		})
		c := clientForServer(t, server)
		task := testutils.NewTask().Build()

		// Current behavior: handleResponse only checks status code on 200 — body is ignored.
		err := c.ExecuteTask(ctx, task)
		if err != nil {
			t.Errorf("characterization: expected nil (body ignored on 200) but got: %v", err)
		}
	})

	t.Run("401 returns authentication error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			TaskHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": false,
					"error":   "invalid token",
				})
			},
		})
		c := clientForServer(t, server)
		task := testutils.NewTask().Build()

		err := c.ExecuteTask(ctx, task)
		if err == nil {
			t.Fatal("expected error for 401 but got nil")
		}
		if !strings.Contains(err.Error(), "authentication failed") {
			t.Errorf("error %q does not contain %q", err.Error(), "authentication failed")
		}
		if !strings.Contains(err.Error(), "invalid token") {
			t.Errorf("error %q does not contain the api error %q", err.Error(), "invalid token")
		}
	})

	t.Run("403 returns security policy error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			TaskHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": false,
					"error":   "blocked command",
				})
			},
		})
		c := clientForServer(t, server)
		task := testutils.NewTask().Build()

		err := c.ExecuteTask(ctx, task)
		if err == nil {
			t.Fatal("expected error for 403 but got nil")
		}
		if !strings.Contains(err.Error(), "command blocked by security policy") {
			t.Errorf("error %q does not contain %q", err.Error(), "command blocked by security policy")
		}
	})

	t.Run("429 returns rate limit error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			TaskHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": false,
					"error":   "slow down",
				})
			},
		})
		c := clientForServer(t, server)
		task := testutils.NewTask().Build()

		err := c.ExecuteTask(ctx, task)
		if err == nil {
			t.Fatal("expected error for 429 but got nil")
		}
		if !strings.Contains(err.Error(), "rate limit exceeded") {
			t.Errorf("error %q does not contain %q", err.Error(), "rate limit exceeded")
		}
	})

	t.Run("unexpected non-200 returns request failed with status code", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			TaskHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": false,
					"error":   "internal error",
				})
			},
		})
		c := clientForServer(t, server)
		task := testutils.NewTask().Build()

		err := c.ExecuteTask(ctx, task)
		if err == nil {
			t.Fatal("expected error for 500 but got nil")
		}
		if !strings.Contains(err.Error(), "request failed (500)") {
			t.Errorf("error %q does not contain %q", err.Error(), "request failed (500)")
		}
	})

	t.Run("non-200 with malformed body returns invalid response format error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			TaskHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("not-json")) //nolint:errcheck
			},
		})
		c := clientForServer(t, server)
		task := testutils.NewTask().Build()

		err := c.ExecuteTask(ctx, task)
		if err == nil {
			t.Fatal("expected error for malformed body but got nil")
		}
		if !strings.Contains(err.Error(), "invalid response format") {
			t.Errorf("error %q does not contain %q", err.Error(), "invalid response format")
		}
	})

	t.Run("server down returns request failed error", func(t *testing.T) {
		c := deadPortClient(t)
		task := testutils.NewTask().Build()

		err := c.ExecuteTask(ctx, task)
		if err == nil {
			t.Fatal("expected error when server is down but got nil")
		}
		if !strings.Contains(err.Error(), "request failed") {
			t.Errorf("error %q does not contain %q", err.Error(), "request failed")
		}
	})

	t.Run("task with all fields transmits complete payload", func(t *testing.T) {
		var receivedBody map[string]interface{}
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			TaskHandler: func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": true,
					"message": "task executed",
				})
			},
		})
		c := clientForServer(t, server)
		task := testutils.NewTask().
			WithName("deploy").
			WithPath("/repo").
			WithCmds("make deploy").
			WithIcon("rocket").
			WithIconColor("blue").
			Build()

		if err := c.ExecuteTask(ctx, task); err != nil {
			t.Fatalf("ExecuteTask: %v", err)
		}

		checkStringField := func(key, want string) {
			t.Helper()
			got, ok := receivedBody[key].(string)
			if !ok || got != want {
				t.Errorf("payload[%q] = %v, want %q", key, receivedBody[key], want)
			}
		}
		checkStringField("name", "deploy")
		checkStringField("path", "/repo")
		checkStringField("icon", "rocket")
		checkStringField("iconColor", "blue")
	})
}

// ---- ExecuteWorkspace tests ------------------------------------------------

func TestClient_ExecuteWorkspace(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path with empty results returns nil", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{})
		c := clientForServer(t, server)
		if err := c.LoadAuthFromToken(testutils.ValidTestToken); err != nil {
			t.Fatalf("LoadAuthFromToken: %v", err)
		}
		ws := testutils.NewWorkspace().Build()

		if err := c.ExecuteWorkspace(ctx, ws); err != nil {
			t.Errorf("expected nil error but got: %v", err)
		}
	})

	t.Run("auth header is transmitted when token is loaded", func(t *testing.T) {
		var capturedAuth string
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			WorkspaceHandler: captureAuthHeader(&capturedAuth, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": true,
					"results": []interface{}{},
				})
			}),
		})
		c := clientForServer(t, server)
		if err := c.LoadAuthFromToken(testutils.ValidTestToken); err != nil {
			t.Fatalf("LoadAuthFromToken: %v", err)
		}
		ws := testutils.NewWorkspace().Build()

		if err := c.ExecuteWorkspace(ctx, ws); err != nil {
			t.Fatalf("ExecuteWorkspace: %v", err)
		}

		want := "Bearer " + testutils.ValidTestToken
		if capturedAuth != want {
			t.Errorf("Authorization header = %q, want %q", capturedAuth, want)
		}
	})

	t.Run("all tasks succeed returns nil", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			WorkspaceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": true,
					"results": []map[string]interface{}{
						{"task": "build", "success": true},
						{"task": "test", "success": true},
					},
				})
			},
		})
		c := clientForServer(t, server)
		task1 := testutils.NewTask().WithName("build").Build()
		task2 := testutils.NewTask().WithName("test").Build()
		ws := testutils.NewWorkspace().WithTasks(task1, task2).Build()

		if err := c.ExecuteWorkspace(ctx, ws); err != nil {
			t.Errorf("expected nil error but got: %v", err)
		}
	})

	t.Run("one task failure returns error listing the failed task", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			WorkspaceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": false,
					"results": []map[string]interface{}{
						{"task": "build", "success": false, "error": "exit status 1"},
					},
				})
			},
		})
		c := clientForServer(t, server)
		ws := testutils.NewWorkspace().
			WithTasks(testutils.NewTask().WithName("build").Build()).
			Build()

		err := c.ExecuteWorkspace(ctx, ws)
		if err == nil {
			t.Fatal("expected error for failed task but got nil")
		}
		if !strings.Contains(err.Error(), "some tasks failed") {
			t.Errorf("error %q does not contain %q", err.Error(), "some tasks failed")
		}
		if !strings.Contains(err.Error(), "build") {
			t.Errorf("error %q does not contain the task name %q", err.Error(), "build")
		}
		if !strings.Contains(err.Error(), "exit status 1") {
			t.Errorf("error %q does not contain the error detail %q", err.Error(), "exit status 1")
		}
	})

	t.Run("multiple tasks with mixed results reports only failures", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			WorkspaceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": false,
					"results": []map[string]interface{}{
						{"task": "lint", "success": true},
						{"task": "build", "success": false, "error": "compile error"},
						{"task": "test-run", "success": false, "error": "timeout"},
					},
				})
			},
		})
		c := clientForServer(t, server)
		ws := testutils.NewWorkspace().Build()

		err := c.ExecuteWorkspace(ctx, ws)
		if err == nil {
			t.Fatal("expected error for mixed results but got nil")
		}
		if strings.Contains(err.Error(), "lint") {
			t.Errorf("error %q should not mention the successful task 'lint'", err.Error())
		}
		if !strings.Contains(err.Error(), "build") {
			t.Errorf("error %q should mention failed task 'build'", err.Error())
		}
		if !strings.Contains(err.Error(), "test-run") {
			t.Errorf("error %q should mention failed task 'test-run'", err.Error())
		}
	})

	t.Run("200 with malformed body returns parse error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			WorkspaceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not-json")) //nolint:errcheck
			},
		})
		c := clientForServer(t, server)
		ws := testutils.NewWorkspace().Build()

		err := c.ExecuteWorkspace(ctx, ws)
		if err == nil {
			t.Fatal("expected error for malformed body but got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse response") {
			t.Errorf("error %q does not contain %q", err.Error(), "failed to parse response")
		}
	})

	t.Run("401 returns authentication error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			WorkspaceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"success": false,
					"error":   "token expired",
				})
			},
		})
		c := clientForServer(t, server)
		ws := testutils.NewWorkspace().Build()

		err := c.ExecuteWorkspace(ctx, ws)
		if err == nil {
			t.Fatal("expected error for 401 but got nil")
		}
		if !strings.Contains(err.Error(), "authentication failed") {
			t.Errorf("error %q does not contain %q", err.Error(), "authentication failed")
		}
	})

	t.Run("non-200 with malformed body returns invalid response format error", func(t *testing.T) {
		server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
			WorkspaceHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("not-json")) //nolint:errcheck
			},
		})
		c := clientForServer(t, server)
		ws := testutils.NewWorkspace().Build()

		err := c.ExecuteWorkspace(ctx, ws)
		if err == nil {
			t.Fatal("expected error for malformed body but got nil")
		}
		if !strings.Contains(err.Error(), "invalid response format") {
			t.Errorf("error %q does not contain %q", err.Error(), "invalid response format")
		}
	})

	t.Run("server down returns request failed error", func(t *testing.T) {
		c := deadPortClient(t)
		ws := testutils.NewWorkspace().Build()

		err := c.ExecuteWorkspace(ctx, ws)
		if err == nil {
			t.Fatal("expected error when server is down but got nil")
		}
		if !strings.Contains(err.Error(), "request failed") {
			t.Errorf("error %q does not contain %q", err.Error(), "request failed")
		}
	})
}

// ---- taskToPayload / tasksToPayload white-box tests ------------------------

func TestClient_taskToPayload(t *testing.T) {
	c := NewClient(9999)

	t.Run("task with empty cmds produces payload with empty cmds slice", func(t *testing.T) {
		task := testutils.NewTask().WithCmds().Build()
		payload := c.taskToPayload(task)

		if payload["name"] != task.Name {
			t.Errorf("name = %v, want %v", payload["name"], task.Name)
		}
		cmds, ok := payload["cmds"].([]string)
		if !ok {
			t.Errorf("cmds field is not []string: %T", payload["cmds"])
		}
		if len(cmds) != 0 {
			t.Errorf("expected empty cmds slice, got %v", cmds)
		}
	})

	t.Run("fully populated task maps all five fields correctly", func(t *testing.T) {
		task := testutils.NewTask().
			WithName("deploy").
			WithPath("/srv/app").
			WithCmds("make deploy", "echo done").
			WithIcon("rocket").
			WithIconColor("blue").
			Build()

		payload := c.taskToPayload(task)

		checks := []struct {
			key  string
			want string
		}{
			{"name", "deploy"},
			{"path", "/srv/app"},
			{"icon", "rocket"},
			{"iconColor", "blue"},
		}
		for _, ch := range checks {
			got, ok := payload[ch.key].(string)
			if !ok || got != ch.want {
				t.Errorf("payload[%q] = %v, want %q", ch.key, payload[ch.key], ch.want)
			}
		}

		cmds, ok := payload["cmds"].([]string)
		if !ok {
			t.Fatalf("cmds is not []string: %T", payload["cmds"])
		}
		if len(cmds) != 2 || cmds[0] != "make deploy" || cmds[1] != "echo done" {
			t.Errorf("cmds = %v, want [make deploy echo done]", cmds)
		}
	})

	t.Run("icon and iconColor default to empty string for zero task", func(t *testing.T) {
		task := models.Task{Name: "minimal"}
		payload := c.taskToPayload(task)

		if payload["icon"] != "" {
			t.Errorf("icon = %v, want empty string", payload["icon"])
		}
		if payload["iconColor"] != "" {
			t.Errorf("iconColor = %v, want empty string", payload["iconColor"])
		}
	})
}

func TestClient_tasksToPayload(t *testing.T) {
	c := NewClient(9999)

	t.Run("nil tasks slice returns empty slice", func(t *testing.T) {
		result := c.tasksToPayload(nil)
		if len(result) != 0 {
			t.Errorf("expected empty slice, got length %d", len(result))
		}
	})

	t.Run("empty tasks slice returns empty slice", func(t *testing.T) {
		result := c.tasksToPayload([]models.Task{})
		if len(result) != 0 {
			t.Errorf("expected empty slice, got length %d", len(result))
		}
	})

	t.Run("single task is converted correctly", func(t *testing.T) {
		task := testutils.NewTask().WithName("single").Build()
		result := c.tasksToPayload([]models.Task{task})

		if len(result) != 1 {
			t.Fatalf("expected 1 result, got %d", len(result))
		}
		if result[0]["name"] != "single" {
			t.Errorf("result[0][name] = %v, want single", result[0]["name"])
		}
	})

	t.Run("multiple tasks are all converted preserving order", func(t *testing.T) {
		tasks := []models.Task{
			testutils.NewTask().WithName("first").Build(),
			testutils.NewTask().WithName("second").Build(),
			testutils.NewTask().WithName("third").Build(),
		}

		result := c.tasksToPayload(tasks)

		if len(result) != 3 {
			t.Fatalf("expected 3 results, got %d", len(result))
		}
		names := []string{"first", "second", "third"}
		for i, want := range names {
			got, ok := result[i]["name"].(string)
			if !ok || got != want {
				t.Errorf("result[%d][name] = %v, want %q", i, result[i]["name"], want)
			}
		}
	})
}
