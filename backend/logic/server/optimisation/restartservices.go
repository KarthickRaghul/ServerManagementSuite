package optimisation

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

type restartServiceRequest struct {
	Host    string `json:"host"`
	Service string `json:"service"`
}

func PostRestartService(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req restartServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.Service == "" {
			sendError(w, "Host and service are required", http.StatusBadRequest)
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
		clientPayload := processRestartServiceRequest(req, device.Os)

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			sendError(w, "Failed to prepare request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/resource/restartservice")

		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			sendError(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

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
		var clientResp interface{}
		if err := json.Unmarshal(body, &clientResp); err != nil {
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processRestartServiceResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
	}
}

// Process restart service request based on OS
func processRestartServiceRequest(req restartServiceRequest, osType string) map[string]string {
	if strings.ToLower(osType) == "windows" {
		return processWindowsRestartServiceRequest(req)
	}

	// Default Linux behavior
	return map[string]string{
		"service": strings.TrimSpace(req.Service),
	}
}

// Process restart service response based on OS
func processRestartServiceResponse(resp interface{}, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsRestartServiceResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific restart service request processing (placeholder for future differences)
func processWindowsRestartServiceRequest(req restartServiceRequest) map[string]string {
	// For now, return same format as Linux
	// Future: Windows might have different service names or commands
	return map[string]string{
		"service": strings.TrimSpace(req.Service),
	}
}

// Windows-specific restart service response processing (placeholder for future differences)
func processWindowsRestartServiceResponse(resp interface{}) interface{} {
	// For now, return same format as Linux
	// Future: Windows might have different response structure
	return resp
}
