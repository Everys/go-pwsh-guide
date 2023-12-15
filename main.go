package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
)

var (
	mu         sync.Mutex
	powerShell *exec.Cmd
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

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		cmd.Stderr = nil // Vous pouvez rediriger la sortie d'erreur si nécessaire

		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		powerShell = cmd

		// Fermer le stdinPipe et stdoutPipe lorsque le programme se termine
		go func() {
			_ = cmd.Wait()
			stdinPipe.Close()
			stdoutPipe.Close()
		}()
	}
}

func cleanupPowerShell() {
	mu.Lock()
	defer mu.Unlock()

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

	mu.Lock()
	defer mu.Unlock()

	if powerShell == nil || powerShell.ProcessState != nil && powerShell.ProcessState.Exited() {
		http.Error(w, "PowerShell session closed", http.StatusInternalServerError)
		return
	}

	stdin, err := powerShell.StdinPipe()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error obtaining PowerShell stdin: %v", err), http.StatusInternalServerError)
		return
	}
	defer stdin.Close()

	stdout, err := powerShell.StdoutPipe()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error obtaining PowerShell stdout: %v", err), http.StatusInternalServerError)
		return
	}
	defer stdout.Close()

	_, err = fmt.Fprint(stdin, script)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error writing to PowerShell stdin: %v", err), http.StatusInternalServerError)
		return
	}

	output, err := io.ReadAll(stdout)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading from PowerShell stdout: %v", err), http.StatusInternalServerError)
		return
	}

	// Envoyer le résultat du script dans la réponse HTTP
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}
