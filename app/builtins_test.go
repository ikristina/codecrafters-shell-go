package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

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
		"happy path - pwd builtin": {
			args:     []string{"pwd"},
			expected: "pwd is a shell builtin\n",
		},
		"happy path - cd builtin": {
			args:     []string{"cd"},
			expected: "cd is a shell builtin\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			shell.handleType(tc.args, w)

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
			var buf bytes.Buffer
			shell.handlePwd(&buf)
			result := strings.TrimSpace(buf.String())

			// Verify we got a valid directory path
			if result == "" {
				t.Error("expected non-empty directory path")
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
		"happy path - relative path": {
			args:        []string{"."},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			shell.handleCd(tc.args, &buf)
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

func TestShell_handleExternal(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		cmd          Command
		expectOutput bool
	}{
		"happy path - echo command": {
			cmd:          Command{Name: "echo", Args: []string{"test"}},
			expectOutput: true,
		},
		"happy path - ls with invalid path": {
			cmd:          Command{Name: "ls", Args: []string{"/nonexistent"}},
			expectOutput: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			rOut, wOut, _ := os.Pipe()
			rErr, wErr, _ := os.Pipe()
			os.Stdout = wOut
			os.Stderr = wErr

			shell.handleExternal(tc.cmd, strings.NewReader(""), wOut)

			wOut.Close()
			wErr.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var bufOut bytes.Buffer
			io.Copy(&bufOut, rOut)
			var bufErr bytes.Buffer
			io.Copy(&bufErr, rErr)

			result := bufOut.String() + bufErr.String()

			hasOutput := result != ""
			if hasOutput != tc.expectOutput {
				t.Errorf("expected output=%v, got output=%v (result: %q)", tc.expectOutput, hasOutput, result)
			}
		})
	}
}

func TestShell_handleHistory(t *testing.T) {
	shell := NewShell()
	shell.history = []string{"echo hello", "echo world", "invalid_command", "history"}

	var buf bytes.Buffer
	shell.handleHistory([]string{}, &buf)

	expected := "    1  echo hello\n    2  echo world\n    3  invalid_command\n    4  history\n"
	if buf.String() != expected {
		t.Errorf("expected output:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestShell_handleHistory_Limit(t *testing.T) {
	shell := NewShell()
	shell.history = []string{"cmd1", "cmd2", "cmd3", "cmd4", "history 2"}

	var buf bytes.Buffer
	shell.handleHistory([]string{"2"}, &buf)

	expected := "    4  cmd4\n    5  history 2\n"
	if buf.String() != expected {
		t.Errorf("expected output:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestShell_handleHistory_Read(t *testing.T) {
	shell := NewShell()
	shell.history = []string{"cmd1"}

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "history")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := []byte("file_cmd1\nfile_cmd2\n")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	shell.handleHistory([]string{"-r", tmpfile.Name()}, &buf)

	// Verify history was updated
	if len(shell.history) != 3 {
		t.Errorf("expected 3 history entries, got %d", len(shell.history))
	}
	if shell.history[1] != "file_cmd1" {
		t.Errorf("expected history[1] to be 'file_cmd1', got %q", shell.history[1])
	}
	if shell.history[2] != "file_cmd2" {
		t.Errorf("expected history[2] to be 'file_cmd2', got %q", shell.history[2])
	}
}

func TestShell_handleHistory_Write(t *testing.T) {
	shell := NewShell()
	shell.history = []string{"cmd1", "cmd2"}

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "history_write")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	var buf bytes.Buffer
	shell.handleHistory([]string{"-w", tmpfile.Name()}, &buf)

	// Read file content
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	expected := "cmd1\ncmd2\n"
	if string(content) != expected {
		t.Errorf("expected file content %q, got %q", expected, string(content))
	}
}

func TestShell_handleHistory_Append(t *testing.T) {
	shell := NewShell()
	shell.history = []string{"cmd3", "cmd4"}

	// Create a temporary file with existing content
	tmpfile, err := os.CreateTemp("", "history_append")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	initialContent := []byte("cmd1\ncmd2\n")
	if _, err := tmpfile.Write(initialContent); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	var buf bytes.Buffer
	shell.handleHistory([]string{"-a", tmpfile.Name()}, &buf)

	// Read file content
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	expected := "cmd1\ncmd2\ncmd3\ncmd4\n"
	if string(content) != expected {
		t.Errorf("expected file content %q, got %q", expected, string(content))
	}
}
