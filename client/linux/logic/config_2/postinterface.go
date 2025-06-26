package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

// UpdateInterfaceRequest represents the request format for updating an interface
type UpdateInterfaceRequest struct {
	Interface string `json:"interface"`
	Status    string `json:"status"`
}

// HandleUpdateInterface handles the POST request to update interface status
func HandleUpdateInterface(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse the request body
	var request UpdateInterfaceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validate request data
	if request.Interface == "" {
		sendError(w, "Interface name is required", http.StatusBadRequest)
		return
	}

	if request.Status != "enable" && request.Status != "disable" {
		sendError(w, "Status must be 'enable' or 'disable'", http.StatusBadRequest)
		return
	}

	// Update interface status
	err = updateInterfaceStatus(request.Interface, request.Status)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to %s interface %s", request.Status, request.Interface), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]string{
		"status":    "success",
		"message":   fmt.Sprintf("Interface %s %sd successfully", request.Interface, request.Status),
		"timestamp": "2025-05-30 14:39:25", // Using provided UTC timestamp
		"user":      "kishore-001",         // Using provided user login
	}

	// Send the response
	fmt.Printf("Interface %s %sd by user %s\n", request.Interface, request.Status, "kishore-001")
	json.NewEncoder(w).Encode(response)
}

// updateInterfaceStatus enables or disables a network interface
func updateInterfaceStatus(interfaceName, status string) error {
	// Determine the command to run based on status
	var cmd *exec.Cmd

	if status == "enable" {
		cmd = exec.Command("ip", "link", "set", "dev", interfaceName, "up")
	} else { // status == "disable"
		cmd = exec.Command("ip", "link", "set", "dev", interfaceName, "down")
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command execution failed: %v, output: %s", err, string(output))
	}

	return nil
}
