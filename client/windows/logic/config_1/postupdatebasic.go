package config_1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

// BasicUpdateRequest defines the expected JSON structure for basic system updates
type BasicUpdateRequest struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
}

// HandleBasicUpdate processes requests to update hostname and timezone (Windows)
func HandleBasicUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var updateReq BasicUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if updateReq.Hostname == "" && updateReq.Timezone == "" {
		sendError(w, "At least one field (hostname or timezone) must be provided", http.StatusBadRequest)
		return
	}

	var errorMessages []string

	if updateReq.Hostname != "" {
		if err := updateHostname(updateReq.Hostname); err != nil {
			errorMessages = append(errorMessages, "hostname update failed: "+err.Error())
		}
	}

	if updateReq.Timezone != "" {
		if err := updateTimezone(updateReq.Timezone); err != nil {
			errorMessages = append(errorMessages, "timezone update failed: "+err.Error())
		}
	}

	if len(errorMessages) > 0 {
		sendError(w, strings.Join(errorMessages, "; "), http.StatusInternalServerError)
		return
	}

	sendPostSuccess(w)
}

// updateHostname uses PowerShell to rename the computer
// updateHostname uses PowerShell to rename the computer
func updateHostname(newHostname string) error {
	// Get current hostname
	currentHostnameBytes, err := exec.Command("cmd", "/C", "hostname").Output()
	if err != nil {
		return fmt.Errorf("failed to get current hostname: %v", err)
	}
	currentHostname := strings.TrimSpace(string(currentHostnameBytes))

	// If new hostname is same as current, return nil (no error)
	if strings.EqualFold(currentHostname, newHostname) {
		return nil
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Rename-Computer -NewName %s -Force -PassThru`, quoteArg(newHostname)),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rename computer: %v (%s)", err, string(output))
	}
	return nil
}

// updateTimezone sets the Windows timezone using PowerShell
func updateTimezone(tz string) error {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Set-TimeZone -Name %s`, quoteArg(tz)),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set timezone: %v (%s)", err, string(output))
	}
	return nil
}

// quoteArg safely escapes strings for PowerShell
func quoteArg(arg string) string {
	return `"` + strings.ReplaceAll(arg, `"`, `""`) + `"`
}
