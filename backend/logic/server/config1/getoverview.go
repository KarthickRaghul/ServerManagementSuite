package config1

import (
	"backend/config"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

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
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization
		_, ok := config.GetUserFromContext(r)
		if !ok {
			sendError(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Parse request body
		var req ServerOverviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			sendError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Host == "" {
			sendError(w, "Host is required", http.StatusBadRequest)
			return
		}

		// Lookup device and get access token
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			log.Printf("Host %s not found in database", req.Host)
			// Host not registered, return offline status
			response := ServerOverviewResponse{
				Status: "offline",
				Uptime: "N/A",
			}
			sendGetSuccess(w, response)
			return
		} else if err != nil {
			log.Printf("Database error for host %s: %v", req.Host, err)
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get uptime from client
		uptime, isOnline := getClientUptime(req.Host, device.AccessToken)

		// Process response based on OS and availability
		response := processOverviewResponse(uptime, isOnline, device.Os)

		// Send successful response
		sendGetSuccess(w, response)
	}
}

// getClientUptime fetches uptime from the client and returns uptime string and online status
func getClientUptime(host, accessToken string) (string, bool) {
	clientURL := config.GetClientURL(host, "/client/config1/uptime")

	req, err := http.NewRequest("GET", clientURL, nil)
	if err != nil {
		log.Printf("Failed to create request for %s: %v", host, err)
		return "", false
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Network error connecting to %s: %v", host, err)
		return "", false
	}
	defer resp.Body.Close()

	// Handle client error responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		// Try to parse client error response (standardized format)
		var clientError ErrorResponse
		if json.Unmarshal(body, &clientError) == nil && clientError.Status == "failed" {
			log.Printf("Client %s error: %s", host, clientError.Message)
		} else {
			log.Printf("Client %s returned status %d: %s", host, resp.StatusCode, string(body))
		}
		return "", false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response from %s: %v", host, err)
		return "", false
	}

	var clientResp ClientUptimeResponse
	if err := json.Unmarshal(body, &clientResp); err != nil {
		log.Printf("Invalid JSON response from %s: %v", host, err)
		return "", false
	}

	if clientResp.Uptime == "" {
		log.Printf("Empty uptime received from %s", host)
		return "Unknown", true
	}

	return clientResp.Uptime, true
}

// Process overview response based on OS
func processOverviewResponse(uptime string, isOnline bool, osType string) ServerOverviewResponse {
	if !isOnline {
		return ServerOverviewResponse{
			Status: "offline",
			Uptime: "N/A",
		}
	}

	// Process uptime based on OS
	processedUptime := processUptimeByOS(uptime, osType)

	return ServerOverviewResponse{
		Status: "online",
		Uptime: processedUptime,
	}
}

// Process uptime format based on OS
func processUptimeByOS(uptime, osType string) string {
	if strings.ToLower(osType) == "windows" {
		return processWindowsUptime(uptime)
	}

	// Default Linux behavior
	return strings.TrimSpace(uptime)
}

// Windows-specific uptime processing (placeholder for future differences)
func processWindowsUptime(uptime string) string {
	// For now, return same format as Linux
	// Add Windows-specific formatting when needed
	return strings.TrimSpace(uptime)
}
