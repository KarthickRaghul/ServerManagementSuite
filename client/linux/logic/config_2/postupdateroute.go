package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

// ChangeRouteRequest represents the request format for changing the route table
type ChangeRouteRequest struct {
	Action      string `json:"action"`      // "add" or "delete"
	Destination string `json:"destination"` // e.g. "192.168.2.0/24"
	Gateway     string `json:"gateway"`     // e.g. "192.168.1.1"
	Interface   string `json:"interface"`   // e.g. "enp3s0", optional
	Metric      string `json:"metric"`      // e.g. "100", optional
}

// HandleUpdateRoute handles the POST request to change the route table
func HandleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req ChangeRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Failed to parse request body", err)
		return
	}

	// Validate action
	if req.Action != "add" && req.Action != "delete" {
		sendError(w, "Action must be 'add' or 'delete'", fmt.Errorf("invalid action: %s", req.Action))
		return
	}
	// Validate required fields
	if req.Destination == "" || req.Gateway == "" {
		sendError(w, "Destination and gateway are required", fmt.Errorf("missing fields"))
		return
	}

	// Build ip route command
	var cmdArgs []string
	if req.Action == "add" {
		cmdArgs = []string{"route", "add", req.Destination, "via", req.Gateway}
	} else {
		cmdArgs = []string{"route", "del", req.Destination, "via", req.Gateway}
	}
	if req.Interface != "" {
		cmdArgs = append(cmdArgs, "dev", req.Interface)
	}
	if req.Metric != "" {
		cmdArgs = append(cmdArgs, "metric", req.Metric)
	}

	cmd := exec.Command("ip", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to change route", fmt.Errorf("%v, output: %s", err, string(output)))
		return
	}

	response := map[string]string{
		"status":    "success",
		"operation": fmt.Sprintf("route %s", req.Action),
		"details":   fmt.Sprintf("Destination: %s, Gateway: %s, Interface: %s, Metric: %s", req.Destination, req.Gateway, req.Interface, req.Metric),
		"timestamp": "2025-06-02 02:58:38", // Use provided timestamp
		"user":      "kishore-001",         // Use provided user login
	}
	json.NewEncoder(w).Encode(response)
}
