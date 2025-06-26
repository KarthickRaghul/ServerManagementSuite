package config2

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

type routerUpdateRequest struct {
	Host        string `json:"host"`
	Action      string `json:"action"`      // "add" or "delete"
	Destination string `json:"destination"` // e.g., "192.168.2.0/24"
	Gateway     string `json:"gateway"`     // e.g., "192.168.1.1"
	Interface   string `json:"interface"`   // optional
	Metric      string `json:"metric"`      // optional
}

func HandlePostUpdateRouter(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req routerUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("❌ [ROUTE] Failed to parse request: %v", err)
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.Action == "" || req.Destination == "" || req.Gateway == "" {
			log.Printf("❌ [ROUTE] Missing required fields")
			sendError(w, "Host, action, destination, and gateway are required", http.StatusBadRequest)
			return
		}

		log.Printf("🔍 [ROUTE] Processing route %s for host: %s, destination: %s", req.Action, req.Host, req.Destination)

		// Lookup device and get access token
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			log.Printf("❌ [ROUTE] Device not found: %s", req.Host)
			sendError(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("❌ [ROUTE] Database error: %v", err)
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Process route update request based on OS
		clientPayload := processRouteUpdateRequest(req, device.Os)

		// Convert to JSON for client
		clientBody, err := json.Marshal(clientPayload)
		if err != nil {
			log.Printf("❌ [ROUTE] Failed to marshal client payload: %v", err)
			sendError(w, "Failed to prepare client request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("🔍 [ROUTE] Client payload: %s", string(clientBody))

		// Use config for client URL
		clientURL := config.GetClientURL(req.Host, "/client/config2/updateroute")
		log.Printf("🔍 [ROUTE] Sending request to: %s", clientURL)

		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(clientBody))
		if err != nil {
			log.Printf("❌ [ROUTE] Failed to create client request: %v", err)
			sendError(w, "Failed to create client request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// Add timeout to prevent hanging
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := httpClient.Do(clientReq)
		if err != nil {
			log.Printf("❌ [ROUTE] Failed to reach client: %v", err)
			sendError(w, "Failed to reach client: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		log.Printf("🔍 [ROUTE] Client response status: %d", resp.StatusCode)

		// Handle client error responses
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("❌ [ROUTE] Client returned non-200 status: %d", resp.StatusCode)

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
			log.Printf("❌ [ROUTE] Failed to read response body: %v", err)
			sendError(w, "Failed to read client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		log.Printf("🔍 [ROUTE] Client response body: %s", string(body))

		// Parse client response
		var clientResp interface{}
		if err := json.Unmarshal(body, &clientResp); err != nil {
			log.Printf("❌ [ROUTE] Failed to parse client response: %v", err)
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processRouteUpdateResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
		log.Printf("✅ [ROUTE] Route %s completed successfully for %s", req.Action, req.Host)
	}
}

// Process route update request based on OS
func processRouteUpdateRequest(req routerUpdateRequest, osType string) map[string]string {
	if strings.ToLower(osType) == "windows" {
		return processWindowsRouteUpdateRequest(req)
	}

	// Default Linux behavior
	clientPayload := map[string]string{
		"action":      req.Action,
		"destination": req.Destination,
		"gateway":     req.Gateway,
	}

	// Add optional fields only if they exist
	if req.Interface != "" {
		clientPayload["interface"] = req.Interface
	}
	if req.Metric != "" {
		clientPayload["metric"] = req.Metric
	}

	return clientPayload
}

// Process route update response based on OS
func processRouteUpdateResponse(resp interface{}, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsRouteUpdateResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific route update request processing (placeholder for future differences)
func processWindowsRouteUpdateRequest(req routerUpdateRequest) map[string]string {
	// For now, return same format as Linux
	// Future: Windows might use different route command syntax
	clientPayload := map[string]string{
		"action":      req.Action,
		"destination": req.Destination,
		"gateway":     req.Gateway,
	}

	if req.Interface != "" {
		clientPayload["interface"] = req.Interface
	}
	if req.Metric != "" {
		clientPayload["metric"] = req.Metric
	}

	return clientPayload
}

// Windows-specific route update response processing (placeholder for future differences)
func processWindowsRouteUpdateResponse(resp interface{}) interface{} {
	// For now, return same format as Linux
	// Future: might need different response handling for Windows route commands
	return resp
}
