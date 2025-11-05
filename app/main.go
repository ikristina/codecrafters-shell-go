package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "$ ")
	// capture the user's command in the "command" variable
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println("error capturing the command.")
		return
	}
	fmt.Println(command[:len(command)-1] + ": command not found")
}
