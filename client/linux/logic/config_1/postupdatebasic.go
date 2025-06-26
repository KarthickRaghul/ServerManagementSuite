package config_1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// BasicUpdateRequest defines the expected JSON structure for basic system updates
type BasicUpdateRequest struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
}

// HandleBasicUpdate processes requests to update hostname and timezone
func HandleBasicUpdate(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse the JSON request
	var updateReq BasicUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate that at least one field is provided
	if updateReq.Hostname == "" && updateReq.Timezone == "" {
		sendError(w, "At least one field (hostname or timezone) must be provided", http.StatusBadRequest)
		return
	}

	// Track update status and error messages
	var errorMessages []string

	// Update hostname if provided
	if updateReq.Hostname != "" {
		if err := updateHostname(updateReq.Hostname); err != nil {
			errorMessages = append(errorMessages, "hostname update failed: "+err.Error())
		}
	}

	// Update timezone if provided
	if updateReq.Timezone != "" {
		if err := updateTimezone(updateReq.Timezone); err != nil {
			errorMessages = append(errorMessages, "timezone update failed: "+err.Error())
		}
	}

	// If any errors occurred, send failed response with combined message
	if len(errorMessages) > 0 {
		sendError(w, strings.Join(errorMessages, "; "), http.StatusInternalServerError)
		return
	}

	// Send success response
	sendPostSuccess(w)
}

// updateHostname changes the system hostname
func updateHostname(hostname string) error {
	// Basic validation
	if strings.TrimSpace(hostname) == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	// Check if running as root/sudo
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root/sudo to change hostname")
	}

	// Execute hostname command to change hostname immediately
	cmd := exec.Command("hostname", hostname)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set hostname temporarily: %w", err)
	}

	// Update /etc/hostname file for persistence across reboots
	if err := os.WriteFile("/etc/hostname", []byte(hostname+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to update /etc/hostname: %w", err)
	}

	// Also update /etc/hosts to ensure the hostname resolves locally
	hostsCmd := exec.Command("sed", "-i", fmt.Sprintf("s/127.0.1.1.*/127.0.1.1\t%s/", hostname), "/etc/hosts")
	if err := hostsCmd.Run(); err != nil {
		return fmt.Errorf("failed to update /etc/hosts: %w", err)
	}

	return nil
}

// updateTimezone changes the system timezone
func updateTimezone(timezone string) error {
	// Basic validation
	if strings.TrimSpace(timezone) == "" {
		return fmt.Errorf("timezone cannot be empty")
	}

	// Check if running as root/sudo
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root/sudo to change timezone")
	}

	// Validate timezone by checking if it exists in /usr/share/zoneinfo/
	_, err := os.Stat(fmt.Sprintf("/usr/share/zoneinfo/%s", timezone))
	if err != nil {
		return fmt.Errorf("invalid timezone: %s", timezone)
	}

	// Execute timedatectl to set timezone
	cmd := exec.Command("timedatectl", "set-timezone", timezone)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set timezone with timedatectl: %w", err)
	}

	return nil
}
