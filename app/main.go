package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

const (
	SingleQuote = '\''
	DoubleQuote = '"'
	Backslash   = '\\'
)

type Shell struct {
	rl       *readline.Instance
	builtins map[string]struct{}
}

type Command struct {
	Name           string
	Args           []string
	Raw            string
	RedirectFile   string
	RedirectStderr bool
	AppendMode     bool
}

// BellListener is to implement Listener interface from readline library
type BellListener struct{}

// OnChange method gets called with every key press.
func (l *BellListener) OnChange(line []rune, pos int, key rune) ([]rune, int, bool) {
	if key == readline.CharTab {
		fmt.Print("\x07") // will ring a bell in the case of successful and unsuccessful autocomplete.
	}
	return line, pos, false
}

func NewShell() *Shell {
	items := []readline.PrefixCompleterInterface{
		readline.PcItem("type"),
		readline.PcItem("echo"),
		readline.PcItem("exit"),
		readline.PcItem("pwd"),
		readline.PcItem("cd"),
	}

	// Add executables from PATH
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		files, _ := os.ReadDir(path)
		for _, file := range files {
			if !file.IsDir() {
				items = append(items, readline.PcItem(file.Name()))
			}
		}
	}
	// Create autocomplete function
	completer := readline.NewPrefixCompleter(items...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		Listener:        &BellListener{},
	})
	if err != nil {
		panic(err)
	}

	return &Shell{
		rl: rl,
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

	// Check for output redirection
	var redirectFile string
	var redirectStderr bool
	var appendMode bool
	for i, arg := range args {
		if (arg == ">" || arg == "1>") && i+1 < len(args) {
			redirectFile = args[i+1]
			redirectStderr = false
			appendMode = false
			args = append(args[:i], args[i+2:]...)
			break
		} else if arg == "2>" && i+1 < len(args) {
			redirectFile = args[i+1]
			redirectStderr = true
			appendMode = false
			args = append(args[:i], args[i+2:]...)
			break
		} else if (arg == ">>" || arg == "1>>") && i+1 < len(args) {
			redirectFile = args[i+1]
			redirectStderr = false
			appendMode = true
			args = append(args[:i], args[i+2:]...)
			break
		} else if arg == "2>>" && i+1 < len(args) {
			redirectFile = args[i+1]
			redirectStderr = true
			appendMode = true
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	commandName := strings.TrimSpace(args[0])
	commandArgs := args[1:]

	return Command{
		Name:           commandName,
		Args:           commandArgs,
		Raw:            input,
		RedirectFile:   redirectFile,
		RedirectStderr: redirectStderr,
		AppendMode:     appendMode,
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
		if cmd.RedirectStderr {
			// Redirecting stderr - create/append empty file, print stdout to terminal
			if cmd.AppendMode {
				f, _ := os.OpenFile(cmd.RedirectFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				f.Close()
			} else {
				os.WriteFile(cmd.RedirectFile, []byte(""), 0o644)
			}
			fmt.Print(output)
		} else {
			// Redirect stdout to file
			if cmd.AppendMode {
				f, _ := os.OpenFile(cmd.RedirectFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				f.WriteString(output)
				f.Close()
			} else {
				os.WriteFile(cmd.RedirectFile, []byte(output), 0o644)
			}
		}
	} else {
		// Print to terminal
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
		if cmd.RedirectStderr {
			// Redirect stderr to file, stdout stays on terminal
			execCmd.Stdout = os.Stdout
			stderrPipe, _ := execCmd.StderrPipe()
			execCmd.Start()
			stderrData, _ := io.ReadAll(stderrPipe)
			if cmd.AppendMode {
				f, _ := os.OpenFile(cmd.RedirectFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				f.Write(stderrData)
				f.Close()
			} else {
				os.WriteFile(cmd.RedirectFile, stderrData, 0o644)
			}
			execCmd.Wait()
		} else {
			// Redirect stdout to file, stderr stays on terminal
			execCmd.Stderr = os.Stderr
			output, _ := execCmd.Output()
			if cmd.AppendMode {
				f, _ := os.OpenFile(cmd.RedirectFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				f.Write(output)
				f.Close()
			} else {
				os.WriteFile(cmd.RedirectFile, output, 0o644)
			}
		}
	} else {
		// Both stdout and stderr to terminal
		output, _ := execCmd.CombinedOutput()
		fmt.Print(string(output))
	}
}
