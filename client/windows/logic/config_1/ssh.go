package config_1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type SSHKeyRequest struct {
	Key string `json:"key"`
}

type SSHKeyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// getWindowsUsername runs `whoami` and extracts the username
func getWindowsUsername() (string, error) {
	cmd := exec.Command("whoami")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run whoami: %v", err)
	}
	fullOutput := strings.TrimSpace(out.String()) // e.g., "DESKTOP-XYZ\\karthick"
	parts := strings.Split(fullOutput, `\`)
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected whoami format: %s", fullOutput)
	}
	return parts[1], nil // Return just "karthick"
}

func HandleSSHUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Only POST method allowed",
		})
		return
	}

	var keyRequest SSHKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&keyRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Invalid JSON input",
		})
		return
	}

	if keyRequest.Key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "SSH key is empty",
		})
		return
	}

	if runtime.GOOS != "windows" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "This endpoint is designed for Windows systems",
		})
		return
	}

	username, err := getWindowsUsername()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Could not get username: %v", err),
		})
		return
	}

	homeDir := filepath.Join("C:\\Users", username)
	sshDir := filepath.Join(homeDir, ".ssh")
	authFile := filepath.Join(sshDir, "authorized_keys")

	if err := os.MkdirAll(sshDir, 0700); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Failed to create .ssh directory",
		})
		return
	}

	sshKey := keyRequest.Key
	if sshKey[len(sshKey)-1] != '\n' {
		sshKey += "\n"
	}

	file, err := os.OpenFile(authFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to open authorized_keys: %v", err),
		})
		return
	}
	defer file.Close()

	if _, err := file.WriteString(sshKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Failed to write SSH key",
		})
		return
	}

	json.NewEncoder(w).Encode(SSHKeyResponse{
		Status:  "success",
		Message: fmt.Sprintf("SSH key uploaded to %s", authFile),
	})
}
