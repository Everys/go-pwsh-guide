package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const newline = "\n"
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func removeAnsiChar(str string) string {
	return re.ReplaceAllString(str, "")
}
func main() {
	router := gin.Default()

	command := exec.Command("pwsh", "-NoExit", "-Command", "-")

	stdinpipe, err := command.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdoutpipe, err := command.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderrpipe, err := command.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = command.Start()
	if err != nil {
		log.Fatal(err)
	}

	router.GET("/download", func(c *gin.Context) {
		// c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "toto.txt"))

		// cmd := "date"
		cmd := "Get-Contaent -Path /Users/U42SF98/Downloads/bigfile.txt -Raw"
		outBoundary := createBoundary()
		errBoundary := createBoundary()
		full := fmt.Sprintf("%s; echo '%s'; [Console]::Error.WriteLine('%s')%s", cmd, outBoundary, errBoundary, newline)
		fmt.Println(full)
		_, err = stdinpipe.Write([]byte(full))
		if err != nil {
			log.Fatal(err)
		}
		waiter := &sync.WaitGroup{}
		waiter.Add(2)

		go streamReaderStdout(stdoutpipe, c.Writer, outBoundary, waiter, 1024)

		stderr := ""
		go streamReaderStderr(stderrpipe, &stderr, errBoundary, waiter, 1024)

		if strings.EqualFold(stderr, "") {
			fmt.Println("no error in stderr")
		} else {
			c.Header("Content-Disposition", "")
			c.JSON(http.StatusInternalServerError, gin.H{"error": stderr})
		}

		waiter.Wait()

	})

	router.Run(":8080")
}

func streamReaderStdout(stream io.Reader, output io.Writer, boundary string, signal *sync.WaitGroup, bufsize int) error {
	defer signal.Done()
	marker := boundary + newline
	for {
		buf := make([]byte, bufsize)
		nr, err := stream.Read(buf)
		if err != nil {
			return err
		}

		st := removeAnsiChar(string(buf[:nr]))
		io.Copy(output, strings.NewReader(strings.TrimSuffix(st, marker)))
		if strings.HasSuffix(st, marker) {
			break
		}
	}
	return nil
}

func streamReaderStderr(stream io.Reader, output *string, boundary string, signal *sync.WaitGroup, bufsize int) error {
	defer signal.Done()
	marker := boundary + newline

	stderroutput := ""
	for {
		buff := make([]byte, bufsize)
		nr, err := stream.Read(buff)
		if err != nil {
			log.Fatal(err)
		}

		st := removeAnsiChar(string(buff[:nr]))

		if strings.HasSuffix(st, marker) {
			break
		} else {
			stderroutput += st
		}

	}
	*output = stderroutput

	return nil
}

func createBoundary() string {
	return "$asm" + createRandomString(12) + "$"
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
