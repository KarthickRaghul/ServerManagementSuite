package config1

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	serverdb "backend/db/gen/server"
)

// Request body from frontend
type basicRequest struct {
	Host string `json:"host"`
}

// Response from client (remote server)
type clientBasicResponse struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
}

// Handler for /api/server/basic
func HandleBasic(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse frontend request
		var req basicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Lookup device and get access token
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			http.Error(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Prepare request to remote client
		clientURL := fmt.Sprintf("http://%s/client/config1/basic", req.Host)
		clientReq, err := http.NewRequest("GET", clientURL, nil)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// Send request to client
		httpClient := &http.Client{}
		resp, err := httpClient.Do(clientReq)
		if err != nil {
			http.Error(w, "Failed to reach client", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			http.Error(w, fmt.Sprintf("Client error: %s", string(body)), resp.StatusCode)
			return
		}

		// Parse client response
		var clientResp clientBasicResponse
		if err := json.NewDecoder(resp.Body).Decode(&clientResp); err != nil {
			http.Error(w, "Invalid client response", http.StatusBadGateway)
			return
		}

		// Return only hostname and timezone to frontend
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clientBasicResponse{
			Hostname: clientResp.Hostname,
			Timezone: clientResp.Timezone,
		})
	}
}
