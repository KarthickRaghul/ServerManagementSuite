package config2

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type routerUpdateRequest struct {
	Host        string `json:"host"`
	Action      string `json:"action"`      // "add" or "delete"
	Destination string `json:"destination"` // e.g., "192.168.2.0/24"
	Gateway     string `json:"gateway"`     // e.g., "192.168.1.1"
	Interface   string `json:"interface"`   // optional
	Metric      string `json:"metric"`      // optional
}

type responseJSON3 struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
}

func HandlePostUpdateRouter(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(responseJSON3{
				Status: "failure",
				Error:  "Method not allowed",
			})
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("❌ [ROUTE] Failed to read request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON3{
				Status: "failure",
				Error:  "Failed to read request body",
			})
			return
		}

		var req routerUpdateRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil ||
			req.Host == "" || req.Action == "" || req.Destination == "" || req.Gateway == "" {
			log.Printf("❌ [ROUTE] Invalid request data: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON3{
				Status: "failure",
				Error:  "Invalid request data - host, action, destination, and gateway are required",
			})
			return
		}

		log.Printf("🔍 [ROUTE] Processing route %s for host: %s, destination: %s", req.Action, req.Host, req.Destination)

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			log.Printf("❌ [ROUTE] Device not found: %s", req.Host)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(responseJSON3{
				Status: "failure",
				Error:  "Device not registered",
			})
			return
		} else if err != nil {
			log.Printf("❌ [ROUTE] Database error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON3{
				Status: "failure",
				Error:  "Database error",
			})
			return
		}

		// ✅ Create client payload WITHOUT the host field
		clientPayload := map[string]string{
			"action":      req.Action,
			"destination": req.Destination,
			"gateway":     req.Gateway,
		}

		// Add optional fields only if they exist
		if req.Interface != "" {
			clientPayload["interface"] = req.Interface
		}
		if req.Metric != "" {
			clientPayload["metric"] = req.Metric
		}

		// Convert to JSON for client
		clientBody, err := json.Marshal(clientPayload)
		if err != nil {
			log.Printf("❌ [ROUTE] Failed to marshal client payload: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON3{
				Status: "failure",
				Error:  "Failed to prepare client request",
			})
			return
		}

		log.Printf("🔍 [ROUTE] Client payload: %s", string(clientBody))

		// Use config for client URL
		clientURL := config.GetClientURL(req.Host, "/client/config2/updateroute")
		log.Printf("🔍 [ROUTE] Sending request to: %s", clientURL)

		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(clientBody))
		if err != nil {
			log.Printf("❌ [ROUTE] Failed to create client request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON3{
				Status: "failure",
				Error:  "Failed to create client request",
			})
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// Add timeout to prevent hanging
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := httpClient.Do(clientReq)
		if err != nil {
			log.Printf("❌ [ROUTE] Failed to reach client: %v", err)
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(responseJSON3{
				Status:  "failure",
				Error:   "Failed to reach client",
				Details: err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		log.Printf("🔍 [ROUTE] Client response status: %d", resp.StatusCode)

		// Read response body for debugging
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("❌ [ROUTE] Failed to read response body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON3{
				Status: "failure",
				Error:  "Failed to read client response",
			})
			return
		}

		log.Printf("🔍 [ROUTE] Client response body: %s", string(body))

		if resp.StatusCode != http.StatusOK {
			log.Printf("❌ [ROUTE] Client returned non-200 status: %d", resp.StatusCode)
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(responseJSON3{
				Status:  "failure",
				Error:   fmt.Sprintf("Client returned status %d", resp.StatusCode),
				Details: string(body),
			})
			return
		}

		// Success - forward client response
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		log.Printf("✅ [ROUTE] Route %s completed successfully for %s", req.Action, req.Host)
	}
}
