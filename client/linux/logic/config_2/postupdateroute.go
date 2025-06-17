package config_2

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// ChangeRouteRequest represents the request format for changing the route table
type ChangeRouteRequest struct {
	Action      string `json:"action"`      // "add" or "delete"
	Destination string `json:"destination"` // e.g. "192.168.2.0/24"
	Gateway     string `json:"gateway"`     // e.g. "192.168.1.1"
	Interface   string `json:"interface"`   // e.g. "enp3s0", optional
	Metric      string `json:"metric"`      // e.g. "100", optional
}

// RouteResponse represents the response format
type RouteResponse struct {
	Status    string `json:"status"`
	Operation string `json:"operation,omitempty"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	User      string `json:"user,omitempty"`
	Error     string `json:"error,omitempty"`
}

// validateAndFixNetwork validates the network address and fixes it if needed
func validateAndFixNetwork(destination string) (string, error) {
	// Check if it contains CIDR notation
	if !strings.Contains(destination, "/") {
		return "", fmt.Errorf("destination must include CIDR notation (e.g., /24)")
	}

	// Parse the CIDR
	_, network, err := net.ParseCIDR(destination)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR format: %v", err)
	}

	// Return the correct network address
	return network.String(), nil
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
		sendRouteErrorResponse(w, "Failed to parse request body", err)
		return
	}

	// Validate action
	if req.Action != "add" && req.Action != "delete" {
		sendRouteErrorResponse(w, "Action must be 'add' or 'delete'", fmt.Errorf("invalid action: %s", req.Action))
		return
	}

	// Validate required fields
	if req.Destination == "" || req.Gateway == "" {
		sendRouteErrorResponse(w, "Destination and gateway are required", fmt.Errorf("missing fields"))
		return
	}

	// Validate and fix network address
	correctedDestination, err := validateAndFixNetwork(req.Destination)
	if err != nil {
		sendRouteErrorResponse(w, "Invalid destination network", err)
		return
	}

	// Validate gateway IP
	if net.ParseIP(req.Gateway) == nil {
		sendRouteErrorResponse(w, "Invalid gateway IP address", fmt.Errorf("gateway %s is not a valid IP", req.Gateway))
		return
	}

	// Build ip route command with corrected destination
	var cmdArgs []string
	if req.Action == "add" {
		cmdArgs = []string{"route", "add", correctedDestination, "via", req.Gateway}
	} else {
		cmdArgs = []string{"route", "del", correctedDestination, "via", req.Gateway}
	}

	if req.Interface != "" {
		cmdArgs = append(cmdArgs, "dev", req.Interface)
	}
	if req.Metric != "" {
		cmdArgs = append(cmdArgs, "metric", req.Metric)
	}

	// Execute the command
	cmd := exec.Command("ip", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendRouteErrorResponse(w, "Failed to change route", fmt.Errorf("%v, output: %s", err, string(output)))
		return
	}

	// Success response
	response := RouteResponse{
		Status:    "success",
		Operation: fmt.Sprintf("route %s", req.Action),
		Details:   fmt.Sprintf("Destination: %s, Gateway: %s, Interface: %s, Metric: %s", correctedDestination, req.Gateway, req.Interface, req.Metric),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		User:      "system",
	}

	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse sends an error response in JSON format
func sendRouteErrorResponse(w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	response := RouteResponse{
		Status: "failed",
		Error:  fmt.Sprintf("%s: %v", message, err),
	}
	json.NewEncoder(w).Encode(response)
}
