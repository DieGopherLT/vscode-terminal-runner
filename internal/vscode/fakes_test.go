// internal/vscode/fakes_test.go
//
// Shared test doubles for package vscode. All types here are package-internal
// (package vscode) and compiled only during testing. No Test* functions are
// defined in this file — only reusable fakes and builders for use in every
// vscode white-box test.
package vscode

import (
	"context"
	"fmt"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/security"
)

// -----------------------------------------------------------------------------
// BridgeClient fake
// -----------------------------------------------------------------------------

// fakeBridgeClient records calls to ExecuteTask and ExecuteWorkspace and returns
// a configurable error for each. Satisfies the BridgeClient interface exactly.
type fakeBridgeClient struct {
	executeTaskErr      error
	executeWorkspaceErr error

	lastTask      *models.Task
	lastWorkspace *models.Workspace
}

func (f *fakeBridgeClient) ExecuteTask(_ context.Context, task models.Task) error {
	f.lastTask = &task
	return f.executeTaskErr
}

func (f *fakeBridgeClient) ExecuteWorkspace(_ context.Context, workspace models.Workspace) error {
	f.lastWorkspace = &workspace
	return f.executeWorkspaceErr
}

// -----------------------------------------------------------------------------
// TaskRepository fake
// -----------------------------------------------------------------------------

// fakeTaskRepository satisfies TaskRepository. If the requested name matches the
// stored task's name the task is returned; otherwise a not-found error is returned.
// Set returnErr to force an error regardless of the name.
type fakeTaskRepository struct {
	task      *models.Task
	returnErr error
}

func (f *fakeTaskRepository) FindByName(name string) (*models.Task, error) {
	if f.returnErr != nil {
		return nil, f.returnErr
	}

	if f.task != nil && f.task.Name == name {
		return f.task, nil
	}

	return nil, fmt.Errorf("task '%s' not found", name)
}

// -----------------------------------------------------------------------------
// WorkspaceRepository fake
// -----------------------------------------------------------------------------

// fakeWorkspaceRepository satisfies WorkspaceRepository with the same contract as
// fakeTaskRepository.
type fakeWorkspaceRepository struct {
	workspace *models.Workspace
	returnErr error
}

func (f *fakeWorkspaceRepository) FindByName(name string) (*models.Workspace, error) {
	if f.returnErr != nil {
		return nil, f.returnErr
	}

	if f.workspace != nil && f.workspace.Name == name {
		return f.workspace, nil
	}

	return nil, fmt.Errorf("workspace '%s' not found", name)
}

// -----------------------------------------------------------------------------
// ProcessNode fake
// -----------------------------------------------------------------------------

// fakeProcessNode represents one node in a synthetic process tree. The parent
// field is typed as ProcessNode (the interface), not *fakeProcessNode, so that a
// root node can store a true nil interface — a nil *fakeProcessNode would be a
// non-nil interface value and would break the nil check in detectParentVSCode.
type fakeProcessNode struct {
	pid     int32
	name    string
	cmdline string
	parent  ProcessNode // nil at the root of the tree
}

func (n *fakeProcessNode) PID() int32               { return n.pid }
func (n *fakeProcessNode) Name() (string, error)    { return n.name, nil }
func (n *fakeProcessNode) Cmdline() (string, error) { return n.cmdline, nil }
func (n *fakeProcessNode) Parent() (ProcessNode, error) {
	return n.parent, nil
}

// buildProcessChain constructs a linear chain of fakeProcessNodes from leaf to
// root. The first element in nodes is the leaf (the process whose PID is
// returned by Getppid); the last element becomes the root (Parent() == nil).
//
// Example — two-level chain, second entry is the VSCode process:
//
//	chain := buildProcessChain(
//	    fakeNodeSpec{pid: 100, name: "shell"},
//	    fakeNodeSpec{pid: 200, name: "code", cmdline: "--folder-uri file:///workspace"},
//	)
func buildProcessChain(nodes ...fakeNodeSpec) map[int32]ProcessNode {
	tree := make(map[int32]ProcessNode, len(nodes))

	for i := len(nodes) - 1; i >= 0; i-- {
		spec := nodes[i]
		node := &fakeProcessNode{
			pid:     spec.pid,
			name:    spec.name,
			cmdline: spec.cmdline,
		}

		if i < len(nodes)-1 {
			// child nodes point to the next (parent) node
			node.parent = tree[nodes[i+1].pid]
		}
		// root node leaves parent as nil (true nil interface)

		tree[spec.pid] = node
	}

	return tree
}

