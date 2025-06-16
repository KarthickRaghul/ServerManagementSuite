package config_1

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
)

type PasswordChange struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"current"` // unused in current logic
	NewPassword     string `json:"new"`
	ConfirmPassword string `json:"confirm"`
}

func HandlePasswordChange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Only POST allowed",
		})
		return
	}

	var p PasswordChange
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}

	if p.NewPassword != p.ConfirmPassword {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "New passwords do not match",
		})
		return
	}

	cmd := exec.Command("powershell", "-Command",
		`Set-LocalUser -Name `+strconv.Quote(p.Username)+
			` -Password (ConvertTo-SecureString `+strconv.Quote(p.NewPassword)+
			` -AsPlainText -Force)`)

	out, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Password change failed: " + string(out),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
