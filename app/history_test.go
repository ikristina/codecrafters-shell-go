package main

import (
	"bytes"
	"os"
	"testing"
)

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
