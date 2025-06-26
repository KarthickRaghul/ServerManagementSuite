package config1

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

type frontendRequestbasic struct {
	HostName string `json:"hostname"`
	TimeZone string `json:"timezone"`
	Host     string `json:"host"`
}

type clientResponsebasic struct {
	Status string `json:"status"`
}

func HandleBasicChange(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req frontendRequestbasic
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.HostName == "" || req.TimeZone == "" {
			sendError(w, "Host, hostname, and timezone are required", http.StatusBadRequest)
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

		// Process request based on OS
		clientPayload := processBasicChangeRequest(req, device.Os)

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			sendError(w, "Failed to prepare request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/config1/basic_update")

		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			sendError(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Set headers with authorization token
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// Send request to client with timeout
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}

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
				sendError(w, "Client error: "+string(body), http.StatusBadGateway)
			}
			return
		}

		// Read client response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			sendError(w, "Failed to read client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Parse client response
		var clientResp clientResponsebasic
		if err := json.Unmarshal(body, &clientResp); err != nil {
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processBasicChangeResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
	}
}

// Process request based on OS
func processBasicChangeRequest(req frontendRequestbasic, osType string) map[string]string {
	if strings.ToLower(osType) == "windows" {
		return processWindowsBasicChangeRequest(req)
	}

	// Default Linux behavior
	return map[string]string{
		"hostname": strings.TrimSpace(req.HostName),
		"timezone": strings.TrimSpace(req.TimeZone),
	}
}

// Process response based on OS
func processBasicChangeResponse(resp clientResponsebasic, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsBasicChangeResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific request processing (placeholder for future differences)
func processWindowsBasicChangeRequest(req frontendRequestbasic) map[string]string {
	// For now, return same format as Linux
	// Add Windows-specific fields when needed
	return map[string]string{
		"hostname": strings.TrimSpace(req.HostName),
		"timezone": strings.TrimSpace(req.TimeZone),
	}
}

// Windows-specific response processing (placeholder for future differences)
func processWindowsBasicChangeResponse(resp clientResponsebasic) interface{} {
	// For now, return same format as Linux
	// Add Windows-specific processing when needed
	return resp
}
