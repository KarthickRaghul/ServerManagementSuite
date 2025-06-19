package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type UpdateFirewallRequest struct {
	Action      string `json:"action"`      // "add" or "delete"
	Rule        string `json:"rule"`        // "accept" or "block"
	Protocol    string `json:"protocol"`    // "TCP", "UDP"
	Port        string `json:"port"`        // e.g. "80"
	Source      string `json:"source"`      // optional (Windows doesn't support source filtering easily)
	Destination string `json:"destination"` // optional
}

func HandleUpdateFirewall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req UpdateFirewallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError1(w, "Failed to parse request body", err)
		return
	}

	// Validate
	if req.Action != "add" && req.Action != "delete" {
		sendError1(w, "Action must be 'add' or 'delete'", fmt.Errorf("invalid action: %s", req.Action))
		return
	}
	if req.Rule != "accept" && req.Rule != "block" {
		sendError1(w, "Rule must be 'accept' or 'block'", fmt.Errorf("invalid rule: %s", req.Rule))
		return
	}
	if req.Protocol == "" || req.Port == "" {
		sendError1(w, "Protocol and port are required", fmt.Errorf("missing required fields"))
		return
	}

	ruleName := fmt.Sprintf("CustomRule-%s-%s", req.Protocol, req.Port)
	var psCmd string

	if req.Action == "add" {
		action := "Allow"
		if req.Rule == "block" {
			action = "Block"
		}
		psCmd = fmt.Sprintf(`New-NetFirewallRule -DisplayName "%s" -Direction Inbound -Action %s -Protocol %s -LocalPort %s -Profile Any`, ruleName, action, req.Protocol, req.Port)
	} else {
		psCmd = fmt.Sprintf(`Remove-NetFirewallRule -DisplayName "%s"`, ruleName)
	}

	// Execute PowerShell
	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError1(w, "Failed to update firewall rule", fmt.Errorf("%v, output: %s", err, string(output)))
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"operation": fmt.Sprintf("firewall rule %s", req.Action),
		"details":   fmt.Sprintf("Protocol: %s, Port: %s, Rule: %s", req.Protocol, req.Port, req.Rule),
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"user":      getCurrentUsername(),
	})
}

func sendError1(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "failed",
		"message": msg,
		"error":   err.Error(),
	})
}

// getCurrentUsername gets the current Windows username
func getCurrentUsername() string {
	out, err := exec.Command("whoami").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
