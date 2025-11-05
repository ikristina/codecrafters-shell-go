package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var PATH = os.Getenv("PATH")

var builtins = map[string]struct{}{
	"type": {},
	"echo": {},
	"exit": {},
}

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")
		// capture the user's command in the "command" variable
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("error capturing the command.")
			return
		}
		command = command[:len(command)-1]
		if err = parseCommand(command); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func parseCommand(command string) error {
	mainCommand := ""
	var args []string
	if len(command) != 0 {
		args = strings.Split(command, " ")
		if len(args) >= 2 {
			mainCommand = args[0]
			args = args[1:]
		}
	}
	switch mainCommand {
	case "exit":
		v, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("incorrect command arguments: %s", command)
		}
		os.Exit(v)
	case "echo":
		fmt.Println(strings.TrimSpace(command[4:]))
		return nil
	case "type":
		v := args[0]
		path := isInThePath(v)
		// fmt.Println("DEBUG: v", v)
		if _, ok := builtins[v]; ok {
			fmt.Printf("%s is a shell builtin\n", v)
			return nil
		} else if path != "" {
			fmt.Printf("%[1]s is %[2]s/%[1]s\n", v, path)
			return nil
		} else {
			fmt.Printf("%s: not found\n", v)
			return nil
		}
	}
	return fmt.Errorf("%s: command not found", command)
}

func isInThePath(s string) string { // return path of the command
	paths := filepath.SplitList(PATH)
	for _, path := range paths {
		// fmt.Println("DEBUG", path)
		_, err := os.Stat(path + "/" + s)
		// DEBUG
		if path == "/tmp/quz" {
			fmt.Println(os.ReadDir(path))
		}
		// DEBUG OFF
		if err == nil {
			return path
		}
		if os.IsNotExist(err) {
			continue
		}
	}
	return ""
}
