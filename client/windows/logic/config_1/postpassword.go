package config_1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
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

// changePassword updates the user's password using the Windows 'net user' command.
func changePassword(username, newPassword string) error {
	// Windows command to change password
	// Must be run with Administrator privileges
	cmd := exec.Command("net", "user", username, newPassword)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %v, output: %s", err, output)
	}
	if !strings.Contains(string(output), "command completed successfully") {
		return fmt.Errorf("unexpected output: %s", output)
	}
	return nil
}

func HandlePasswordChange(w http.ResponseWriter, r *http.Request) {
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

	if p.Username == "" || p.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(PasswordChangeResponse{
			Status:  "failed",
			Message: "Username and new password are required",
		})
		return
	}

	if err := changePassword(p.Username, p.NewPassword); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(PasswordChangeResponse{
			Status:  "failed",
			Message: fmt.Sprintf("Password change failed: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(PasswordChangeResponse{
		Status:  "success",
		Message: fmt.Sprintf("Password changed successfully for user: %s", p.Username),
	})
}
