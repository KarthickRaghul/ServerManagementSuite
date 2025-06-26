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

type FrontendRequestCmd struct {
	Command string `json:"command"`
	Host    string `json:"host"`
}

type ClientResponseCmd struct {
	Status string `json:"status"`
	Output string `json:"output"`
}

func HandleCommand(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req FrontendRequestCmd
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.Command == "" {
			sendError(w, "Host and command are required", http.StatusBadRequest)
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

		// Process command based on OS
		clientPayload := processCommandRequest(req, device.Os)

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			sendError(w, "Failed to prepare request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/config1/cmd")

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

		// Read client response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			sendError(w, "Failed to read client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Handle client error responses
		if resp.StatusCode != http.StatusOK {
			// Try to parse client error response (standardized format)
			var clientError ErrorResponse
			if json.Unmarshal(body, &clientError) == nil && clientError.Status == "failed" {
				sendError(w, "Client error: "+clientError.Message, http.StatusBadGateway)
			} else {
				sendError(w, "Client error: "+string(body), http.StatusBadGateway)
			}
			return
		}

		// Parse client response (expecting {status: "success", output: "command output"})
		var clientResp ClientResponseCmd
		if err := json.Unmarshal(body, &clientResp); err != nil {
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processCommandResponse(clientResp, device.Os)

		// Send successful response (special case: includes output field)
		sendGetSuccess(w, processedResp)
	}
}

// Process command request based on OS
func processCommandRequest(req FrontendRequestCmd, osType string) map[string]string {
	if strings.ToLower(osType) == "windows" {
		return processWindowsCommandRequest(req)
	}

	// Default Linux behavior
	return map[string]string{
		"command": strings.TrimSpace(req.Command),
	}
}

// Process command response based on OS
func processCommandResponse(resp ClientResponseCmd, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsCommandResponse(resp)
	}

	// Default Linux behavior - return the special format with status and output
	return ClientResponseCmd{
		Status: resp.Status,
		Output: resp.Output,
	}
}

// Windows-specific command request processing (placeholder for future differences)
func processWindowsCommandRequest(req FrontendRequestCmd) map[string]string {
	// For now, return same format as Linux
	// Future: might need to translate Linux commands to Windows equivalents
	return map[string]string{
		"command": strings.TrimSpace(req.Command),
	}
}

// Windows-specific command response processing (placeholder for future differences)
func processWindowsCommandResponse(resp ClientResponseCmd) interface{} {
	// For now, return same format as Linux
	// Future: might need to format Windows command output differently
	return ClientResponseCmd{
		Status: resp.Status,
		Output: resp.Output,
	}
}
