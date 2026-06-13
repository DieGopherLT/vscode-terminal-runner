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
	result := statusCheckResult{
		extensionInstalled: true,
		bridgePort:         51234,
		bridgeWorkspace:    "workspace",
		bridgeErr:          nil,
	}

	allOK := renderStatus(result)
	if !allOK {
		t.Error("expected allOK=true for successful result")
	}
}
