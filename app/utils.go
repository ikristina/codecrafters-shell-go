package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

const (
	SingleQuote = '\''
	DoubleQuote = '"'
	Backslash   = '\\'

	// FilePermission is 0o644 (rw-r--r--): owner can read/write, others can read
	FilePermission = 0o644
	// ExecPermission is 0o111 (--x--x--x): checks if any execute bit is set
	ExecPermission = 0o111
)

// BellListener implements readline.Listener to ring a bell on TAB press
type BellListener struct{}

// OnChange is called on every keypress
func (l *BellListener) OnChange(line []rune, pos int, key rune) ([]rune, int, bool) {
	if key == readline.CharTab {
		fmt.Print("\x07")
	}
	return line, pos, false
}

func (s *Shell) isInPath(command string) string {
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		file := filepath.Join(path, command)
		info, err := os.Stat(file)
		if err == nil && info.Mode()&ExecPermission != 0 {
			return file
		}
	}
	return ""
}

func (s *Shell) writeToFile(path string, data []byte, append bool) {
	if append {
		if f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, FilePermission); err == nil {
			_, _ = f.Write(data)
			_ = f.Close()
		}
	} else {
		_ = os.WriteFile(path, data, FilePermission)
	}
}
