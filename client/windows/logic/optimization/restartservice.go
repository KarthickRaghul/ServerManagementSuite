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

func RestartServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RestartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	serviceName := strings.TrimSpace(req.Service)
	if serviceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	resp := RestartResponse{
		Service:   serviceName,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Stop the service
	stopCmd := exec.Command("powershell", "-Command", "Stop-Service -Name \""+serviceName+"\" -Force")
	stopOutput, stopErr := stopCmd.CombinedOutput()
	if stopErr != nil {
		resp.Status = "error"
		resp.Message = fmt.Sprintf("Failed to stop service: %v, Output: %s", stopErr, string(stopOutput))
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Start the service
	startCmd := exec.Command("powershell", "-Command", "Start-Service -Name \""+serviceName+"\"")
	startOutput, startErr := startCmd.CombinedOutput()
	if startErr != nil {
		resp.Status = "error"
		resp.Message = fmt.Sprintf("Failed to start service: %v, Output: %s", startErr, string(startOutput))
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Status = "success"
	resp.Message = fmt.Sprintf("Service '%s' stopped and started successfully", serviceName)
	json.NewEncoder(w).Encode(resp)
}
