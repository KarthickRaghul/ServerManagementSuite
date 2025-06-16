package config_1

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

type BasicInfo struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
}

func HandleBasicInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	hostname, _ := exec.Command("hostname").Output()

	// Run timedatectl command
	timedatectlOutput, _ := exec.Command("timedatectl").Output()

	// Extract only the timezone information from timedatectl output
	timezone := extractTimezone(string(timedatectlOutput))

	data := BasicInfo{
		Hostname: string(hostname),
		Timezone: timezone,
	}

	json.NewEncoder(w).Encode(data)
}

// extractTimezone parses the timedatectl output to extract just the timezone value
func extractTimezone(output string) string {
	// Split output into lines
	lines := strings.Split(output, "\n")

	// Look for the line containing "Time zone:"
	for _, line := range lines {
		if strings.Contains(line, "Time zone:") {
			// Extract timezone using regex to get value after "Time zone:" and before " ("
			re := regexp.MustCompile(`Time zone:\s+([\w/]+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				return matches[1]
			}

			// If regex doesn't match, get everything after "Time zone:" and trim whitespace
			parts := strings.SplitN(line, "Time zone:", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	// Return empty string if timezone not found
	return ""
}
