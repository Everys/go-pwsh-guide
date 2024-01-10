package main

import (
	"fmt"
	"os"
	"os/exec"
)

var server = "VotreServeurWinRM"
var command = "VotreCommandePowerShell"

func runPowerShellScript() error {
	scriptPath := "C:\\Chemin\\Vers\\Le\\Script.ps1"

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath, "-server", server, "-command", command)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func main() {
	// Exécuter le script PowerShell dans le même processus Go
	err := runPowerShellScript()
	if err != nil {
		fmt.Println("Erreur lors de l'exécution du script PowerShell:", err)
	}
}
