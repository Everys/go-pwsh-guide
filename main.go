package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
)

var (
	mu          sync.Mutex
	powerShell  *exec.Cmd
	powerShellIn *chan string // canal pour écrire des commandes dans le processus pwsh
)

func main() {
	initializePowerShell()
	defer cleanupPowerShell()

	http.HandleFunc("/execute", executeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initializePowerShell() {
	mu.Lock()
	defer mu.Unlock()

	if powerShell == nil {
		cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-Command", "-")
		stdinPipe, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			defer stdinPipe.Close()
			for {
				select {
				case cmd := <-*powerShellIn:
					_, err := fmt.Fprintln(stdinPipe, cmd)
					if err != nil {
						log.Printf("Error writing to PowerShell stdin: %v", err)
					}
				}
			}
		}()

		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		powerShell = cmd
		powerShellIn = &chan string
	}
}

func cleanupPowerShell() {
	if powerShell != nil {
		powerShell.Wait()
	}
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	script := r.FormValue("script")

	if script == "" {
		http.Error(w, "Missing 'script' parameter", http.StatusBadRequest)
		return
	}

	*powerShellIn <- script

	// Vous pouvez également ajouter une logique pour récupérer la sortie du script PowerShell si nécessaire

	fmt.Fprint(w, "Script en cours d'exécution...")
}
