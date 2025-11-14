package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
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

// Shell represents a POSIX-compliant shell with readline support
type Shell struct {
	rl          *readline.Instance
	allCommands []string
}

var builtinCommands = map[string]struct{}{
	"type": {},
	"echo": {},
	"exit": {},
	"pwd":  {},
	"cd":   {},
}

type Command struct {
	Name           string
	Args           []string
	RedirectFile   string
	RedirectStderr bool
	AppendMode     bool
}

// BellListener implements readline.Listener to ring a bell on TAB press
type BellListener struct{}

// OnChange is called on every keypress
func (l *BellListener) OnChange(line []rune, pos int, key rune) ([]rune, int, bool) {
	if key == readline.CharTab {
		fmt.Print("\x07")
	}
	return line, pos, false
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

	shell := &Shell{allCommands: allCommands}

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
	return shell
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

func main() {
	shell := NewShell()
	shell.Run()
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

		if err = s.executeCommand(commandLine); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func (s *Shell) parseInput(input string) Command {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return Command{}
	}

	var args []string

	if strings.ContainsAny(input, "'\"\\") {
		args = s.parseQuotedArgs(input)
	} else {
		args = strings.Fields(input)
	}

	for i, arg := range args {
		if i+1 >= len(args) {
			continue
		}
		switch arg {
		case ">", "1>":
			redirectFile := args[i+1]
			args = append(args[:i], args[i+2:]...)
			return Command{Name: strings.TrimSpace(args[0]), Args: args[1:], RedirectFile: redirectFile}
		case "2>":
			redirectFile := args[i+1]
			args = append(args[:i], args[i+2:]...)
			return Command{Name: strings.TrimSpace(args[0]), Args: args[1:], RedirectFile: redirectFile, RedirectStderr: true}
		case ">>", "1>>":
			redirectFile := args[i+1]
			args = append(args[:i], args[i+2:]...)
			return Command{Name: strings.TrimSpace(args[0]), Args: args[1:], RedirectFile: redirectFile, AppendMode: true}
		case "2>>":
			redirectFile := args[i+1]
			args = append(args[:i], args[i+2:]...)
			return Command{Name: strings.TrimSpace(args[0]), Args: args[1:], RedirectFile: redirectFile, RedirectStderr: true, AppendMode: true}
		}
	}

	return Command{Name: strings.TrimSpace(args[0]), Args: args[1:]}
}

func (s *Shell) executeCommand(commandLine string) error {
	cmd := s.parseInput(commandLine)
	if cmd.Name == "" {
		return nil
	}

	if !s.validateCommand(cmd.Name) {
		fmt.Printf("%s: command not found\n", commandLine)
		return nil
	}

	switch cmd.Name {
	case "exit":
		s.handleExit(commandLine, cmd.Args)
	case "echo":
		s.handleEcho(cmd)
	case "type":
		s.handleType(cmd.Args)
	case "pwd":
		s.handlePwd()
	case "cd":
		s.handleCd(cmd.Args)
	default:
		s.handleExternal(cmd)
	}
	return nil
}

func (s *Shell) validateCommand(name string) bool {
	if _, ok := builtinCommands[name]; ok {
		return true
	}
	return s.isInPath(name) != ""
}

func (s *Shell) parseQuotedArgs(input string) []string {
	var args []string
	var currentArg strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(input); i++ {
		c := input[i]

		if c == Backslash && i+1 < len(input) && quoteChar != SingleQuote {
			nextChar := input[i+1]
			if quoteChar == DoubleQuote {
				if nextChar == '\\' || nextChar == '"' {
					currentArg.WriteByte(nextChar)
					i++
				} else {
					currentArg.WriteByte(c)
				}
			} else {
				currentArg.WriteByte(nextChar)
				i++
			}
		} else if !inQuotes && (c == SingleQuote || c == DoubleQuote) {
			inQuotes = true
			quoteChar = c
		} else if inQuotes && c == quoteChar {
			inQuotes = false
			quoteChar = 0
		} else if c == ' ' && !inQuotes {
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		} else {
			currentArg.WriteByte(c)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}
	return args
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

func (s *Shell) handleExit(commandLine string, args []string) {
	if len(args) == 0 {
		os.Exit(0)
		return
	}
	v, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("incorrect command arguments: %s", commandLine)
		return
	}
	os.Exit(v)
}

func (s *Shell) handleEcho(cmd Command) {
	output := strings.Join(cmd.Args, " ") + "\n"
	if cmd.RedirectFile != "" && !cmd.RedirectStderr {
		s.writeToFile(cmd.RedirectFile, []byte(output), cmd.AppendMode)
	} else {
		if cmd.RedirectFile != "" {
			s.writeToFile(cmd.RedirectFile, []byte(""), cmd.AppendMode)
		}
		fmt.Print(output)
	}
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

func (s *Shell) handleType(args []string) {
	if len(args) == 0 {
		fmt.Println("no command found")
		return
	}
	commandName := args[0]
	filePath := s.isInPath(commandName)
	if _, ok := builtinCommands[commandName]; ok {
		fmt.Printf("%s is a shell builtin\n", commandName)
	} else if filePath != "" {
		fmt.Printf("%[1]s is %[2]s\n", commandName, filePath)
	} else {
		fmt.Printf("%s: not found\n", commandName)
	}
}

func (s *Shell) handlePwd() {
	dir, err := os.Getwd()
	if err == nil {
		fmt.Printf("%s\n", dir)
	} else {
		fmt.Printf("error getting pwd %s\n", err)
	}
}

func (s *Shell) handleCd(args []string) {
	dir := os.Getenv("HOME")
	if len(args) > 0 && args[0] != "~" {
		dir = args[0]
	}
	if err := os.Chdir(dir); err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", dir)
	}
}

func (s *Shell) handleExternal(cmd Command) {
	execCmd := exec.Command(cmd.Name, cmd.Args...)

	if cmd.RedirectFile != "" {
		if cmd.RedirectStderr {
			execCmd.Stdout = os.Stdout
			if stderr, err := execCmd.StderrPipe(); err == nil {
				if execCmd.Start() == nil {
					if data, err := io.ReadAll(stderr); err == nil {
						s.writeToFile(cmd.RedirectFile, data, cmd.AppendMode)
					}
					_ = execCmd.Wait()
				}
			}
		} else {
			execCmd.Stderr = os.Stderr
			output, _ := execCmd.Output()
			s.writeToFile(cmd.RedirectFile, output, cmd.AppendMode)
		}
	} else {
		output, _ := execCmd.CombinedOutput()
		fmt.Print(string(output))
	}
}
