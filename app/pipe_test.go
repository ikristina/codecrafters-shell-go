package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestShell_parseInput_Pipes(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		input    string
		expected Command
	}{
		"happy path - simple pipe": {
			input: "echo hello | cat",
			expected: Command{
				Name: "echo",
				Args: []string{"hello"},
				Next: &Command{
					Name: "cat",
					Args: []string{},
				},
			},
		},
		"happy path - multiple pipes": {
			input: "echo hello | cat | wc",
			expected: Command{
				Name: "echo",
				Args: []string{"hello"},
				Next: &Command{
					Name: "cat",
					Args: []string{},
					Next: &Command{
						Name: "wc",
						Args: []string{},
					},
				},
			},
		},
		"happy path - pipe with args": {
			input: "ls -la | grep main",
			expected: Command{
				Name: "ls",
				Args: []string{"-la"},
				Next: &Command{
					Name: "grep",
					Args: []string{"main"},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := shell.parseInput(tc.input)

			// Check first command
			if result.Name != tc.expected.Name {
				t.Errorf("expected Name %q, got %q", tc.expected.Name, result.Name)
			}
			if len(result.Args) != len(tc.expected.Args) {
				t.Errorf("expected %d args, got %d", len(tc.expected.Args), len(result.Args))
			}

			// Check next command
			if tc.expected.Next != nil {
				if result.Next == nil {
					t.Fatal("expected Next command, got nil")
				}
				if result.Next.Name != tc.expected.Next.Name {
					t.Errorf("expected Next.Name %q, got %q", tc.expected.Next.Name, result.Next.Name)
				}

				// Check second next command if exists
				if tc.expected.Next.Next != nil {
					if result.Next.Next == nil {
						t.Fatal("expected Next.Next command, got nil")
					}
					if result.Next.Next.Name != tc.expected.Next.Next.Name {
						t.Errorf("expected Next.Next.Name %q, got %q", tc.expected.Next.Next.Name, result.Next.Next.Name)
					}
				}
			}
		})
	}
}

func TestShell_runCommand_Pipes(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		cmd      Command
		input    string
		expected string
	}{
		"happy path - echo pipe cat": {
			cmd: Command{
				Name: "echo",
				Args: []string{"hello"},
				Next: &Command{
					Name: "cat",
					Args: []string{},
				},
			},
			input:    "",
			expected: "hello\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := shell.runCommand(tc.cmd, strings.NewReader(tc.input), &buf)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if buf.String() != tc.expected {
				t.Errorf("expected output %q, got %q", tc.expected, buf.String())
			}
		})
	}
}
