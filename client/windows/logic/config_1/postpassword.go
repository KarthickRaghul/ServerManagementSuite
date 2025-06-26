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

// changePassword updates the user's password using the Windows 'net user' command.
// Must be run with Administrator privileges.
func changePassword(username, newPassword string) error {
	cmd := exec.Command("net", "user", username, newPassword)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %v, output: %s", err, output)
	}
	if !strings.Contains(strings.ToLower(string(output)), "command completed successfully") {
		return fmt.Errorf("unexpected output: %s", output)
	}
	return nil
}

func HandlePasswordChange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var p PasswordChange
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if p.Username == "" || p.NewPassword == "" {
		sendError(w, "Username and new password are required", http.StatusBadRequest)
		return
	}

	if err := changePassword(p.Username, p.NewPassword); err != nil {
		sendError(w, fmt.Sprintf("Password change failed: %v", err), http.StatusInternalServerError)
		return
	}

	sendPostSuccess(w)
}
