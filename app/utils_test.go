package main

import (
	"path/filepath"
	"testing"
)

func TestShell_isInPath(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		command     string
		expectFound bool
	}{
		"happy path - ls command exists": {
			command:     "ls",
			expectFound: true,
		},
		"happy path - cat command exists": {
			command:     "cat",
			expectFound: true,
		},
		"sad path - nonexistent command": {
			command:     "nonexistent_command_xyz_123",
			expectFound: false,
		},
		"happy path - echo command exists": {
			command:     "echo",
			expectFound: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := shell.isInPath(tc.command)
			found := result != ""

			if found != tc.expectFound {
				t.Errorf("expected found=%v, got found=%v (result: %q)", tc.expectFound, found, result)
			}

			if found && !filepath.IsAbs(result) {
				t.Errorf("expected absolute path, got %q", result)
			}
		})
	}
}
