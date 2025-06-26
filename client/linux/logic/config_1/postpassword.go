package config_1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"os/user"
	"strings"
)

type PasswordChange struct {
	Username    string `json:"username"`
	NewPassword string `json:"new"`
}

// Secure password change function using chpasswd
func changePassword(username, newPassword string) error {
	// Check if running as root
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	if currentUser.Uid != "0" {
		return fmt.Errorf("API must be running as root/sudo to change passwords")
	}

	// Use chpasswd to change password securely
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, newPassword))
	return cmd.Run()
}

func HandlePasswordChange(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var p PasswordChange
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate input
	if p.Username == "" || p.NewPassword == "" {
		sendError(w, "Username and new password are required", http.StatusBadRequest)
		return
	}

	// Additional validation for username (basic security check)
	if strings.Contains(p.Username, ":") || strings.Contains(p.Username, "\n") {
		sendError(w, "Invalid username format", http.StatusBadRequest)
		return
	}

	// Use the secure password change function
	if err := changePassword(p.Username, p.NewPassword); err != nil {
		sendError(w, "Password change failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send successful POST response
	sendPostSuccess(w)
}
