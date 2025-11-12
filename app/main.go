package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	SingleQuote = '\''
	DoubleQuote = '"'
	Backslash   = '\\'
)

type Shell struct {
	reader   *bufio.Reader
	builtins map[string]struct{}
}

type Command struct {
	Name         string
	Args         []string
	Raw          string
	RedirectFile string
}

func NewShell() *Shell {
	return &Shell{
		reader: bufio.NewReader(os.Stdin),
		builtins: map[string]struct{}{
			"type": {},
			"echo": {},
			"exit": {},
			"pwd":  {},
			"cd":   {},
		},
	}
}

func main() {
	shell := NewShell()
	shell.Run()
}

func (s *Shell) Run() {
	for {
		fmt.Fprint(os.Stdout, "$ ")
		commandLine, err := s.reader.ReadString('\n')
		if err != nil {
			fmt.Println("error capturing the command.")
			return
		}
		commandLine = commandLine[:len(commandLine)-1]

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

	// Check for output redirection
	var redirectFile string
	for i, arg := range args {
		if arg == ">" && i+1 < len(args) || arg == "1>" && i+1 < len(args) {
			redirectFile = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	commandName := strings.TrimSpace(args[0])
	commandArgs := args[1:]

	return Command{
		Name:         commandName,
		Args:         commandArgs,
		Raw:          input,
		RedirectFile: redirectFile,
	}
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
	if _, ok := s.builtins[name]; ok {
		return true
	}
	return s.isInPath(name) != ""
}

func (s *Shell) parseQuotedArgs(input string) []string {
	args := []string{}
	inQuotes := false
	quoteChar := byte(0)
	currentArg := ""

	for i := 0; i < len(input); i++ {
		c := input[i]

		if c == Backslash && i+1 < len(input) && quoteChar != SingleQuote {
			nextChar := input[i+1]
			if quoteChar == DoubleQuote {
				// Inside double quotes, only escape specific characters
				switch nextChar {
				case '\\':
					currentArg += "\\"
					i++
				case '"':
					currentArg += "\""
					i++
				default:
					// Keep backslash literal for other characters
					currentArg += string(c)
				}
			} else {
				// Outside quotes, escape next character
				currentArg += string(nextChar)
				i++
			}
		} else if !inQuotes && (c == SingleQuote || c == DoubleQuote) {
			inQuotes = true
			quoteChar = c
		} else if inQuotes && c == quoteChar {
			inQuotes = false
			quoteChar = 0
		} else if c == ' ' && !inQuotes {
			if currentArg != "" {
				args = append(args, currentArg)
				currentArg = ""
			}
		} else {
			currentArg += string(c)
		}
	}

	if currentArg != "" {
		args = append(args, currentArg)
	}
	return args
}

func (s *Shell) isInPath(command string) string {
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		file := filepath.Join(path, command)
		info, err := os.Stat(file)
		if err == nil && info.Mode()&0o111 != 0 {
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
	output := ""
	if cmd.Args != nil {
		output = strings.Join(cmd.Args, " ") + "\n"
	} else {
		output = "\n"
	}

	if cmd.RedirectFile != "" {
		os.WriteFile(cmd.RedirectFile, []byte(output), 0644)
	} else {
		fmt.Print(output)
	}
}

func (s *Shell) handleType(args []string) {
	if len(args) == 0 {
		fmt.Println("no command found")
		return
	}
	commandName := args[0]
	filePath := s.isInPath(commandName)
	if _, ok := s.builtins[commandName]; ok {
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
	var dir string
	if len(args) != 0 {
		dir = args[0]
	}
	if dir == "~" {
		dir = os.Getenv("HOME")
	}
	if dir == "" {
		dir = os.Getenv("HOME")
	}
	err := os.Chdir(dir)
	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", dir)
		return
	}
}

func (s *Shell) handleExternal(cmd Command) {
	execCmd := exec.Command(cmd.Name, cmd.Args...)
	
	if cmd.RedirectFile != "" {
		// Redirect stdout to file, stderr stays on terminal
		execCmd.Stderr = os.Stderr
		output, _ := execCmd.Output()
		os.WriteFile(cmd.RedirectFile, output, 0644)
	} else {
		// Both stdout and stderr to terminal
		output, _ := execCmd.CombinedOutput()
		fmt.Print(string(output))
	}
}
