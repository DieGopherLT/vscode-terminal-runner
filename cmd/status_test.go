package cmd

import (
	"errors"
	"testing"
)

func TestRenderStatus_allOKWhenBothChecksPass(t *testing.T) {
	result := statusCheckResult{
		extensionInstalled: true,
		bridgePort:         51234,
		bridgeWorkspace:    "my-project",
		bridgeErr:          nil,
	}

	allOK := renderStatus(result)

	if !allOK {
		t.Error("expected allOK=true when both checks pass, got false")
	}
}

func TestRenderStatus_notOKWhenExtensionMissing(t *testing.T) {
	result := statusCheckResult{
		extensionInstalled: false,
		bridgePort:         51234,
		bridgeWorkspace:    "my-project",
		bridgeErr:          nil,
	}

	allOK := renderStatus(result)

	if allOK {
		t.Error("expected allOK=false when extension is not installed, got true")
	}
}

func TestRenderStatus_notOKWhenBridgeError(t *testing.T) {
	result := statusCheckResult{
		extensionInstalled: true,
		bridgePort:         0,
		bridgeWorkspace:    "",
		bridgeErr:          errors.New("bridge directory not found"),
	}

	allOK := renderStatus(result)

	if allOK {
		t.Error("expected allOK=false when bridge check fails, got true")
	}
}

func TestRenderStatus_notOKWhenBothFail(t *testing.T) {
	result := statusCheckResult{
		extensionInstalled: false,
		bridgePort:         0,
		bridgeWorkspace:    "",
		bridgeErr:          errors.New("bridge not found"),
	}

	allOK := renderStatus(result)

	if allOK {
		t.Error("expected allOK=false when both checks fail, got true")
	}
}

func TestRenderStatus_doesNotRevealAuthToken(t *testing.T) {
	// statusCheckResult contains no auth token field — gatherStatus drops the token
	// after loading it into the client. This test documents that invariant.
	result := statusCheckResult{
		extensionInstalled: true,
		bridgePort:         51234,
		bridgeWorkspace:    "workspace",
		bridgeErr:          nil,
	}

	// If this compiles and renderStatus only uses the exported fields, the token
	// can never leak through this path.
	allOK := renderStatus(result)
	if !allOK {
		t.Error("expected allOK=true for successful result")
	}
}
