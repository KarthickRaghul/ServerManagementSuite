package config_1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SSHKeyRequest struct {
	Key string `json:"key"`
}

func HandleSSHUpload(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var keyRequest SSHKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&keyRequest); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate key
	if keyRequest.Key == "" {
		sendError(w, "SSH key is empty", http.StatusBadRequest)
		return
	}

	// Get Windows username
	username, err := getWindowsUsername()
	if err != nil {
		sendError(w, "Could not get username: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build paths
	homeDir := filepath.Join("C:\\Users", username)
	sshDir := filepath.Join(homeDir, ".ssh")
	authFile := filepath.Join(sshDir, "authorized_keys")

	// Create .ssh directory
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		sendError(w, "Failed to create .ssh directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Append newline to key if needed
	sshKey := keyRequest.Key
	if !strings.HasSuffix(sshKey, "\n") {
		sshKey += "\n"
	}

	// Write to authorized_keys
	file, err := os.OpenFile(authFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		sendError(w, "Failed to open authorized_keys: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(sshKey); err != nil {
		sendError(w, "Failed to write SSH key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Success response (matches Linux)
	sendPostSuccess(w)
}

// getWindowsUsername extracts username using whoami command
func getWindowsUsername() (string, error) {
	cmd := exec.Command("whoami")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run whoami: %v", err)
	}

	fullOutput := strings.TrimSpace(string(out))
	parts := strings.Split(fullOutput, `\`)
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected whoami format: %s", fullOutput)
	}
	return parts[1], nil
}
