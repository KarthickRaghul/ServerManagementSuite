package config1

import (
	"backend/config"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	serverdb "backend/db/gen/server"
)

// Request body from frontend
type basicRequest struct {
	Host string `json:"host"`
}

// Response from client (remote server)
type clientBasicResponse struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
}

// Standard response structures
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Handler for /api/admin/server/config1/basic
func HandleBasic(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req basicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
			sendError(w, "Invalid request body: host is required", http.StatusBadRequest)
			return
		}

		// Lookup device and get access token
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			sendError(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/config1/basic")

		clientReq, err := http.NewRequest("GET", clientURL, nil)
		if err != nil {
			sendError(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// Send request to client
		httpClient := &http.Client{}
		resp, err := httpClient.Do(clientReq)
		if err != nil {
			sendError(w, "Failed to reach client: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Handle client error responses
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)

			// Try to parse client error response (standardized format)
			var clientError ErrorResponse
			if json.Unmarshal(body, &clientError) == nil && clientError.Status == "failed" {
				sendError(w, "Client error: "+clientError.Message, http.StatusBadGateway)
			} else {
				sendError(w, fmt.Sprintf("Client error: %s", string(body)), http.StatusBadGateway)
			}
			return
		}

		// Parse client response
		var clientResp clientBasicResponse
		if err := json.NewDecoder(resp.Body).Decode(&clientResp); err != nil {
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processBasicResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
	}
}

// Standard response functions
func sendGetSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResp := ErrorResponse{
		Status:  "failed",
		Message: message,
	}
	json.NewEncoder(w).Encode(errorResp)
}

// Process response based on OS from database
func processBasicResponse(resp clientBasicResponse, osType string) interface{} {
	// Clean data
	hostname := strings.TrimSpace(resp.Hostname)
	timezone := strings.TrimSpace(resp.Timezone)

	// Check OS type from database
	if strings.ToLower(osType) == "windows" {
		return processWindowsBasicResponse(hostname, timezone)
	}

	// Default to Linux behavior
	return clientBasicResponse{
		Hostname: hostname,
		Timezone: timezone,
	}
}

// Windows-specific processing (placeholder for future differences)
func processWindowsBasicResponse(hostname, timezone string) interface{} {
	// For now, return same format as Linux
	// Add Windows-specific fields when needed
	return clientBasicResponse{
		Hostname: hostname,
		Timezone: timezone,
	}
}

