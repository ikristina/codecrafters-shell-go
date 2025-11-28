package main

import (
	"bytes"
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
			expected: Command{Name: "echo", Args: []string{"hello"}},
		},
		"happy path - command with single quotes": {
			input:    "echo 'hello example'",
			expected: Command{Name: "echo", Args: []string{"hello example"}},
		},
		"happy path - command with multiple args": {
			input:    "ls -la /tmp",
			expected: Command{Name: "ls", Args: []string{"-la", "/tmp"}},
		},
		"happy path - command only": {
			input:    "pwd",
			expected: Command{Name: "pwd", Args: []string{}},
		},
		"edge case - empty input": {
			input:    "",
			expected: Command{Name: "", Args: nil},
		},
		"edge case - whitespace only": {
			input:    "   ",
			expected: Command{Name: "", Args: nil},
		},
		"happy path - stdout redirect": {
			input:    "echo hello > file.txt",
			expected: Command{Name: "echo", Args: []string{"hello"}, RedirectFile: "file.txt", RedirectStderr: false, AppendMode: false},
		},
		"happy path - stderr redirect": {
			input:    "cat file 2> error.txt",
			expected: Command{Name: "cat", Args: []string{"file"}, RedirectFile: "error.txt", RedirectStderr: true, AppendMode: false},
		},
		"happy path - stdout append": {
			input:    "echo hello >> file.txt",
			expected: Command{Name: "echo", Args: []string{"hello"}, RedirectFile: "file.txt", RedirectStderr: false, AppendMode: true},
		},
		"happy path - stderr append": {
			input:    "cat file 2>> error.txt",
			expected: Command{Name: "cat", Args: []string{"file"}, RedirectFile: "error.txt", RedirectStderr: true, AppendMode: true},
		},
		"happy path - command with quotes and redirect": {
			input:    `echo "hello world" > file.txt`,
			expected: Command{Name: "echo", Args: []string{"hello world"}, RedirectFile: "file.txt", RedirectStderr: false, AppendMode: false},
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
		"happy path - pwd builtin": {
			command:  "pwd",
			expected: true,
		},
		"happy path - cd builtin": {
			command:  "cd",
			expected: true,
		},
		"happy path - exit builtin": {
			command:  "exit",
			expected: true,
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

func TestShell_parseQuotedArgs(t *testing.T) {
	shell := NewShell()

	tests := map[string]struct {
		input    string
		expected []string
	}{
		"happy path - simple args": {
			input:    "echo hello world",
			expected: []string{"echo", "hello", "world"},
		},
		"happy path - double quotes": {
			input:    `echo "hello world"`,
			expected: []string{"echo", "hello world"},
		},
		"happy path - single quotes": {
			input:    `echo 'hello world'`,
			expected: []string{"echo", "hello world"},
		},
		"happy path - escaped double quotes": {
			input:    `echo "hello \"world\""`,
			expected: []string{"echo", `hello "world"`},
		},
		"happy path - backslash in single quotes": {
			input:    `echo 'hello\'world'`,
			expected: []string{"echo", `hello\world`},
		},
		"happy path - escaped space": {
			input:    `echo hello\ world`,
			expected: []string{"echo", "hello world"},
		},
		"happy path - escaped backslash in double quotes": {
			input:    `echo "hello\\"`,
			expected: []string{"echo", `hello\`},
		},
		"happy path - mixed quotes": {
			input:    `echo "hello" 'world'`,
			expected: []string{"echo", "hello", "world"},
		},
		"happy path - empty quotes": {
			input:    `echo ""`,
			expected: []string{"echo"},
		},
		"edge case - multiple spaces": {
			input:    `echo   hello    world`,
			expected: []string{"echo", "hello", "world"},
		},
		"edge case - trailing space": {
			input:    `echo hello `,
			expected: []string{"echo", "hello"},
		},
		"edge case - leading space": {
			input:    ` echo hello`,
			expected: []string{"echo", "hello"},
		},
		"happy path - nested quotes": {
			input:    `echo "it's working"`,
			expected: []string{"echo", "it's working"},
		},
		"happy path - adjacent quotes": {
			input:    `echo "hello""world"`,
			expected: []string{"echo", "helloworld"},
		},
		"edge case - only spaces between quotes": {
			input:    `echo "  "`,
			expected: []string{"echo", "  "},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := shell.parseQuotedArgs(tc.input)
			if len(result) != len(tc.expected) {
				t.Errorf("expected %d args, got %d", len(tc.expected), len(result))
				return
			}
			for i := range result {
				if result[i] != tc.expected[i] {
					t.Errorf("arg[%d]: expected %q, got %q", i, tc.expected[i], result[i])
				}
			}
		})
	}
}

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
