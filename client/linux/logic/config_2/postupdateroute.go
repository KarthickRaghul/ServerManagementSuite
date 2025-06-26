package config_2

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
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
	// Handle special cases
	if destination == "default" || destination == "0.0.0.0" {
		return "0.0.0.0/0", nil
	}

	// Check if it already contains CIDR notation
	if strings.Contains(destination, "/") {
		// Parse the CIDR to validate it
		_, network, err := net.ParseCIDR(destination)
		if err != nil {
			return "", fmt.Errorf("invalid CIDR format: %v", err)
		}
		// Return the correct network address
		return network.String(), nil
	}

	// If no CIDR notation, try to determine the appropriate one
	ip := net.ParseIP(destination)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", destination)
	}

	// Determine CIDR based on IP class
	if ip.IsLoopback() {
		return destination + "/32", nil // Host route for loopback
	}

	// For other IPs, assume /24 for private networks, /32 for host routes
	if isPrivateIP(ip) {
		// Check if it looks like a network address (ends with .0)
		if strings.HasSuffix(destination, ".0") {
			return destination + "/24", nil
		} else {
			return destination + "/32", nil // Host route
		}
	}

	// Default to /32 for host routes
	return destination + "/32", nil
}

// isPrivateIP checks if an IP is in private range
func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// HandleUpdateRoute handles the POST request to change the route table
func HandleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req ChangeRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate action
	if req.Action != "add" && req.Action != "delete" {
		sendError(w, fmt.Sprintf("Action must be 'add' or 'delete', got '%s'", req.Action), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Destination == "" || req.Gateway == "" {
		sendError(w, "Destination and gateway are required", http.StatusBadRequest)
		return
	}

	// Validate and fix network address
	correctedDestination, err := validateAndFixNetwork(req.Destination)
	if err != nil {
		sendError(w, "Invalid destination network: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate gateway IP
	if net.ParseIP(req.Gateway) == nil {
		sendError(w, fmt.Sprintf("Invalid gateway IP address: %s", req.Gateway), http.StatusBadRequest)
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
		sendError(w, fmt.Sprintf("Failed to %s route: %v, output: %s", req.Action, err, string(output)), http.StatusInternalServerError)
		return
	}

	// Log the action
	fmt.Printf("Route %s successful: Destination: %s, Gateway: %s\n", req.Action, correctedDestination, req.Gateway)

	// Send success response
	sendPostSuccess(w)
}
