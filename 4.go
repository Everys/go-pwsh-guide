package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	command := exec.Command("pwsh", "-NoExit", "-Command", "-")

	stdin, err := command.StdinPipe()
	if err != nil {
		fmt.Println("Could not get hold of the PowerShell's stdin stream")
		return
	}

	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Println("Could not get hold of the PowerShell's stdout stream")
		return
	}

	command.Stderr = os.Stderr

	err = command.Start()
	if err != nil {
		fmt.Println("Could not spawn PowerShell process")
		return
	}

	stdin.Write([]byte("date\n"))

	buf := make([]byte, 1024)

	for {
		nr, err := stdout.Read(buf)
		if err != nil {
			fmt.Println("ERROR READ")
		}

		st := string(buf[:nr])
		fmt.Println("st : " + st)
	}
}
