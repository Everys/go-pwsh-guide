package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
)

const newline = "\r\n"

type Shell interface {
	Execute(cmd string) (string, string, error)
}

type shell struct {
	stdin  io.Writer
	stdout io.Reader
	stderr io.Reader
}

func main() {
	shell, err := New()
	if err != nil {
		panic(err)
	}
	stdout, stderr, err := shell.Execute("New-PSSession -ComputerName windows11 -Credential toto")
	if err != nil {
		fmt.Println(stderr)
		panic(err)
	}
	fmt.Println(stdout)

	stdout2, stderr, err := shell.Execute("Get-PSSession")
	if err != nil {
		fmt.Println(stderr)
		panic(err)
	}

	fmt.Println(stdout2)

}

func New() (Shell, error) {

	command := exec.Command("powershell.exe", "-NoExit", "-Command", "-")

	stdin, err := command.StdinPipe()
	if err != nil {
		fmt.Println("Could not get hold of the PowerShell's stdin stream")
		return nil, err
	}

	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Println("Could not get hold of the PowerShell's stdout stream")
		return nil, err
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		fmt.Println("Could not get hold of the PowerShell's stderr stream")
		return nil, err
	}

	err = command.Start()
	if err != nil {
		fmt.Println("Could not spawn PowerShell process")
		return nil, err
	}

	return &shell{stdin, stdout, stderr}, nil
}

func (s *shell) Execute(cmd string) (string, string, error) {
	outBoundary := createBoundary()
	errBoundary := createBoundary()

	// wrap the command in special markers so we know when to stop reading from the pipes
	full := fmt.Sprintf("%s; echo '%s'; [Console]::Error.WriteLine('%s')%s", cmd, outBoundary, errBoundary, newline)

	_, err := s.stdin.Write([]byte(full))
	if err != nil {
		fmt.Println("Could not send PowerShell command")
		return "", "", err
	}

	// read stdout and stderr
	sout := ""
	serr := ""

	waiter := &sync.WaitGroup{}
	waiter.Add(2)

	go streamReader(s.stdout, outBoundary, &sout, waiter)
	go streamReader(s.stderr, errBoundary, &serr, waiter)

	waiter.Wait()

	if len(serr) > 0 {
		fmt.Println("err")
		return "", "", err
	}

	return sout, serr, nil
}

func streamReader(stream io.Reader, boundary string, buffer *string, signal *sync.WaitGroup) error {
	// read all output until we have found our boundary token
	output := ""
	bufsize := 64
	marker := boundary + newline

	for {
		buf := make([]byte, bufsize)
		read, err := stream.Read(buf)
		if err != nil {
			return err
		}

		output = output + string(buf[:read])

		if strings.HasSuffix(output, marker) {
			break
		}
	}

	*buffer = strings.TrimSuffix(output, marker)
	signal.Done()

	return nil
}

func createBoundary() string {
	return "$gorilla" + createRandomString(12) + "$"
}
func createRandomString(bytes int) string {
	c := bytes
	b := make([]byte, c)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)
}
