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

type networkUpdateRequest struct {
	Host    string `json:"host"`
	Method  string `json:"method"`
	IP      string `json:"ip,omitempty"`
	Subnet  string `json:"subnet,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	DNS     string `json:"dns,omitempty"`
}

type responseJSON1 struct {
	Status string `json:"status"`
}

func HandlePostNetwork(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req networkUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" || req.Method == "" {
			sendError(w, "Invalid request body: host and method are required", http.StatusBadRequest)
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
		clientPayload := processNetworkUpdateRequest(req, device.Os)

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			sendError(w, "Failed to prepare request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/config2/updatenetwork")

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
		var clientResp responseJSON1
		if err := json.Unmarshal(body, &clientResp); err != nil {
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processNetworkUpdateResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
	}
}

// Process network update request based on OS
func processNetworkUpdateRequest(req networkUpdateRequest, osType string) map[string]string {
	if strings.ToLower(osType) == "windows" {
		return processWindowsNetworkUpdateRequest(req)
	}

	// Default Linux behavior
	return map[string]string{
		"method":  strings.TrimSpace(req.Method),
		"ip":      strings.TrimSpace(req.IP),
		"subnet":  strings.TrimSpace(req.Subnet),
		"gateway": strings.TrimSpace(req.Gateway),
		"dns":     strings.TrimSpace(req.DNS),
	}
}

// Process network update response based on OS
func processNetworkUpdateResponse(resp responseJSON1, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsNetworkUpdateResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific network update request processing (placeholder for future differences)
func processWindowsNetworkUpdateRequest(req networkUpdateRequest) map[string]string {
	// For now, return same format as Linux
	return map[string]string{
		"method":  strings.TrimSpace(req.Method),
		"ip":      strings.TrimSpace(req.IP),
		"subnet":  strings.TrimSpace(req.Subnet),
		"gateway": strings.TrimSpace(req.Gateway),
		"dns":     strings.TrimSpace(req.DNS),
	}
}

// Windows-specific network update response processing (placeholder for future differences)
func processWindowsNetworkUpdateResponse(resp responseJSON1) interface{} {
	// For now, return same format as Linux
	return resp
}
