package main

import (
	"bytes"
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
