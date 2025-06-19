package config1

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type ServerOverviewRequest struct {
	Host string `json:"host"`
}

type ClientUptimeResponse struct {
	Uptime string `json:"uptime"`
}

type ServerOverviewResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

func HandleServerOverview(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a POST request
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization
		_, ok := config.GetUserFromContext(r)
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Parse request body
		var req ServerOverviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" {
			http.Error(w, "Host is required", http.StatusBadRequest)
			return
		}

		// Get device access token from database
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			log.Printf("Host %s not found in database", req.Host)
			// Host not registered, return offline status
			response := ServerOverviewResponse{
				Status: "offline",
				Uptime: "N/A",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		} else if err != nil {
			log.Printf("Database error for host %s: %v", req.Host, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Try to get uptime from client
		uptime, isOnline := getClientUptime(req.Host, device.AccessToken)

		// Prepare response based on client availability
		var response ServerOverviewResponse
		if isOnline {
			response = ServerOverviewResponse{
				Status: "online",
				Uptime: uptime,
			}
		} else {
			response = ServerOverviewResponse{
				Status: "offline",
				Uptime: "N/A",
			}
			log.Printf("Host %s is offline", req.Host)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// getClientUptime fetches uptime from the client and returns uptime string and online status
func getClientUptime(host, accessToken string) (string, bool) {
	// ✅ Use config for client URL (reads from .env file)
	clientURL := config.GetClientURL(host, "/client/config1/uptime")

	req, err := http.NewRequest("GET", clientURL, nil)
	if err != nil {
		log.Printf("Failed to create request for %s: %v", host, err)
		return "", false
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second, // 10 second timeout
	}

	// Send request to client
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Network error connecting to %s: %v", host, err)
		return "", false
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Client %s returned status %d", host, resp.StatusCode)
		return "", false
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response from %s: %v", host, err)
		return "", false
	}

	// Parse client response
	var clientResp ClientUptimeResponse
	if err := json.Unmarshal(body, &clientResp); err != nil {
		log.Printf("Invalid JSON response from %s: %v", host, err)
		return "", false
	}

	// Validate uptime is not empty
	if clientResp.Uptime == "" {
		log.Printf("Empty uptime received from %s", host)
		return "Unknown", true // Still consider online but uptime unknown
	}

	// Client is online and responded with uptime
	return clientResp.Uptime, true
}
