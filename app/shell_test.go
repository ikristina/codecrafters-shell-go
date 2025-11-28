package main

import (
	"testing"
)

func TestShell_Do(t *testing.T) {
	shell := &Shell{
		allCommands: []string{"cat", "cd", "echo", "exit", "ls", "pwd", "type"},
	}

	tests := map[string]struct {
		input          string
		expectedCount  int
		expectedSuffix string
	}{
		"single match": {
			input:          "ech",
			expectedCount:  1,
			expectedSuffix: "o ",
		},
		"multiple matches": {
			input:         "c",
			expectedCount: 0, // Returns nil for multiple matches
		},
		"no matches": {
			input:         "xyz",
			expectedCount: 0,
		},
		"exact match": {
			input:          "echo",
			expectedCount:  1,
			expectedSuffix: " ",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, _ := shell.Do([]rune(tc.input), len(tc.input))
			if len(result) != tc.expectedCount {
				t.Errorf("expected %d results, got %d", tc.expectedCount, len(result))
			}
			if tc.expectedCount == 1 && len(result) > 0 && string(result[0]) != tc.expectedSuffix {
				t.Errorf("expected suffix %q, got %q", tc.expectedSuffix, string(result[0]))
			}
		})
	}
}
