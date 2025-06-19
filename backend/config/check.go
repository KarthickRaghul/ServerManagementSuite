package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	serverdb "backend/db/gen/server"
)

type CheckRequest struct {
	Host string `json:"host"`
}

type CheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func HandleCheck(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Only allow POST method
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: "Method not allowed. Use POST.",
			})
			return
		}

		// Parse request body
		var req CheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: "Invalid request body. Host is required.",
			})
			return
		}

		// Get device from database
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: "Device not registered. Please register the device first.",
			})
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: "Database error occurred while checking device.",
			})
			return
		}

		// Try to connect to client
		clientURL := GetClientURL(req.Host, "/client/health")

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// Create request to client
		clientReq, err := http.NewRequest("GET", clientURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: "Failed to create request to client.",
			})
			return
		}

		// Add authorization header
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// Make request to client
		resp, err := client.Do(clientReq)
		if err != nil {
			w.WriteHeader(http.StatusOK) // Return 200 but with failed status
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: "Cannot reach client. Check if the device is online and firewall allows connections.",
			})
			return
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: fmt.Sprintf("Client returned error status: %d. Check client service status.", resp.StatusCode),
			})
			return
		}

		// Try to read response body to ensure client is responding properly
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: "Client connection unstable. Failed to read response.",
			})
			return
		}

		// Check if response is valid JSON (basic health check)
		var healthResponse map[string]interface{}
		if err := json.Unmarshal(body, &healthResponse); err != nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CheckResponse{
				Status:  "failed",
				Message: "Client returned invalid response. Service may be malfunctioning.",
			})
			return
		}

		// All checks passed
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CheckResponse{
			Status: "success",
		})
	}
}
