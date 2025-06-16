package config_1

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// SSHKeyRequest defines the expected JSON structure for SSH key uploads
type SSHKeyRequest struct {
	Key string `json:"key"`
}

// SSHKeyResponse defines the JSON response structure
type SSHKeyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func HandleSSHUpload(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Only POST method allowed",
		})
		return
	}

	// Parse the JSON request
	var keyRequest SSHKeyRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&keyRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Failed to parse JSON request",
		})
		return
	}

	// Validate that the key was provided
	if keyRequest.Key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "SSH key is empty",
		})
		return
	}

	// Get the user's home directory and construct the .ssh path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Failed to determine home directory",
		})
		return
	}

	// Create the .ssh directory if it doesn't exist
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Failed to create .ssh directory",
		})
		return
	}

	// Path to the authorized_keys file
	sshPath := filepath.Join(sshDir, "authorized_keys")

	// Append a newline if the key doesn't end with one
	sshKey := keyRequest.Key
	if sshKey[len(sshKey)-1] != '\n' {
		sshKey += "\n"
	}

	// Write the SSH key to the file
	// Using os.O_APPEND to add to existing keys rather than overwrite
	file, err := os.OpenFile(sshPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SSHKeyResponse{
			Status:  "failed",
			Message: "Failed to open authorized_keys file",
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

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SSHKeyResponse{
		Status:  "success",
		Message: "SSH key uploaded successfully",
	})
}
