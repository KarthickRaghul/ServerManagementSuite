package optimization

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"time"
)

// ServiceInfo represents service details in JSON response
type ServiceInfo struct {
	Name string `json:"name"`
}

// ServiceListResponse represents the JSON structure to return restartable services
type ServiceListResponse struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Services  []ServiceInfo `json:"services"`
	Timestamp string        `json:"timestamp"`
}

// HandleRestartableServices writes JSON with only restartable services
func HandleRestartableServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cmd := exec.Command("powershell", "-Command", `
		Get-Service | Where-Object {$_.Status -eq "Running"} | 
		Select-Object Name,CanStop | ConvertTo-Json -Depth 1
	`)
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error fetching running services: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Unmarshal PowerShell output
	var services []map[string]interface{}
	if err := json.Unmarshal(output, &services); err != nil {
		// fallback in case of single object (not array)
		var singleService map[string]interface{}
		if err2 := json.Unmarshal(output, &singleService); err2 != nil {
			http.Error(w, "Failed to parse service data", http.StatusInternalServerError)
			return
		}
		services = append(services, singleService)
	}

	restartable := []ServiceInfo{}

	for _, svc := range services {
		canStop, ok := svc["CanStop"].(bool)
		if ok && canStop {
			name, _ := svc["Name"].(string)
			restartable = append(restartable, ServiceInfo{Name: name})
		}
	}

	response := ServiceListResponse{
		Status:    "success",
		Message:   "List of restartable running services",
		Services:  restartable,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	json.NewEncoder(w).Encode(response)
}
