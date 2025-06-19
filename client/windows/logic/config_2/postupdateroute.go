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

type ChangeRouteRequest struct {
	Action      string `json:"action"`      // "add" or "delete"
	Destination string `json:"destination"` // e.g. "192.168.2.0/24"
	Gateway     string `json:"gateway"`     // e.g. "192.168.1.1"
	Interface   string `json:"interface"`   // Not used on Windows
	Metric      string `json:"metric"`      // e.g. "100"
}

type RouteResponse struct {
	Status    string `json:"status"`
	Operation string `json:"operation,omitempty"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	User      string `json:"user,omitempty"`
	Error     string `json:"error,omitempty"`
}

func HandleUpdateRoute(w http.ResponseWriter, r *http.Request) {
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

	if req.Action != "add" && req.Action != "delete" {
		sendRouteErrorResponse(w, "Action must be 'add' or 'delete'", fmt.Errorf("invalid action: %s", req.Action))
		return
	}

	if req.Destination == "" || req.Gateway == "" {
		sendRouteErrorResponse(w, "Destination and gateway are required", fmt.Errorf("missing fields"))
		return
	}

	correctedDestination, err := validateAndFixNetwork(req.Destination)
	if err != nil {
		sendRouteErrorResponse(w, "Invalid destination network", err)
		return
	}

	if net.ParseIP(req.Gateway) == nil {
		sendRouteErrorResponse(w, "Invalid gateway IP address", fmt.Errorf("gateway %s is not valid", req.Gateway))
		return
	}

	// Split destination into IP and mask for Windows
	destIP, ipNet, err := net.ParseCIDR(correctedDestination)
	if err != nil {
		sendRouteErrorResponse(w, "Invalid destination format", err)
		return
	}
	mask := net.IP(ipNet.Mask).String()

	var cmdArgs []string
	if req.Action == "add" {
		cmdArgs = []string{"route", "ADD", destIP.String(), "MASK", mask, req.Gateway}
		if req.Metric != "" {
			cmdArgs = append(cmdArgs, "METRIC", req.Metric)
		}
	} else {
		cmdArgs = []string{"route", "DELETE", destIP.String()}
	}

	cmd := exec.Command("cmd", "/C", strings.Join(cmdArgs, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendRouteErrorResponse(w, "Failed to change route", fmt.Errorf("%v, output: %s", err, string(output)))
		return
	}

	response := RouteResponse{
		Status:    "success",
		Operation: fmt.Sprintf("route %s", req.Action),
		Details:   fmt.Sprintf("Destination: %s, Gateway: %s, Metric: %s", correctedDestination, req.Gateway, req.Metric),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		User:      getCurrentUsername1(),
	}
	json.NewEncoder(w).Encode(response)
}

func validateAndFixNetwork(destination string) (string, error) {
	if !strings.Contains(destination, "/") {
		return "", fmt.Errorf("destination must include CIDR notation")
	}
	_, network, err := net.ParseCIDR(destination)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR format: %v", err)
	}
	return network.String(), nil
}

func getCurrentUsername1() string {
	out, err := exec.Command("whoami").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func sendRouteErrorResponse(w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(RouteResponse{
		Status: "failed",
		Error:  fmt.Sprintf("%s: %v", message, err),
	})
}
