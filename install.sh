#!/bin/sh

set -e

if ! command -v go >/dev/null 2>&1; then
    echo "Error: Go is required to install vstr." >&2
    echo "Install it from https://go.dev/dl/" >&2
    exit 1
fi

echo "Installing vstr..."

GOBIN_DIR="$(go env GOPATH)/bin"

GOBIN="$GOBIN_DIR" go install github.com/DieGopherLT/vscode-terminal-runner@latest

# go install names the binary after the last segment of the module path
if [ -f "$GOBIN_DIR/vscode-terminal-runner" ]; then
    mv "$GOBIN_DIR/vscode-terminal-runner" "$GOBIN_DIR/vstr"
fi

echo ""
echo "vstr installed to $GOBIN_DIR/vstr"
echo ""
echo "If $GOBIN_DIR is not in your PATH, add this to your shell config:"
echo "  export PATH=\"\$PATH:$GOBIN_DIR\""
echo ""
echo "Then run:"
echo "  vstr setup"
