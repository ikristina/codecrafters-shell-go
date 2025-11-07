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

var builtins = map[string]struct{}{
	"type": {},
	"echo": {},
	"exit": {},
	"pwd": {},
}

func main() {
	buffer := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, "$ ")
		// capture the user's command in the "command" variable
		command, err := buffer.ReadString('\n')
		if err != nil {
			fmt.Println("error capturing the command.")
			return
		}
		command = command[:len(command)-1] // the last character is "\n" which we delete here.

		if err = parseCommand(command); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func parseCommand(command string) error {
	command = strings.TrimSpace(command)
	mainCommand := ""
	var args []string
	if len(command) != 0 {
		args = strings.Split(command, " ")
		if len(args) >= 1 {
			mainCommand = strings.TrimSpace(args[0])
			args = args[1:]
		}
	}
	if !validateCmd(mainCommand) {
		fmt.Printf("%s: command not found\n", command)
		return nil
	}
	switch mainCommand {
	case "exit":
		exit(command, args)
	case "echo":
		fmt.Println(strings.TrimSpace(command[4:]))
		return nil
	case "type":
		typeFunc(args)
		return nil
	case "pwd":
		pwd()
		return nil
	default:
		output, err := exec.Command(mainCommand, args...).Output()
		if err == nil {
			fmt.Print(string(output))
			return nil
		}
	}
	return fmt.Errorf("%s: command not found", command)
}

func validateCmd(s string) bool { // true if the command exists, false otherwise
	if _, ok := builtins[s]; !ok {
		if isInThePath(s) == "" {
			return false
		}
	}
	return true
}

func isInThePath(s string) string { // return path of the command
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		file := filepath.Join(path, s)
		info, err := os.Stat(file)
		if err == nil && info.Mode()&0o111 != 0 {
			return file
		}
	}
	return ""
}

func exit(command string, args []string) {
	if len(args) == 0 {
		os.Exit(0)
		return
	}
	v, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("incorrect command arguments: %s", command)
		return
	}
	os.Exit(v)
}

func typeFunc(args []string) {
	if len(args) == 0 {
		fmt.Println("no command found")
		return
	}
	v := args[0]
	file := isInThePath(v)
	if _, ok := builtins[v]; ok {
		fmt.Printf("%s is a shell builtin\n", v)
	} else if file != "" {
		fmt.Printf("%[1]s is %[2]s\n", v, file)
	} else {
		fmt.Printf("%s: not found\n", v)
	}
}

func pwd() {
	dir, err := os.Getwd()
	if err == nil {
		fmt.Printf("%s\n", dir)
	} else {
		fmt.Printf("error getting pwd %s\n", err)
	}
}
