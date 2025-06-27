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

// ✅ Fixed: Add Windows firewall rule with proper output handling
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

	// ✅ Fixed: Build PowerShell command with proper output suppression
	psCmd := fmt.Sprintf(`
	try {
		$rule = New-NetFirewallRule -DisplayName "%s" -Direction %s -Action %s -Protocol %s -LocalPort %s -Profile %s -Enabled True`,
		ruleName, direction, action, strings.ToUpper(req.Protocol), port, profile)

	// ✅ Add optional parameters if provided
	if req.RemotePort != "" {
		psCmd += fmt.Sprintf(` -RemotePort "%s"`, req.RemotePort)
	}
	if req.LocalAddress != "" && req.LocalAddress != "Any" {
		psCmd += fmt.Sprintf(` -LocalAddress "%s"`, req.LocalAddress)
	}
	if req.RemoteAddress != "" && req.RemoteAddress != "Any" {
		psCmd += fmt.Sprintf(` -RemoteAddress "%s"`, req.RemoteAddress)
	}
	if req.Program != "" {
		psCmd += fmt.Sprintf(` -Program "%s"`, req.Program)
	}
	if req.Service != "" {
		psCmd += fmt.Sprintf(` -Service "%s"`, req.Service)
	}

	// ✅ Complete the PowerShell command with proper success indication
	psCmd += `
		if ($rule) {
			Write-Output "SUCCESS: Firewall rule added successfully"
		} else {
			Write-Output "ERROR: Failed to create firewall rule"
		}
	} catch {
		Write-Output "ERROR: $($_.Exception.Message)"
	}`

	// Execute PowerShell command
	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Command execution failed: %v, output: %s", err, string(output))
	}

	result := strings.TrimSpace(string(output))
	lines := strings.Split(result, "\n")

	// ✅ Check the last line for SUCCESS/ERROR (ignore verbose object output)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "SUCCESS") {
			return true, ""
		} else if strings.HasPrefix(line, "ERROR") {
			return false, line
		}
	}

	// ✅ If no explicit SUCCESS/ERROR found, check if rule was created
	return verifyRuleExists(ruleName)
}

// ✅ Fixed: Delete Windows firewall rule with proper output handling
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

	// ✅ Fixed: Build PowerShell command with proper output handling
	psCmd := fmt.Sprintf(`
	try {
		$rules = Get-NetFirewallRule -DisplayName "%s" -ErrorAction SilentlyContinue
		if ($rules) {
			Remove-NetFirewallRule -DisplayName "%s" -Confirm:$false -ErrorAction Stop
			Write-Output "SUCCESS: Firewall rule deleted successfully"
		} else {
			Write-Output "SUCCESS: Rule not found (already deleted)"
		}
	} catch {
		Write-Output "ERROR: $($_.Exception.Message)"
	}`, ruleName, ruleName)

	// Execute PowerShell command
	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Command execution failed: %v, output: %s", err, string(output))
	}

	result := strings.TrimSpace(string(output))
	lines := strings.Split(result, "\n")

	// ✅ Check the last line for SUCCESS/ERROR
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "SUCCESS") {
			return true, ""
		} else if strings.HasPrefix(line, "ERROR") {
			return false, line
		}
	}

	return true, ""
}

// ✅ New helper function to verify rule creation
func verifyRuleExists(ruleName string) (bool, string) {
	psCmd := fmt.Sprintf(`
	try {
		$rule = Get-NetFirewallRule -DisplayName "%s" -ErrorAction SilentlyContinue
		if ($rule) {
			Write-Output "SUCCESS: Rule verified to exist"
		} else {
			Write-Output "ERROR: Rule not found after creation"
		}
	} catch {
		Write-Output "ERROR: $($_.Exception.Message)"
	}`, ruleName)

	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Verification failed: %v", err)
	}

	result := strings.TrimSpace(string(output))
	if strings.Contains(result, "SUCCESS") {
		return true, ""
	} else {
		return false, result
	}
}
