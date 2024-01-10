package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

// CommandRequest représente la structure de la requête JSON attendue dans le corps de la requête POST.
type CommandRequest struct {
	Server  string `json:"server"`
	Command string `json:"command"`
}

var sessionMap sync.Map

func executeCommand(w http.ResponseWriter, r *http.Request) {
	// Lire le corps de la requête
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erreur de lecture du corps de la requête", http.StatusBadRequest)
		return
	}

	// Désérialiser la requête JSON
	var request CommandRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Erreur de désérialisation JSON", http.StatusBadRequest)
		return
	}

	// Vérifier si une session existe déjà pour le serveur
	session, ok := sessionMap.Load(request.Server)
	if !ok {
		// Si la session n'existe pas, créer une nouvelle session WinRM
		session, err = createWinRMSession(request.Server)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erreur de création de la session WinRM: %v", err), http.StatusInternalServerError)
			return
		}
		// Stocker la session dans la carte
		sessionMap.Store(request.Server, session)
	}

	// Exécuter la commande avec la session existante
	output, err := executeWinRMCommand(session.(*exec.Cmd), request.Command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur d'exécution de la commande WinRM: %v", err), http.StatusInternalServerError)
		return
	}

	// Envoyer la réponse au client HTTP
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func createWinRMSession(server string) (*exec.Cmd, error) {
	// Construction de la commande PowerShell pour créer une nouvelle session WinRM
	psCommand := fmt.Sprintf("$session = New-PSSession -ComputerName %s", server)
	powershellCommand := exec.Command("powershell", "-Command", psCommand)

	// Exécution de la commande PowerShell
	if err := powershellCommand.Run(); err != nil {
		return nil, err
	}

	return powershellCommand, nil
}

func executeWinRMCommand(session *exec.Cmd, command string) ([]byte, error) {
	// Construction de la commande PowerShell pour exécuter la commande WinRM avec la session existante
	psCommand := fmt.Sprintf("Invoke-Command -Session $session -ScriptBlock {%s}", command)
	powershellCommand := exec.Command("powershell", "-Command", psCommand)

	// Capture de la sortie standard et d'erreur
	output, err := powershellCommand.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func main() {
	// Définir la route pour la gestion des requêtes POST /execute
	http.HandleFunc("/execute", executeCommand)

	// Démarrer le serveur HTTP sur le port 8080
	fmt.Println("Serveur en écoute sur le port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erreur lors du démarrage du serveur:", err)
		os.Exit(1)
	}
}
