package cmd

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersionFrom(t *testing.T) {
	tests := []struct {
		name     string
		info     *debug.BuildInfo
		ok       bool
		expected string
	}{
		{
			name:     "build info not available",
			info:     nil,
			ok:       false,
			expected: "dev",
		},
		{
			name:     "devel build",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:       true,
			expected: "dev",
		},
		{
			name:     "empty version",
			info:     &debug.BuildInfo{Main: debug.Module{Version: ""}},
			ok:       true,
			expected: "dev",
		},
		{
			name:     "tagged release",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "v1.2.0"}},
			ok:       true,
			expected: "v1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveVersionFrom(tt.info, tt.ok)
			if got != tt.expected {
				t.Errorf("resolveVersionFrom() = %q, want %q", got, tt.expected)
			}
		})
	}
}
