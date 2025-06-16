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

type FrontendRequestCmd struct {
	Command string `json:"command"`
	Host    string `json:"host"`
}

type ClientResponseCmd struct {
	Output string `json:"output"`
}

func HandleCommand(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req FrontendRequestCmd
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.Command == "" {
			http.Error(w, "Host and command are required", http.StatusBadRequest)
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
			"command": req.Command,
		}

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
			return
		}

		// Create request to remote client
		clientURL := fmt.Sprintf("http://%s/client/config1/cmd", req.Host)
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
			// Return error output if client unreachable
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ClientResponseCmd{
				Output: fmt.Sprintf("Error: Failed to reach client %s", req.Host),
			})
			return
		}
		defer resp.Body.Close()

		// Read client response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ClientResponseCmd{
				Output: "Error: Failed to read response from client",
			})
			return
		}

		// Check if client returned error status
		if resp.StatusCode != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ClientResponseCmd{
				Output: fmt.Sprintf("Error: Client returned status %d: %s", resp.StatusCode, string(body)),
			})
			return
		}

		// Parse client response
		var clientResp ClientResponseCmd
		if err := json.Unmarshal(body, &clientResp); err != nil {
			// If can't parse JSON, return raw response as output
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ClientResponseCmd{
				Output: string(body),
			})
			return
		}

		// Forward the response from client to frontend
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clientResp)
	}
}
