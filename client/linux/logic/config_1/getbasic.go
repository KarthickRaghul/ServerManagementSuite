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

// Standard response structures
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func HandleBasicInfo(w http.ResponseWriter, r *http.Request) {
	// Check for GET method
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Get hostname
	hostnameOutput, err := exec.Command("hostname").Output()
	if err != nil {
		sendError(w, "Failed to get hostname: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get timezone information
	timedatectlOutput, err := exec.Command("timedatectl").Output()
	if err != nil {
		sendError(w, "Failed to get timezone information: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract timezone from timedatectl output
	timezone := extractTimezone(string(timedatectlOutput))
	if timezone == "" {
		sendError(w, "Failed to extract timezone from system output", http.StatusInternalServerError)
		return
	}

	// Prepare response data
	data := BasicInfo{
		Hostname: strings.TrimSpace(string(hostnameOutput)),
		Timezone: timezone,
	}

	// Send successful GET response with data
	sendGetSuccess(w, data)
}

// sendGetSuccess sends successful GET response with data
func sendGetSuccess(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// sendError sends standardized error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	errorResp := ErrorResponse{
		Status:  "failed",
		Message: message,
	}
	json.NewEncoder(w).Encode(errorResp)
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
				// Extract just the timezone part before any additional info
				timezone := strings.TrimSpace(parts[1])
				// Remove anything after the first space or parenthesis
				if idx := strings.Index(timezone, " "); idx != -1 {
					timezone = timezone[:idx]
				}
				if idx := strings.Index(timezone, "("); idx != -1 {
					timezone = timezone[:idx]
				}
				return strings.TrimSpace(timezone)
			}
		}
	}

	// Return empty string if timezone not found
	return ""
}

// sendPostSuccess sends successful POST response
func sendPostSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	response := SuccessResponse{
		Status: "success",
	}
	json.NewEncoder(w).Encode(response)
}
