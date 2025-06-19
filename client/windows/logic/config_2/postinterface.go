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
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var request UpdateInterfaceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError1(w, "Failed to parse request body", err)
		return
	}

	if request.Interface == "" {
		sendError1(w, "Interface name is required", fmt.Errorf("missing interface name"))
		return
	}

	if request.Status != "enable" && request.Status != "disable" {
		sendError1(w, "Status must be 'enable' or 'disable'", fmt.Errorf("invalid status: %s", request.Status))
		return
	}

	err = updateInterfaceStatusWindows(request.Interface, request.Status)
	if err != nil {
		sendError1(w, fmt.Sprintf("Failed to %s interface %s", request.Status, request.Interface), err)
		return
	}

	response := map[string]string{
		"status":    "success",
		"message":   fmt.Sprintf("Interface %s %sd successfully", request.Interface, request.Status),
		"timestamp": "2025-05-30 14:39:25",
		"user":      "kishore-001",
	}

	fmt.Printf("Interface %s %sd by user %s\n", request.Interface, request.Status, "kishore-001")
	json.NewEncoder(w).Encode(response)
}

// updateInterfaceStatusWindows enables or disables a network interface using PowerShell
func updateInterfaceStatusWindows(interfaceName, status string) error {
	action := "Disable-NetAdapter"
	if status == "enable" {
		action = "Enable-NetAdapter"
	}

	// Construct PowerShell command
	psCmd := fmt.Sprintf(`%s -Name "%s" -Confirm:$false`, action, interfaceName)
	cmd := exec.Command("powershell", "-Command", psCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PowerShell command failed: %v\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}
