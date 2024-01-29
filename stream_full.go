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
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "toto.txt"))

		// cmd := "date"
		cmd := "Get-Content -Path /Users/U42SF98/Downloads/bigfile.txt -Raw"
		outBoundary := createBoundary()
		errBoundary := createBoundary()
		full := fmt.Sprintf("%s; echo '%s'; [Console]::Error.WriteLine('%s')%s", cmd, outBoundary, errBoundary, newline)
		fmt.Println(full)
		_, err = stdinpipe.Write([]byte(full))
		if err != nil {
			log.Fatal(err)
		}

		bufsize := 1024
		markerOut := outBoundary + newline
		markerErr := errBoundary + newline

		// waiter := &sync.WaitGroup{}
		// waiter.Add(2)

		// // go...
		// // go...

		// // -> , signal *sync.WaitGroup
		// waiter.Wait()
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for {
				buf := make([]byte, bufsize)
				nr, err := stdoutpipe.Read(buf)
				if err != nil {
					log.Fatal(err)
				}

				// strings.TrimSuffix(output, marker)

				st := removeAnsiChar(string(buf[:nr]))
				io.Copy(c.Writer, strings.NewReader(strings.TrimSuffix(st, markerOut)))
				if strings.HasSuffix(st, markerOut) {
					break
				}
			}
		}()

		go func() {

			stderroutput := ""
			for {
				buff := make([]byte, bufsize)
				nr, err := stderrpipe.Read(buff)
				if err != nil {
					log.Fatal(err)
				}

				st := removeAnsiChar(string(buff[:nr]))

				if strings.HasSuffix(st, markerErr) {
					break
				} else {
					stderroutput += st
				}

			}

			if strings.EqualFold(stderroutput, "") {
				fmt.Println("no error in stderr")
				wg.Done()
			} else {
				c.Header("Content-Disposition", "")
				c.JSON(http.StatusInternalServerError, gin.H{"error": stderroutput})
				wg.Done()
			}
		}()

		wg.Wait()

	})

	router.Run(":8080")
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
