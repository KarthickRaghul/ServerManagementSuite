package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

// UpdateInterfaceRequest represents the request format for updating an interface
type UpdateInterfaceRequest struct {
	Interface string `json:"interface"`
	Status    string `json:"status"`
}

// HandleUpdateInterface handles the POST request to update interface status (Windows version)
func HandleUpdateInterface(w http.ResponseWriter, r *http.Request) {
	// Check for POST method - same as Linux
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var request UpdateInterfaceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate interface name
	if request.Interface == "" {
		sendError(w, "Interface name is required", http.StatusBadRequest)
		return
	}

	// Validate status
	if request.Status != "enable" && request.Status != "disable" {
		sendError(w, "Status must be 'enable' or 'disable'", http.StatusBadRequest)
		return
	}

	// Update interface status
	err = updateInterfaceStatusWindows(request.Interface, request.Status)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to %s interface %s: %v", request.Status, request.Interface, err), http.StatusInternalServerError)
		return
	}

	// Send standard success response - exactly like Linux
	sendPostSuccess(w)
}

// updateInterfaceStatusWindows enables or disables a network interface using PowerShell
func updateInterfaceStatusWindows(interfaceName, status string) error {
	action := "Disable-NetAdapter"
	if status == "enable" {
		action = "Enable-NetAdapter"
	}

	// Enhanced PowerShell command with error handling
	psCmd := fmt.Sprintf(`
	try {
		%s -Name "%s" -Confirm:$false
		Write-Output "SUCCESS: Interface %s %sd"
	} catch {
		Write-Output "ERROR: $($_.Exception.Message)"
	}
	`, action, interfaceName, interfaceName, status)

	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command execution failed: %v, output: %s", err, string(output))
	}

	result := strings.TrimSpace(string(output))
	if strings.HasPrefix(result, "SUCCESS") {
		fmt.Printf("✅ Interface %s %sd successfully\n", interfaceName, status)
		return nil
	} else {
		return fmt.Errorf("PowerShell error: %s", result)
	}
}
