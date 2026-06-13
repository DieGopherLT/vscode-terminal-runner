package repository

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenSource_stdinWhenDash(t *testing.T) {
	reader, err := OpenSource("-")
	if err != nil {
		t.Fatalf("OpenSource(\"-\") returned unexpected error: %v", err)
	}
	if reader == nil {
		t.Fatal("OpenSource(\"-\") returned nil reader")
	}
	// Closing a NopCloser wrapping stdin must not fail and must not close the real stdin.
	if err := reader.Close(); err != nil {
		t.Fatalf("Close() on stdin source returned unexpected error: %v", err)
	}
}

func TestOpenSource_opensExistingFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(f, []byte(`[]`), 0666); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	reader, err := OpenSource(f)
	if err != nil {
		t.Fatalf("OpenSource returned unexpected error for an existing file: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll returned unexpected error: %v", err)
	}
	if string(data) != "[]" {
		t.Fatalf("expected file content %q, got %q", "[]", string(data))
	}
}

func TestOpenSource_returnsErrorForMissingFile(t *testing.T) {
	_, err := OpenSource("/nonexistent/path/to/file.json")
	if err == nil {
		t.Fatal("OpenSource should return an error for a nonexistent path")
	}
}
