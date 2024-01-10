package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var server = "VotreServeurWinRM"
var command = "VotreCommandePowerShell"

func runPowerShellScript() error {
	scriptPath := "C:\\Chemin\\Vers\\Le\\Script.ps1"

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath, "-server", server, "-command", command)

	// Capturer la sortie standard
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Démarrer le processus
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Lire la sortie standard
	output, err := io.ReadAll(stdout)
	if err != nil {
		return err
	}

	// Attendre la fin de l'exécution du processus
	err = cmd.Wait()
	if err != nil {
		return err
	}

	// Afficher le résultat
	fmt.Println("Résultat de la commande PowerShell :")
	fmt.Println(string(output))

	return nil
}

func main() {
	// Exécuter le script PowerShell dans le même processus Go
	err := runPowerShellScript()
	if err != nil {
		fmt.Println("Erreur lors de l'exécution du script PowerShell:", err)
	}
}
