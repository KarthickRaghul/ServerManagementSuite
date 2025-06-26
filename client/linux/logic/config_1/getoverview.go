package config_1

import (
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

type Overview struct {
	Uptime string `json:"uptime"`
}

func HandleOverview(w http.ResponseWriter, r *http.Request) {
	// Check for GET method
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Get system uptime
	uptime := getSystemUptime()

	// Prepare response data
	data := Overview{
		Uptime: uptime,
	}

	// Send successful GET response with data
	sendGetSuccess(w, data)
}

// getSystemUptime returns the system uptime in a human-readable format
func getSystemUptime() string {
	// Try uptime -p command first (Linux/macOS)
	if out, err := exec.Command("uptime", "-p").Output(); err == nil {
		uptime := strings.TrimSpace(string(out))
		// Remove "up " prefix if present
		uptime = strings.TrimPrefix(uptime, "up ")
		return uptime
	}

	// Fallback: Try regular uptime command and parse output
	if out, err := exec.Command("uptime").Output(); err == nil {
		uptimeStr := string(out)

		// Parse uptime from output like "up 2 days, 3:45"
		re := regexp.MustCompile(`up\s+(.+?),\s+\d+\s+users?`)
		matches := re.FindStringSubmatch(uptimeStr)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}

		// Alternative parsing for different uptime formats
		re2 := regexp.MustCompile(`up\s+(.+?)\s+load`)
		matches2 := re2.FindStringSubmatch(uptimeStr)
		if len(matches2) > 1 {
			return strings.TrimSpace(matches2[1])
		}
	}

	// If all else fails, return a default message
	return "Unable to determine uptime"
}

