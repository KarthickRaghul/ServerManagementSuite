package log

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"backend/config"
	serverdb "backend/db/gen/server"
)

// Updated request structure to match frontend
type LogFilterRequest struct {
	Host  string `json:"host"`
	Date  string `json:"date,omitempty"`  // Format: "YYYY-MM-DD"
	Time  string `json:"time,omitempty"`  // Format: "HH:MM:SS"
	Lines int    `json:"lines,omitempty"` // Number of lines to fetch
}

// Client filter structure (without host)
type ClientLogFilter struct {
	Date string `json:"date,omitempty"`
	Time string `json:"time,omitempty"`
}

// Standard response structures
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func GetLog(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LogFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
			log.Printf("❌ [LOG] Invalid request body: %v", err)
			sendError(w, "Invalid request body: host is required", http.StatusBadRequest)
			return
		}

		log.Printf("🔍 [LOG] Received request: %+v", req)

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			sendError(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Build client URL with query parameters
		clientURL := config.GetClientURL(req.Host, "/client/log")
		parsedURL, err := url.Parse(clientURL)
		if err != nil {
			sendError(w, "Failed to parse client URL: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Add lines parameter to query string
		query := parsedURL.Query()
		if req.Lines > 0 {
			query.Set("lines", strconv.Itoa(req.Lines))
		} else {
			query.Set("lines", "100") // Default
		}
		parsedURL.RawQuery = query.Encode()

		log.Printf("🔍 [LOG] Client URL with query: %s", parsedURL.String())

		// Process log request based on OS
		requestBody, method := processLogRequest(req, device.Os)

		log.Printf("🔍 [LOG] Using method: %s", method)

		// Create request to client
		clientReq, err := http.NewRequest(method, parsedURL.String(), requestBody)
		if err != nil {
			log.Printf("❌ [LOG] Failed to create client request: %v", err)
			sendError(w, "Failed to create request to client: "+err.Error(), http.StatusInternalServerError)
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// Add timeout to prevent hanging
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		log.Printf("🔍 [LOG] Sending %s request to client: %s", method, parsedURL.String())

		resp, err := httpClient.Do(clientReq)
		if err != nil {
			log.Printf("❌ [LOG] Failed to reach client: %v", err)
			sendError(w, "Failed to reach client: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		log.Printf("✅ [LOG] Client response status: %d", resp.StatusCode)

		// Handle client error responses
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)

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
		processedResp := processLogResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
	}
}

// Standard response functions
func sendGetSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResp := ErrorResponse{
		Status:  "failed",
		Message: message,
	}
	json.NewEncoder(w).Encode(errorResp)
}

// Process log request based on OS
func processLogRequest(req LogFilterRequest, osType string) (io.Reader, string) {
	if strings.ToLower(osType) == "windows" {
		return processWindowsLogRequest(req)
	}

	// Default Linux behavior
	return processLinuxLogRequest(req)
}

// Process Linux log request
func processLinuxLogRequest(req LogFilterRequest) (io.Reader, string) {
	// Prepare request body for date/time filters (excluding host and lines)
	method := "GET"

	if req.Date != "" || req.Time != "" {
		clientFilter := ClientLogFilter{
			Date: req.Date,
			Time: req.Time,
		}

		bodyBytes, _ := json.Marshal(clientFilter)
		method = "POST" // Use POST when sending date/time filters

		log.Printf("🔍 [LOG] Sending Linux filter body: %s", string(bodyBytes))
		return bytes.NewReader(bodyBytes), method
	}

	return nil, method
}

// Process Windows log request (placeholder for future differences)
func processWindowsLogRequest(req LogFilterRequest) (io.Reader, string) {
	// For now, return same format as Linux
	// Future: Windows might use different log filtering (Event Viewer vs journalctl)
	return processLinuxLogRequest(req)
}

// Process log response based on OS
func processLogResponse(resp interface{}, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsLogResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific log response processing (placeholder for future differences)
func processWindowsLogResponse(resp interface{}) interface{} {
	// For now, return same format as Linux
	// Future: Windows Event Viewer logs might have different structure than journalctl
	return resp
}
