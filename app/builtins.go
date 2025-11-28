package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var builtinCommands = map[string]struct{}{
	"type":    {},
	"echo":    {},
	"exit":    {},
	"pwd":     {},
	"cd":      {},
	"history": {},
}

func (s *Shell) handleExit(args []string) {
	// Save history to HISTFILE if set
	if histfile := os.Getenv("HISTFILE"); histfile != "" {
		content := strings.Join(s.history, "\n") + "\n"
		os.WriteFile(histfile, []byte(content), 0o644)
	}

	if len(args) == 0 {
		os.Exit(0)
		return
	}
	v, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("incorrect command arguments")
		return
	}
	os.Exit(v)
}

func (s *Shell) handleEcho(cmd Command, stdout io.Writer) {
	output := strings.Join(cmd.Args, " ") + "\n"
	if cmd.RedirectFile != "" && !cmd.RedirectStderr {
		s.writeToFile(cmd.RedirectFile, []byte(output), cmd.AppendMode)
	} else {
		if cmd.RedirectFile != "" {
			s.writeToFile(cmd.RedirectFile, []byte(""), cmd.AppendMode)
		}
		fmt.Fprint(stdout, output)
	}
}

func (s *Shell) handleType(args []string, stdout io.Writer) {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "no command found")
		return
	}
	commandName := args[0]
	filePath := s.isInPath(commandName)
	if _, ok := builtinCommands[commandName]; ok {
		fmt.Fprintf(stdout, "%s is a shell builtin\n", commandName)
	} else if filePath != "" {
		fmt.Fprintf(stdout, "%[1]s is %[2]s\n", commandName, filePath)
	} else {
		fmt.Fprintf(stdout, "%s: not found\n", commandName)
	}
}

func (s *Shell) handlePwd(stdout io.Writer) {
	dir, err := os.Getwd()
	if err == nil {
		fmt.Fprintf(stdout, "%s\n", dir)
	} else {
		fmt.Fprintf(stdout, "error getting pwd %s\n", err)
	}
}

func (s *Shell) handleCd(args []string, stderr io.Writer) {
	dir := os.Getenv("HOME")
	if len(args) > 0 && args[0] != "~" {
		dir = args[0]
	}
	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(stderr, "cd: %s: No such file or directory\n", dir)
	}
}

func (s *Shell) handleHistory(args []string, stdout io.Writer) {
	if len(args) > 0 && args[0] == "-r" {
		if len(args) < 2 {
			fmt.Fprintln(stdout, "history: missing argument")
			return
		}
		filePath := args[1]
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(stdout, "history: %s\n", err)
			return
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if line != "" {
				s.history = append(s.history, line)
			}
		}
		return
	}

	if len(args) > 0 && args[0] == "-w" {
		if len(args) < 2 {
			fmt.Fprintln(stdout, "history: missing argument")
			return
		}
		filePath := args[1]
		content := strings.Join(s.history, "\n") + "\n"
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			fmt.Fprintf(stdout, "history: %s\n", err)
			return
		}
		return
	}

	if len(args) > 0 && args[0] == "-a" {
		if len(args) < 2 {
			fmt.Fprintln(stdout, "history: missing argument")
			return
		}
		filePath := args[1]
		// append history to file
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(stdout, "history: %s\n", err)
			return
		}
		defer f.Close()

		newLines := s.history[s.historyAppendedCount:]
		if len(newLines) > 0 {
			content := strings.Join(newLines, "\n") + "\n"
			if _, err := f.WriteString(content); err != nil {
				fmt.Fprintf(stdout, "history: %s\n", err)
				return
			}
			s.historyAppendedCount = len(s.history)
		}
		return
	}

	var num int
	var err error
	if len(args) > 0 {
		if num, err = strconv.Atoi(args[0]); err != nil {
			fmt.Fprintln(stdout, "history: invalid number")
			return
		}
	}

	start := 0
	if num > 0 && num < len(s.history) {
		start = len(s.history) - num
	}

	for i := start; i < len(s.history); i++ {
		fmt.Fprintf(stdout, "    %d  %s\n", i+1, s.history[i])
	}
}

func (s *Shell) handleExternal(cmd Command, stdin io.Reader, stdout io.Writer) {
	execCmd := exec.Command(cmd.Name, cmd.Args...)
	execCmd.Stdin = stdin

	if cmd.RedirectFile != "" {
		if cmd.RedirectStderr {
			execCmd.Stdout = stdout
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
		execCmd.Stdout = stdout
		execCmd.Stderr = os.Stderr
		_ = execCmd.Run()
	}
}
