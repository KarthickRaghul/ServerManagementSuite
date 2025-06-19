package config_1

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

type BasicInfo struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
}

func HandleBasicInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Only GET allowed",
		})
		return
	}

	hostnameBytes, err := exec.Command("cmd", "/C", "hostname").Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Failed to get hostname",
		})
		return
	}
	hostname := strings.TrimSpace(string(hostnameBytes))

	timezoneBytes, err := exec.Command("powershell", "-Command", "(Get-TimeZone).Id").Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Failed to get timezone",
		})
		return
	}
	timezone := strings.TrimSpace(string(timezoneBytes))

	// On success, return plain data
	json.NewEncoder(w).Encode(BasicInfo{
		Hostname: hostname,
		Timezone: timezone,
	})
}
