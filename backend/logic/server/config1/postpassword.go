package config1

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type PasswordChangeRequest struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClientPasswordRequest struct {
	Username string `json:"username"`
	New      string `json:"new"`
}

type ClientResponsePassword struct {
	Status string `json:"status"`
}

func HandlePasswordChange(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req PasswordChangeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.Username == "" || req.Password == "" {
			http.Error(w, "Host, username, and password are required", http.StatusBadRequest)
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
		clientPayload := ClientPasswordRequest{
			Username: req.Username,
			New:      req.Password,
		}

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
			return
		}

		// ✅ Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/config1/pass")

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
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ClientResponsePassword{
				Status: "failed",
			})
			return
		}
		defer resp.Body.Close()

		// Read client response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ClientResponsePassword{
				Status: "failed",
			})
			return
		}

		// Check if client returned error status
		if resp.StatusCode != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ClientResponsePassword{
				Status: "failed",
			})
			return
		}

		// Parse client response
		var clientResp ClientResponsePassword
		if err := json.Unmarshal(body, &clientResp); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ClientResponsePassword{
				Status: "failed",
			})
			return
		}

		// Forward the exact response from client to frontend
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clientResp)
	}
}
