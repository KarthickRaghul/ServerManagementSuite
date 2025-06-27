package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

// Updated to match backend request structure
type UpdateFirewallRequest struct {
	Action      string `json:"action"`      // "add" or "delete"
	Rule        string `json:"rule"`        // "accept" or "block" (Linux compatibility)
	Protocol    string `json:"protocol"`    // "TCP", "UDP"
	Port        string `json:"port"`        // e.g. "80"
	Source      string `json:"source"`      // optional
	Destination string `json:"destination"` // optional

	// Windows-specific fields from backend
	Name          string `json:"name,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	Direction     string `json:"direction,omitempty"`     // Inbound/Outbound
	ActionType    string `json:"actionType,omitempty"`    // Allow/Block
	Enabled       string `json:"enabled,omitempty"`       // True/False
	Profile       string `json:"profile,omitempty"`       // Public/Private/Domain/Any
	LocalPort     string `json:"localPort,omitempty"`     // Windows local port
	RemotePort    string `json:"remotePort,omitempty"`    // Windows remote port
	LocalAddress  string `json:"localAddress,omitempty"`  // Windows local address
	RemoteAddress string `json:"remoteAddress,omitempty"` // Windows remote address
	Program       string `json:"program,omitempty"`       // Windows program path
	Service       string `json:"service,omitempty"`       // Windows service name
}

func HandleUpdateFirewall(w http.ResponseWriter, r *http.Request) {
	// Check for POST method - same as Linux
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	var req UpdateFirewallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Enhanced validation for both Linux and Windows fields
	if req.Action != "add" && req.Action != "delete" {
		sendError(w, "Action must be 'add' or 'delete'", http.StatusBadRequest)
		return
	}

	// Windows firewall rule creation/deletion
	var success bool
	var errorMsg string

	if req.Action == "add" {
		success, errorMsg = addWindowsFirewallRule(req)
	} else {
		success, errorMsg = deleteWindowsFirewallRule(req)
	}

	if !success {
		sendError(w, errorMsg, http.StatusInternalServerError)
		return
	}

	// Send standard success response - exactly like Linux
	sendPostSuccess(w)
}

// Add Windows firewall rule with enhanced logic
func addWindowsFirewallRule(req UpdateFirewallRequest) (bool, string) {
	// Determine rule name - use provided name or generate one
	ruleName := req.Name
	if ruleName == "" && req.DisplayName != "" {
		ruleName = req.DisplayName
	}
	if ruleName == "" {
		// Generate rule name from protocol and port
		ruleName = fmt.Sprintf("SMS-Rule-%s-%s", strings.ToUpper(req.Protocol), req.Port)
	}

	// Determine action - support both Linux (rule) and Windows (actionType) formats
	action := "Allow"
	if req.Rule == "block" || req.ActionType == "Block" {
		action = "Block"
	}

	// Determine direction - default to Inbound if not specified
	direction := req.Direction
	if direction == "" {
		direction = "Inbound"
	}

	// Determine profile - default to Any if not specified
	profile := req.Profile
	if profile == "" {
		profile = "Any"
	}

	// Determine port - use localPort if available, otherwise use port
	port := req.LocalPort
	if port == "" {
		port = req.Port
	}

	// Build PowerShell command
	psCmd := fmt.Sprintf(`
	try {
		New-NetFirewallRule -DisplayName "%s" -Direction %s -Action %s -Protocol %s -LocalPort %s -Profile %s -Enabled True
		Write-Output "SUCCESS: Firewall rule added"
	} catch {
		Write-Output "ERROR: $($_.Exception.Message)"
	}
	`, ruleName, direction, action, strings.ToUpper(req.Protocol), port, profile)

	// Execute PowerShell command
	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Command execution failed: %v, output: %s", err, string(output))
	}

	result := strings.TrimSpace(string(output))
	if strings.HasPrefix(result, "SUCCESS") {
		return true, ""
	} else {
		return false, result
	}
}

// Delete Windows firewall rule
func deleteWindowsFirewallRule(req UpdateFirewallRequest) (bool, string) {
	// Determine rule name to delete
	ruleName := req.Name
	if ruleName == "" && req.DisplayName != "" {
		ruleName = req.DisplayName
	}
	if ruleName == "" {
		// Generate rule name from protocol and port (same as add)
		ruleName = fmt.Sprintf("SMS-Rule-%s-%s", strings.ToUpper(req.Protocol), req.Port)
	}

	// Build PowerShell command to remove rule
	psCmd := fmt.Sprintf(`
	try {
		Remove-NetFirewallRule -DisplayName "%s" -ErrorAction Stop
		Write-Output "SUCCESS: Firewall rule deleted"
	} catch {
		if ($_.Exception.Message -like "*No MSFT_NetFirewallRule objects found*") {
			Write-Output "SUCCESS: Rule not found (already deleted)"
		} else {
			Write-Output "ERROR: $($_.Exception.Message)"
		}
	}
	`, ruleName)

	// Execute PowerShell command
	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Command execution failed: %v, output: %s", err, string(output))
	}

	result := strings.TrimSpace(string(output))
	if strings.HasPrefix(result, "SUCCESS") {
		return true, ""
	} else {
		return false, result
	}
}

// getCurrentUsername gets the current Windows username
func getCurrentUsername() string {
	out, err := exec.Command("whoami").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

