// pkg/testutils/bridge_server.go
package testutils

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ValidTestToken is a 32-byte token that satisfies security.MinTokenLength.
// Use it in tests that call c.LoadAuthFromToken to avoid each test fabricating
// its own minimum-length string.
const ValidTestToken = "00000000000000000000000000000000"

// BridgeHandlerConfig holds per-test handler overrides for BridgeTestServer.
// Every field is optional; a zero value keeps the default happy-path handler.
type BridgeHandlerConfig struct {
	// PingHandler replaces the default /ping handler when set.
	PingHandler http.HandlerFunc
	// TaskHandler replaces the default /task handler when set.
	TaskHandler http.HandlerFunc
	// WorkspaceHandler replaces the default /workspace handler when set.
	WorkspaceHandler http.HandlerFunc
}

// NewBridgeTestServer starts an httptest.Server that mimics the VSTR-Bridge
// extension. All endpoints respond with the happy-path by default; override
// individual handlers via cfg to exercise error paths.
//
// The server is closed automatically when the test ends via t.Cleanup.
//
// Example — default happy-path:
//
//	server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{})
//
// Example — force /ping to return 401:
//
//	server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{
//	    PingHandler: func(w http.ResponseWriter, r *http.Request) {
//	        w.WriteHeader(http.StatusUnauthorized)
//	        json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "invalid token"})
//	    },
//	})
func NewBridgeTestServer(t *testing.T, cfg BridgeHandlerConfig) *httptest.Server {
	t.Helper()

	pingHandler := cfg.PingHandler
	if pingHandler == nil {
		pingHandler = defaultPingHandler()
	}

	taskHandler := cfg.TaskHandler
	if taskHandler == nil {
		taskHandler = defaultTaskHandler()
	}

	workspaceHandler := cfg.WorkspaceHandler
	if workspaceHandler == nil {
		workspaceHandler = defaultWorkspaceHandler()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/task", taskHandler)
	mux.HandleFunc("/workspace", workspaceHandler)

	// Bind on "localhost" (not 127.0.0.1) so server.URL and NewClient's
	// hard-coded "http://localhost:%d" resolve via the same name. Without
	// this, systems where "localhost" prefers ::1 would fail to connect
	// to a listener bound on 127.0.0.1 by httptest.NewServer.
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("NewBridgeTestServer: failed to listen on localhost: %v", err)
	}

	server := httptest.NewUnstartedServer(mux)
	server.Listener.Close()
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	return server
}

// defaultPingHandler returns a handler that satisfies client.TestConnection:
// status 200 + JSON body with secure:true.
func defaultPingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
			"status":            "ok",
			"secure":            true,
			"security_features": []string{"token-auth"},
		})
	}
}

// defaultTaskHandler returns a handler that satisfies client.ExecuteTask:
// status 200 + JSON body with success:true.
func defaultTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
			"success": true,
			"message": "task executed",
		})
	}
}

// defaultWorkspaceHandler returns a handler that satisfies client.ExecuteWorkspace:
// status 200 + JSON body with success:true and an empty results list.
func defaultWorkspaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
			"success": true,
			"results": []interface{}{},
		})
	}
}
