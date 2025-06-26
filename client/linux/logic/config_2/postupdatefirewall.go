package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

// UpdateFirewallRequest represents the request format for updating firewall rules
type UpdateFirewallRequest struct {
	Action      string `json:"action"`      // "add" or "delete"
	Rule        string `json:"rule"`        // "accept", "drop", "reject"
	Protocol    string `json:"protocol"`    // "tcp", "udp"
	Port        string `json:"port"`        // e.g. "80", "443"
	Source      string `json:"source"`      // optional, e.g. "192.168.1.0/24"
	Destination string `json:"destination"` // optional, e.g. "10.0.0.5"
}

// HandleUpdateFirewall handles the POST request to update firewall rules
func HandleUpdateFirewall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req UpdateFirewallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Action != "add" && req.Action != "delete" {
		sendError(w, "Action must be 'add' or 'delete'",http.StatusInternalServerError)
		return
	}
	if req.Rule == "" || req.Protocol == "" || req.Port == "" {
		sendError(w, "Rule, protocol, and port are required", http.StatusInternalServerError)
		return
	}
	if req.Rule != "accept" && req.Rule != "drop" && req.Rule != "reject" {
		sendError(w, "Rule must be 'accept', 'drop', or 'reject'", http.StatusInternalServerError)
		return
	}

	// Build iptables command
	var cmdArgs []string
	if req.Action == "add" {
		cmdArgs = append(cmdArgs, "-A", "INPUT")
	} else {
		cmdArgs = append(cmdArgs, "-D", "INPUT")
	}
	cmdArgs = append(cmdArgs, "-p", req.Protocol, "--dport", req.Port)
	if req.Source != "" {
		cmdArgs = append(cmdArgs, "-s", req.Source)
	}
	if req.Destination != "" {
		cmdArgs = append(cmdArgs, "-d", req.Destination)
	}
	switch req.Rule {
	case "accept":
		cmdArgs = append(cmdArgs, "-j", "ACCEPT")
	case "drop":
		cmdArgs = append(cmdArgs, "-j", "DROP")
	case "reject":
		cmdArgs = append(cmdArgs, "-j", "REJECT")
	}

	cmd := exec.Command("iptables", cmdArgs...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to update firewall rule", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":    "success",
		"operation": fmt.Sprintf("firewall rule %s", req.Action),
		"details":   fmt.Sprintf("Protocol: %s, Port: %s, Rule: %s, Source: %s, Destination: %s", req.Protocol, req.Port, req.Rule, req.Source, req.Destination),
		"timestamp": "2025-06-02 03:10:05",
		"user":      "kishore-001",
	}
	json.NewEncoder(w).Encode(response)
}

