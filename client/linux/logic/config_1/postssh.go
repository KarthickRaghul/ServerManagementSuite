package config_1

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SSHKeyRequest defines the expected JSON structure for SSH key uploads
type SSHKeyRequest struct {
	Key string `json:"key"`
}

func HandleSSHUpload(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse the JSON request
	var keyRequest SSHKeyRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&keyRequest); err != nil {
		sendError(w, "Failed to parse JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate that the key was provided
	if keyRequest.Key == "" {
		sendError(w, "SSH key is empty", http.StatusBadRequest)
		return
	}

	// Basic validation of SSH key format
	if !isValidSSHKey(keyRequest.Key) {
		sendError(w, "Invalid SSH key format", http.StatusBadRequest)
		return
	}

	// Get the user's home directory and construct the .ssh path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		sendError(w, "Failed to determine home directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the .ssh directory if it doesn't exist
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		sendError(w, "Failed to create .ssh directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Path to the authorized_keys file
	sshPath := filepath.Join(sshDir, "authorized_keys")

	// Append a newline if the key doesn't end with one
	sshKey := keyRequest.Key
	if !strings.HasSuffix(sshKey, "\n") {
		sshKey += "\n"
	}

	// Write the SSH key to the file
	// Using os.O_APPEND to add to existing keys rather than overwrite
	file, err := os.OpenFile(sshPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		sendError(w, "Failed to open authorized_keys file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(sshKey); err != nil {
		sendError(w, "Failed to write SSH key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send successful POST response
	sendPostSuccess(w)
}

// isValidSSHKey performs basic validation of SSH key format
func isValidSSHKey(key string) bool {
	key = strings.TrimSpace(key)

	// Check if key starts with common SSH key types
	validPrefixes := []string{
		"ssh-rsa",
		"ssh-dss",
		"ssh-ed25519",
		"ecdsa-sha2-nistp256",
		"ecdsa-sha2-nistp384",
		"ecdsa-sha2-nistp521",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(key, prefix) {
			// Basic check: should have at least 3 parts (type, key, optional comment)
			parts := strings.Fields(key)
			return len(parts) >= 2
		}
	}

	return false
}
