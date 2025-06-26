package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
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
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
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

	resp := RestartResponse{
		Service:   serviceName,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// PowerShell command to restart the service
	powershellCmd := fmt.Sprintf("Restart-Service -Name '%s' -Force", serviceName)
	cmd := exec.Command("powershell", "-Command", powershellCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		resp.Status = "failed"
		resp.Message = fmt.Sprintf("Failed to restart service: %v, Output: %s", err, string(output))
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Status = "success"
	resp.Message = fmt.Sprintf("Service '%s' restarted successfully", serviceName)
	json.NewEncoder(w).Encode(resp)
}
