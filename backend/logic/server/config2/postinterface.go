package config2

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

type interfaceUpdateRequest struct {
	Host      string `json:"host"`
	Interface string `json:"interface"`
	Status    string `json:"status"`
}

type responseJSON struct {
	Status string `json:"status"`
}

func HandlePostInterface(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req interfaceUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" || req.Interface == "" || req.Status == "" {
			sendError(w, "Invalid request body: host, interface, and status are required", http.StatusBadRequest)
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
		clientURL := config.GetClientURL(req.Host, "/client/config2/updateinterface")

		// Prepare payload for client
		clientPayload := processInterfaceUpdateRequest(req, device.Os)

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			sendError(w, "Failed to prepare request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			sendError(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(clientReq)
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
		var clientResp responseJSON
		if err := json.Unmarshal(body, &clientResp); err != nil {
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processInterfaceUpdateResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
	}
}

// Process interface update request based on OS
func processInterfaceUpdateRequest(req interfaceUpdateRequest, osType string) map[string]string {
	if strings.ToLower(osType) == "windows" {
		return processWindowsInterfaceUpdateRequest(req)
	}

	// Default Linux behavior
	return map[string]string{
		"interface": strings.TrimSpace(req.Interface),
		"status":    strings.TrimSpace(req.Status),
	}
}

// Process interface update response based on OS
func processInterfaceUpdateResponse(resp responseJSON, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsInterfaceUpdateResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific interface update request processing (placeholder for future differences)
func processWindowsInterfaceUpdateRequest(req interfaceUpdateRequest) map[string]string {
	// For now, return same format as Linux
	return map[string]string{
		"interface": strings.TrimSpace(req.Interface),
		"status":    strings.TrimSpace(req.Status),
	}
}

// Windows-specific interface update response processing (placeholder for future differences)
func processWindowsInterfaceUpdateResponse(resp responseJSON) interface{} {
	// For now, return same format as Linux
	return resp
}
