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
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CmdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Command == "" {
		sendError(w, "Command cannot be empty", http.StatusBadRequest)
		return
	}

	// Use 'cmd /C' for Windows command execution
	out, err := exec.Command("cmd", "/C", req.Command).CombinedOutput()

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		sendError(w, "Command execution failed: "+string(out), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status": "success",
		"output": string(out),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
