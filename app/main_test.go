package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShell_parseInput(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		input    string
		expected Command
	}{
		"happy path - simple command": {
			input:    "echo hello",
			expected: Command{Name: "echo", Args: []string{"hello"}, Raw: "echo hello"},
		},
		"happy path - command with single quotes": {
			input:    "echo 'hello example'",
			expected: Command{Name: "echo", Args: []string{"hello example"}, Raw: "echo 'hello example'"},
		},
		"happy path - command with multiple args": {
			input:    "ls -la /tmp",
			expected: Command{Name: "ls", Args: []string{"-la", "/tmp"}, Raw: "ls -la /tmp"},
		},
		"happy path - command only": {
			input:    "pwd",
			expected: Command{Name: "pwd", Args: []string{}, Raw: "pwd"},
		},
		"edge case - empty input": {
			input:    "",
			expected: Command{Name: "", Args: nil, Raw: ""},
		},
		"edge case - whitespace only": {
			input:    "   ",
			expected: Command{Name: "", Args: nil, Raw: ""},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := shell.parseInput(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("expected Name %q, got %q", tc.expected.Name, result.Name)
			}
			if len(result.Args) != len(tc.expected.Args) {
				t.Errorf("expected %d args, got %d", len(tc.expected.Args), len(result.Args))
			}
			for i, arg := range result.Args {
				if i < len(tc.expected.Args) && arg != tc.expected.Args[i] {
					t.Errorf("expected arg[%d] %q, got %q", i, tc.expected.Args[i], arg)
				}
			}
			if result.Raw != tc.expected.Raw {
				t.Errorf("expected Raw %q, got %q", tc.expected.Raw, result.Raw)
			}
		})
	}
}

func TestShell_validateCommand(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		command  string
		expected bool
	}{
		"happy path - builtin command": {
			command:  "echo",
			expected: true,
		},
		"happy path - type builtin": {
			command:  "type",
			expected: true,
		},
		"happy path - system command": {
			command:  "ls",
			expected: true,
		},
		"sad path - invalid command": {
			command:  "nonexistent_command_xyz",
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := shell.validateCommand(tc.command)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestShell_handleType(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		args     []string
		expected string
	}{
		"happy path - builtin command": {
			args:     []string{"echo"},
			expected: "echo is a shell builtin\n",
		},
		"happy path - system command": {
			args:     []string{"ls"},
			expected: "ls is /bin/ls\n",
		},
		"sad path - nonexistent command": {
			args:     []string{"nonexistent_xyz"},
			expected: "nonexistent_xyz: not found\n",
		},
		"sad path - no arguments": {
			args:     []string{},
			expected: "no command found\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			shell.handleType(tc.args)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			result := buf.String()

			// For system commands, we need to be flexible about the path
			if strings.Contains(tc.expected, " is /") && strings.Contains(result, " is /") {
				parts := strings.Split(result, " is ")
				if len(parts) == 2 && strings.HasPrefix(tc.expected, parts[0]+" is /") {
					return // Path found, test passes
				}
			}

			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestShell_handlePwd(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		description string
	}{
		"happy path - get current directory": {
			description: "should return current working directory",
		},
	}

	for name := range tests {
		t.Run(name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			shell.handlePwd()

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			result := strings.TrimSpace(buf.String())

			// Verify we got a valid directory path
			if result == "" {
				t.Error("expected non-empty directory path")
			}
			if !filepath.IsAbs(result) {
				t.Errorf("expected absolute path, got %q", result)
			}
		})
	}
}

func TestShell_handleCd(t *testing.T) {
	shell := NewShell()
	originalDir, _ := os.Getwd()

	// Ensure we return to original directory after tests
	defer os.Chdir(originalDir)

	tests := map[string]struct {
		args        []string
		expectError bool
	}{
		"happy path - change to /tmp": {
			args:        []string{"/tmp"},
			expectError: false,
		},
		"happy path - change to home": {
			args:        []string{},
			expectError: false,
		},
		"happy path - change to home with tilde": {
			args:        []string{"~"},
			expectError: false,
		},
		"sad path - nonexistent directory": {
			args:        []string{"/nonexistent_directory_xyz"},
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Capture stdout for error messages
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			shell.handleCd(tc.args)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if tc.expectError {
				if !strings.Contains(output, "No such file or directory") {
					t.Errorf("expected error message, got %q", output)
				}
			} else {
				if output != "" {
					t.Errorf("expected no output, got %q", output)
				}
			}

			// Reset to original directory for next test
			os.Chdir(originalDir)
		})
	}
}

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

func TestShell_handleExternal(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		command      string
		args         []string
		expectOutput bool
	}{
		"happy path - echo command": {
			command:      "echo",
			args:         []string{"test"},
			expectOutput: true,
		},
		"happy path - ls with invalid path": {
			command:      "ls",
			args:         []string{"/nonexistent"},
			expectOutput: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			shell.handleExternal(tc.command, tc.args)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			result := buf.String()

			hasOutput := result != ""
			if hasOutput != tc.expectOutput {
				t.Errorf("expected output=%v, got output=%v (result: %q)", tc.expectOutput, hasOutput, result)
			}
		})
	}
}
