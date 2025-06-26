package config_1

import (
	"net/http"
	"os/exec"
	"strings"
)

type Overview struct {
	Uptime string `json:"uptime"`
}

func HandleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := getSystemUptimeWindows()

	data := Overview{
		Uptime: uptime,
	}

	sendGetSuccess(w, data)
}

// getSystemUptimeWindows returns system uptime in a human-readable format
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
