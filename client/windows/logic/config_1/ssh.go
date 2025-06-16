package config_1

import (
	"encoding/json"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

type SSHKeyPayload struct {
	PublicKey string `json:"public_key"`
}

func HandleSSHUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Only POST method is allowed",
		})
		return
	}

	var payload SSHKeyPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Invalid JSON format",
		})
		return
	}

	if payload.PublicKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Public key is required",
		})
		return
	}

	currentUser, err := user.Current()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Unable to determine current user",
		})
		return
	}

	sshDir := filepath.Join(currentUser.HomeDir, ".ssh")
	authKeysPath := filepath.Join(sshDir, "authorized_keys")

	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "failed",
				"message": "Failed to create .ssh directory",
			})
			return
		}
	}

	f, err := os.OpenFile(authKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Failed to open authorized_keys",
		})
		return
	}
	defer f.Close()

	if _, err := f.WriteString(payload.PublicKey + "\n"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Failed to write SSH key",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
