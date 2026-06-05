package client

import (
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

// clientForServer creates a Client pointed at the given httptest.Server.
// It extracts the port from the server URL and calls NewClient so that
// the production constructor is exercised in every test.
//
// NewBridgeTestServer binds its listener on "localhost" (not 127.0.0.1) to
// match the host NewClient dials, so the IPv4/IPv6 resolution is consistent.
// The permanent fix — accepting a base URL in NewClient — belongs to the
// adapter layer.
//
// The returned client has no auth token loaded. Call c.LoadAuthFromToken with
// testutils.ValidTestToken for tests that exercise authenticated paths.
//
// Usage:
//
//	server := testutils.NewBridgeTestServer(t, testutils.BridgeHandlerConfig{})
//	c := clientForServer(t, server)
func clientForServer(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("clientForServer: failed to parse server URL %q: %v", server.URL, err)
	}

	portStr := parsed.Port()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("clientForServer: non-numeric port in server URL %q: %v", server.URL, err)
	}

	return NewClient(port)
}
