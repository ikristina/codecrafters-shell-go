package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

// Shell represents a POSIX-compliant shell with readline support
type Shell struct {
	rl                   *readline.Instance
	allCommands          []string
	history              []string
	historyAppendedCount int
}

// NewShell creates and initializes a new Shell instance with autocomplete support
func NewShell() *Shell {
	paths := strings.Split(os.Getenv("PATH"), ":")
	executables := make(map[string]struct{})

	for _, path := range paths {
		if files, err := os.ReadDir(path); err == nil {
			for _, file := range files {
				if !file.IsDir() {
					executables[file.Name()] = struct{}{}
				}
			}
		}
	}

	for cmd := range builtinCommands {
		executables[cmd] = struct{}{}
	}

	allCommands := make([]string, 0, len(executables))
	for cmd := range executables {
		allCommands = append(allCommands, cmd)
	}
	sort.Strings(allCommands)

	shell := &Shell{
		allCommands: allCommands,
		history:     []string{},
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    shell,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		Listener:        &BellListener{},
	})
	if err != nil {
		panic(err)
	}

	shell.rl = rl

	// Load history from HISTFILE if set
	if histfile := os.Getenv("HISTFILE"); histfile != "" {
		if content, err := os.ReadFile(histfile); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if line != "" {
					shell.history = append(shell.history, line)
				}
			}
		}
	}

	return shell
}

// Run starts the shell's REPL (Read-Eval-Print Loop)
func (s *Shell) Run() {
	defer s.rl.Close()

	for {
		commandLine, err := s.rl.Readline()
		if err != nil {
			fmt.Println("\x07")
			return
		}

		s.history = append(s.history, commandLine)
		if err = s.executeCommand(commandLine); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

// Do implements readline.AutoCompleter interface
func (s *Shell) Do(line []rune, pos int) ([][]rune, int) {
	lineStr := string(line[:pos])
	matches := []string{}
	for _, cmd := range s.allCommands {
		if strings.HasPrefix(cmd, lineStr) {
			matches = append(matches, cmd)
		}
	}

	if len(matches) == 0 {
		return nil, len(lineStr)
	}

	if len(matches) == 1 {
		suffix := matches[0][len(lineStr):] + " "
		return [][]rune{[]rune(suffix)}, len(lineStr)
	}

	// Find longest common prefix
	commonPrefix := matches[0]
	for _, match := range matches[1:] {
		for i := 0; i < len(commonPrefix) && i < len(match); i++ {
			if commonPrefix[i] != match[i] {
				commonPrefix = commonPrefix[:i]
				break
			}
		}
		if len(match) < len(commonPrefix) {
			commonPrefix = match
		}
	}

	// If common prefix is longer than what user typed, complete to it
	if len(commonPrefix) > len(lineStr) {
		suffix := commonPrefix[len(lineStr):]
		return [][]rune{[]rune(suffix)}, len(lineStr)
	}

	// Otherwise show all matches
	fmt.Println()
	for i, match := range matches {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(match)
	}
	fmt.Println()
	fmt.Printf("$ %s", lineStr)

	return nil, len(lineStr)
}
