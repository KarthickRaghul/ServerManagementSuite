package config2

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type hostExtract6 struct {
	Host string `json:"host"`
}

func HandleGetNetworkBasics(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		var req hostExtract6
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
			sendError(w, "Invalid request body: host is required", http.StatusBadRequest)
			return
		}

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			sendError(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/config2/network")

		clientReq, err := http.NewRequest("GET", clientURL, nil)
		if err != nil {
			sendError(w, "Failed to create request to client: "+err.Error(), http.StatusInternalServerError)
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

			// Try to parse client error response (standardized format)
			var clientError ErrorResponse
			if json.Unmarshal(body, &clientError) == nil && clientError.Status == "failed" {
				sendError(w, "Client error: "+clientError.Message, http.StatusBadGateway)
			} else {
				sendError(w, "Client error: "+string(body), http.StatusBadGateway)
			}
			return
		}

		// Parse client response
		var clientResp interface{}
		if err := json.NewDecoder(resp.Body).Decode(&clientResp); err != nil {
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processNetworkBasicsResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
	}
}

// Process network basics response based on OS
func processNetworkBasicsResponse(resp interface{}, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsNetworkBasicsResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific network basics response processing (placeholder for future differences)
func processWindowsNetworkBasicsResponse(resp interface{}) interface{} {
	// For now, return same format as Linux
	// Future: Windows network info structure differs from Linux (ipconfig vs ip command)
	return resp
}
