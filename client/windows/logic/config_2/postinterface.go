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

// HandleUpdateInterface handles the POST request to update interface status
func HandleUpdateInterface(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var request UpdateInterfaceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, "Failed to parse request body", err)
		return
	}

	if request.Interface == "" {
		sendError(w, "Interface name is required", fmt.Errorf("missing interface name"))
		return
	}

	if request.Status != "enable" && request.Status != "disable" {
		sendError(w, "Status must be 'enable' or 'disable'", fmt.Errorf("invalid status: %s", request.Status))
		return
	}

	// Normalize interface name
	actualInterface, err := resolveInterfaceName(request.Interface)
	if err != nil {
		sendError(w, "Could not resolve interface name", err)
		return
	}

	err = updateInterfaceStatus(actualInterface, request.Status)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to %s interface %s", request.Status, actualInterface), err)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Interface %s %sd successfully", actualInterface, request.Status),
	}

	fmt.Printf("Interface %s %sd by user %s\n", actualInterface, request.Status, "kishore-001")
	json.NewEncoder(w).Encode(response)
}

// resolveInterfaceName attempts to match user-friendly names to actual adapter names
func resolveInterfaceName(userInput string) (string, error) {
	userInput = strings.ToLower(strings.TrimSpace(userInput))

	// Special mapping for Wi-Fi
	if userInput == "wifi" || userInput == "wi-fi" {
		return "Wi-Fi", nil
	}

	// Check for variations of "Ethernet"
	if strings.Contains(userInput, "ethernet") {
		// Get all interface names
		cmd := exec.Command("powershell", "-Command", "Get-NetAdapter | Select-Object -ExpandProperty Name")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to get network interfaces: %v", err)
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToLower(line), "ethernet") {
				return line, nil // e.g., Ethernet, Ethernet 2, Ethernet3
			}
		}

		return "", fmt.Errorf("no ethernet interfaces found")
	}

	// Default fallback to original input
	return userInput, nil
}

// updateInterfaceStatus enables or disables a network interface on Windows
func updateInterfaceStatus(interfaceName, status string) error {
	var cmd *exec.Cmd
	action := "enabled"
	if status == "disable" {
		action = "disabled"
	}

	powershellCmd := fmt.Sprintf(`netsh interface set interface name="%s" admin=%s`, interfaceName, action)
	cmd = exec.Command("powershell", "-Command", powershellCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command execution failed: %v, output: %s", err, string(output))
	}

	return nil
}

// sendError sends a JSON error response
func sendError(w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "failed",
		"message": message,
		"error":   err.Error(),
	})
}
