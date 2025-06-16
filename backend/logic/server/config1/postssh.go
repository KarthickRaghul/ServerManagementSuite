package config1

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req SSHKeyManagementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.Key == "" {
			http.Error(w, "Host and key are required", http.StatusBadRequest)
			return
		}

		// Get device access token from database
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			http.Error(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Prepare payload for client
		clientPayload := map[string]string{
			"key": req.Key,
		}

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
			return
		}

		// Create request to remote client
		clientURL := fmt.Sprintf("http://%s/client/config1/ssh", req.Host)
		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
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
			// Return error response if client unreachable
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SSHKeyManagementResponse{
				Status:  "failed",
				Message: "Failed to reach client",
			})
			return
		}
		defer resp.Body.Close()

		// Read client response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SSHKeyManagementResponse{
				Status:  "failed",
				Message: "Failed to read response",
			})
			return
		}

		// Check if client returned error status
		if resp.StatusCode != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SSHKeyManagementResponse{
				Status:  "failed",
				Message: fmt.Sprintf("Client error: %s", string(body)),
			})
			return
		}

		// Parse client response
		var clientResp SSHKeyManagementResponse
		if err := json.Unmarshal(body, &clientResp); err != nil {
			// If can't parse, return raw response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SSHKeyManagementResponse{
				Status:  "failed",
				Message: string(body),
			})
			return
		}

		// Forward the exact response from client to frontend
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clientResp)
	}
}
