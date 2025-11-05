package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

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
		} else {
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
	if mainCommand == "exit" {
		v, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("incorrect command arguments: %s", command)
		}
		fmt.Printf("exit with status %d\n", v)
		os.Exit(v)
	}
	if mainCommand == "echo" {
		fmt.Println(strings.TrimSpace(command[4:]))
		return nil
	}
	return fmt.Errorf("%s: command not found", command)
}
