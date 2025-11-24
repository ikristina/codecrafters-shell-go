package main

import (
	"bytes"
	"testing"
)

func TestShell_handleHistory(t *testing.T) {
	shell := NewShell()
	shell.history = []string{"echo hello", "echo world", "invalid_command", "history"}

	var buf bytes.Buffer
	shell.handleHistory(&buf)

	expected := "    1  echo hello\n    2  echo world\n    3  invalid_command\n    4  history\n"
	if buf.String() != expected {
		t.Errorf("expected output:\n%q\ngot:\n%q", expected, buf.String())
	}
}
