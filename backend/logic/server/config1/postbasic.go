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

type frontendRequestbasic struct {
	HostName string `json:"hostname"`
	TimeZone string `json:"timezone"`
	Host     string `json:"host"`
}

type clientResponsebasic struct {
	Status string `json:"status"`
}

func HandleBasicChange(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req frontendRequestbasic
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" || req.HostName == "" || req.TimeZone == "" {
			http.Error(w, "Host, hostname, and timezone are required", http.StatusBadRequest)
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
			"hostname": req.HostName,
			"timezone": req.TimeZone,
		}

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
			return
		}

		// Create request to remote client
		clientURL := fmt.Sprintf("http://%s/client/config1/basic_update", req.Host)
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
			// Return failed status if client unreachable
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clientResponsebasic{Status: "failed"})
			return
		}
		defer resp.Body.Close()

		// Read client response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clientResponsebasic{Status: "failed"})
			return
		}

		// Parse client response
		var clientResp clientResponsebasic
		if err := json.Unmarshal(body, &clientResp); err != nil {
			// If can't parse, assume failed
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clientResponsebasic{Status: "failed"})
			return
		}

		// Forward the exact response from client to frontend
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clientResp)
	}
}

