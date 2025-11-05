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
		mainCommand := ""
		var args []string
		if len(command) != 0 {
			args = strings.Split(command, " ")
			mainCommand = args[0]
			args = args[1:]
		}
		if mainCommand == "exit" {
			v, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("incorrect command arguments: ", command)
			}
			os.Exit(v)
		}
		fmt.Println(command + ": command not found")
	}
}
