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
	"time"

	"backend/config"
	serverdb "backend/db/gen/server"
)

// ✅ Updated request structure to match frontend
type LogFilterRequest struct {
	Host  string `json:"host"`
	Date  string `json:"date,omitempty"`  // Format: "YYYY-MM-DD"
	Time  string `json:"time,omitempty"`  // Format: "HH:MM:SS"
	Lines int    `json:"lines,omitempty"` // Number of lines to fetch
}

// ✅ Client filter structure (without host)
type ClientLogFilter struct {
	Date string `json:"date,omitempty"`
	Time string `json:"time,omitempty"`
}

func GetLog(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LogFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
			log.Printf("❌ [LOG] Invalid request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		log.Printf("🔍 [LOG] Received request: %+v", req)

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			http.Error(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// ✅ Build client URL with query parameters
		clientURL := config.GetClientURL(req.Host, "/client/log")
		parsedURL, err := url.Parse(clientURL)
		if err != nil {
			http.Error(w, "Failed to parse client URL", http.StatusInternalServerError)
			return
		}

		// ✅ Add lines parameter to query string
		query := parsedURL.Query()
		if req.Lines > 0 {
			query.Set("lines", strconv.Itoa(req.Lines))
		} else {
			query.Set("lines", "100") // Default
		}
		parsedURL.RawQuery = query.Encode()

		log.Printf("🔍 [LOG] Client URL with query: %s", parsedURL.String())

		// ✅ Prepare request body for date/time filters (excluding host and lines)
		var requestBody io.Reader
		method := "GET"

		if req.Date != "" || req.Time != "" {
			clientFilter := ClientLogFilter{
				Date: req.Date,
				Time: req.Time,
			}

			bodyBytes, err := json.Marshal(clientFilter)
			if err != nil {
				log.Printf("❌ [LOG] Failed to marshal filter body: %v", err)
				http.Error(w, "Failed to marshal filter body", http.StatusInternalServerError)
				return
			}

			requestBody = bytes.NewReader(bodyBytes)
			method = "POST" // Use POST when sending date/time filters

			log.Printf("🔍 [LOG] Sending filter body: %s", string(bodyBytes))
		}

		// ✅ Create request to client
		clientReq, err := http.NewRequest(method, parsedURL.String(), requestBody)
		if err != nil {
			log.Printf("❌ [LOG] Failed to create client request: %v", err)
			http.Error(w, "Failed to create request to client", http.StatusInternalServerError)
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// ✅ Add timeout to prevent hanging
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		log.Printf("🔍 [LOG] Sending %s request to client: %s", method, parsedURL.String())

		resp, err := httpClient.Do(clientReq)
		if err != nil {
			log.Printf("❌ [LOG] Failed to reach client: %v", err)
			http.Error(w, "Failed to reach client", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		log.Printf("✅ [LOG] Client response status: %d", resp.StatusCode)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
