package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Command struct {
	Name           string
	Args           []string
	RedirectFile   string
	RedirectStderr bool
	AppendMode     bool
	Next           *Command
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
		case "|":
			nextCmd := s.parseInput(strings.Join(args[i+1:], " "))
			return Command{Name: strings.TrimSpace(args[0]), Args: args[1:i], Next: &nextCmd}
		}
	}

	return Command{Name: strings.TrimSpace(args[0]), Args: args[1:]}
}

func (s *Shell) executeCommand(commandLine string) error {
	cmd := s.parseInput(commandLine)
	return s.runCommand(cmd, os.Stdin, os.Stdout)
}

func (s *Shell) runCommand(cmd Command, stdin io.Reader, stdout io.Writer) error {
	if cmd.Name == "" {
		return nil
	}

	if cmd.Next != nil {
		r, w, err := os.Pipe()
		if err != nil {
			return err
		}

		go func() {
			currentCmd := cmd
			currentCmd.Next = nil
			s.runCommand(currentCmd, stdin, w)
			w.Close()
		}()

		return s.runCommand(*cmd.Next, r, stdout)
	}

	if !s.validateCommand(cmd.Name) {
		fmt.Printf("%s: command not found\n", cmd.Name)
		return nil
	}

	switch cmd.Name {
	case "exit":
		s.handleExit(cmd.Args)
	case "echo":
		s.handleEcho(cmd, stdout)
	case "type":
		s.handleType(cmd.Args, stdout)
	case "pwd":
		s.handlePwd(stdout)
	case "cd":
		s.handleCd(cmd.Args, os.Stderr)
	case "history":
		s.handleHistory(cmd.Args, stdout)
	default:
		s.handleExternal(cmd, stdin, stdout)
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
