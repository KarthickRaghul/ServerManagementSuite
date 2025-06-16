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

	// Use systemctl to restart the service
	cmd := exec.Command("systemctl", "restart", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		resp.Status = "error"
		resp.Message = fmt.Sprintf("Failed to restart service: %v, Output: %s", err, string(output))
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Status = "success"
	resp.Message = fmt.Sprintf("Service '%s' restarted successfully", serviceName)
	json.NewEncoder(w).Encode(resp)
}

