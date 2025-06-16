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

type PasswordChangeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
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
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(PasswordChangeResponse{
			Status:  "failed",
			Message: "Only POST method allowed",
		})
		return
	}

	var p PasswordChange
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(PasswordChangeResponse{
			Status:  "failed",
			Message: "Invalid input format",
		})
		return
	}

	// Validate input
	if p.Username == "" || p.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(PasswordChangeResponse{
			Status:  "failed",
			Message: "Username and new password are required",
		})
		return
	}

	// Use the secure password change function
	if err := changePassword(p.Username, p.NewPassword); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(PasswordChangeResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Password change failed: %v", err),
		})
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PasswordChangeResponse{
		Status:  "success",
		Message: fmt.Sprintf("Password changed successfully for user: %s", p.Username),
	})
}
