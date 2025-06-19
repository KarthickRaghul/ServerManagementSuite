package config_1

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

type Overview struct {
	Uptime string `json:"uptime"`
}

func HandleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	data := Overview{
		Uptime: getSystemUptimeWindows(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// getSystemUptimeWindows returns system uptime using PowerShell
func getSystemUptimeWindows() string {
	cmd := exec.Command("powershell", "-Command", `
		$uptime = (Get-CimInstance Win32_OperatingSystem).LastBootUpTime
		$uptimeSpan = (Get-Date) - $uptime
		"{0} days, {1} hours, {2} minutes" -f $uptimeSpan.Days, $uptimeSpan.Hours, $uptimeSpan.Minutes
	`)

	out, err := cmd.Output()
	if err != nil {
		return "Unable to determine uptime"
	}
	return strings.TrimSpace(string(out))
}
