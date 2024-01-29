package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"

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

	// stderrpipe, err := command.StderrPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	err = command.Start()
	if err != nil {
		log.Fatal(err)
	}

	router.GET("/download", func(c *gin.Context) {
		// c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "toto.txt"))

		// cmd := "date"
		cmd := "Get-Content -Path /Users/U42SF98/Downloads/bigfile.txt -Raw"
		outBoundary := "--asm--"
		full := fmt.Sprintf("%s; echo '%s';%s", cmd, outBoundary, newline)

		_, err = stdinpipe.Write([]byte(full))
		if err != nil {
			log.Fatal(err)
		}

		bufsize := 1024
		marker := outBoundary + newline

		for {
			buf := make([]byte, bufsize)
			nr, err := stdoutpipe.Read(buf)
			if err != nil {
				log.Fatal(err)
			}

			st := removeAnsiChar(string(buf[:nr]))

			if strings.HasSuffix(st, marker) {
				break
			} else {
				io.Copy(c.Writer, strings.NewReader(st))
			}
		}
	})

	router.Run(":8080")
}
