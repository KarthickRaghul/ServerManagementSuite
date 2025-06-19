package config_1

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

// BasicUpdateRequest defines the expected JSON structure
type BasicUpdateRequest struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
}

// HandleBasicUpdate processes hostname and timezone updates for Windows
func HandleBasicUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var updateReq BasicUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		log.Println("Failed to decode request:", err)
		writeStatus(w, false)
		return
	}

	log.Printf("Received Update Request: Hostname=%s, Timezone=%s\n", updateReq.Hostname, updateReq.Timezone)

	status := true

	if updateReq.Hostname != "" {
		if err := updateHostname(updateReq.Hostname); err != nil {
			log.Println("Hostname update error:", err)
			status = false
		}
	}

	if updateReq.Timezone != "" {
		if err := updateTimezone(updateReq.Timezone); err != nil {
			log.Println("Timezone update error:", err)
			status = false
		}
	}

	writeStatus(w, status)
}

func writeStatus(w http.ResponseWriter, success bool) {
	w.Header().Set("Content-Type", "application/json")
	status := "success"
	if !success {
		status = "failed"
	}
	json.NewEncoder(w).Encode(map[string]string{
		"status": status,
	})
}

// updateHostname uses PowerShell to rename the computer
func updateHostname(newHostname string) error {
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
