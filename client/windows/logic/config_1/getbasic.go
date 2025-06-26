package config_1

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)


func sendPostSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	response := SuccessResponse{
		Status: "success",
	}
	json.NewEncoder(w).Encode(response)
}

type BasicInfo struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Status string `json:"status"`
}

func HandleBasicInfo(w http.ResponseWriter, r *http.Request) {
	// Check for GET method
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get hostname
	hostnameOutput, err := exec.Command("cmd", "/C", "hostname").Output()
	if err != nil {
		sendError(w, "Failed to get hostname: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get timezone
	timezoneOutput, err := exec.Command("powershell", "-Command", "(Get-TimeZone).Id").Output()
	if err != nil {
		sendError(w, "Failed to get timezone: "+err.Error(), http.StatusInternalServerError)
		return
	}

	timezone := strings.TrimSpace(string(timezoneOutput))
	hostname := strings.TrimSpace(string(hostnameOutput))

	// Prepare response data
	data := BasicInfo{
		Hostname: hostname,
		Timezone: timezone,
	}

	// Send success response
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

