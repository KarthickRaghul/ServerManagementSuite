package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type RestartRequest struct {
	Service string `json:"service"`
}

type RestartResponse struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func HandleRestartService(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	var req RestartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	serviceName := strings.TrimSpace(req.Service)
	if serviceName == "" {
		sendError(w, "Service name is required", http.StatusBadRequest)
		return
	}

	// Use systemctl to restart the service
	cmd := exec.Command("systemctl", "restart", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to restart service '%s': %v, Output: %s", serviceName, err, string(output)), http.StatusInternalServerError)
		return
	}

	// Send success response
	sendPostSuccess(w)
}
