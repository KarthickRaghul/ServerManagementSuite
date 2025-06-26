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

	"backend/config"
	serverdb "backend/db/gen/server"
)

type SSHKeyManagementRequest struct {
	Key  string `json:"key"`
	Host string `json:"host"`
}

type SSHKeyManagementResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func HandleSSHKeyManagement(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req SSHKeyManagementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.Key == "" {
			sendError(w, "Host and key are required", http.StatusBadRequest)
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

		// Process SSH key request based on OS
		clientPayload := processSSHKeyRequest(req, device.Os)

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			sendError(w, "Failed to prepare request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/config1/ssh")

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
		var clientResp SSHKeyManagementResponse
		if err := json.Unmarshal(body, &clientResp); err != nil {
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processSSHKeyResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
	}
}

// Process SSH key request based on OS
func processSSHKeyRequest(req SSHKeyManagementRequest, osType string) map[string]string {
	if strings.ToLower(osType) == "windows" {
		return processWindowsSSHKeyRequest(req)
	}

	// Default Linux behavior
	return map[string]string{
		"key": strings.TrimSpace(req.Key),
	}
}

// Process SSH key response based on OS
func processSSHKeyResponse(resp SSHKeyManagementResponse, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsSSHKeyResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific SSH key request processing (placeholder for future differences)
func processWindowsSSHKeyRequest(req SSHKeyManagementRequest) map[string]string {
	// For now, return same format as Linux
	// Future: Windows might use different SSH key management (OpenSSH for Windows)
	return map[string]string{
		"key": strings.TrimSpace(req.Key),
	}
}

// Windows-specific SSH key response processing (placeholder for future differences)
func processWindowsSSHKeyResponse(resp SSHKeyManagementResponse) interface{} {
	// For now, return same format as Linux
	// Future: might need different response handling for Windows SSH
	return resp
}