// fakeNodeSpec is the data-only spec used by buildProcessChain.
type fakeNodeSpec struct {
	pid     int32
	name    string
	cmdline string
}

// -----------------------------------------------------------------------------
// ProcessInspector fake
// -----------------------------------------------------------------------------

// fakeProcessInspector satisfies ProcessInspector. Supply a pre-built tree via
// buildProcessChain and set ppid to the leaf PID that detectParentVSCode starts
// from.
type fakeProcessInspector struct {
	ppid int
	tree map[int32]ProcessNode
}

func (f *fakeProcessInspector) Getppid() int { return f.ppid }

func (f *fakeProcessInspector) NewProcess(pid int32) (ProcessNode, error) {
	node, ok := f.tree[pid]
	if !ok {
		return nil, fmt.Errorf("fakeProcessInspector: process %d not found", pid)
	}

	return node, nil
}

// -----------------------------------------------------------------------------
// BridgeInfo builder
// -----------------------------------------------------------------------------

// bridgeInfoBuilder constructs a BridgeInfo that satisfies validateBridgeStructure
// by default. Override individual fields to exercise specific validation paths.
type bridgeInfoBuilder struct {
	info BridgeInfo
}

// newValidBridgeInfo returns a builder whose defaults produce a BridgeInfo that
// passes validateBridgeStructure: Port in range, PID > 0, AuthToken >= 32 bytes,
// Secure = true.
func newValidBridgeInfo() *bridgeInfoBuilder {
	return &bridgeInfoBuilder{
		info: BridgeInfo{
			Port:          8765,
			PID:           12345,
			InstanceID:    1,
			WorkspacePath: "/home/user/projects/myapp",
			WorkspaceName: "myapp",
			Timestamp:     time.Now(),
			AuthToken:     minValidToken(),
			Secure:        true,
		},
	}
}

// WithPort overrides the port.
func (b *bridgeInfoBuilder) WithPort(port int) *bridgeInfoBuilder {
	b.info.Port = port
	return b
}

// WithPID overrides the PID.
func (b *bridgeInfoBuilder) WithPID(pid int) *bridgeInfoBuilder {
	b.info.PID = pid
	return b
}

// WithAuthToken overrides the auth token.
func (b *bridgeInfoBuilder) WithAuthToken(token string) *bridgeInfoBuilder {
	b.info.AuthToken = token
	return b
}

// WithSecure overrides the Secure flag.
func (b *bridgeInfoBuilder) WithSecure(secure bool) *bridgeInfoBuilder {
	b.info.Secure = secure
	return b
}

// WithWorkspacePath overrides the workspace path.
func (b *bridgeInfoBuilder) WithWorkspacePath(path string) *bridgeInfoBuilder {
	b.info.WorkspacePath = path
	return b
}

// WithWorkspaceName overrides the workspace name.
func (b *bridgeInfoBuilder) WithWorkspaceName(name string) *bridgeInfoBuilder {
	b.info.WorkspaceName = name
	return b
}

// WithShortToken sets a token shorter than security.MinTokenLength, which forces
// validateBridgeStructure to return an error. Useful for testing the invalid-token
// path without calculating exact lengths at each call site.
func (b *bridgeInfoBuilder) WithShortToken() *bridgeInfoBuilder {
	b.info.AuthToken = "short"
	return b
}

// Build returns the finished BridgeInfo.
func (b *bridgeInfoBuilder) Build() BridgeInfo {
	return b.info
}

// minValidToken returns a token string of exactly security.MinTokenLength bytes,
// useful when constructing BridgeInfo values in tests that do not use the builder.
func minValidToken() string {
	return fmt.Sprintf("%0*d", security.MinTokenLength, 0)
}
