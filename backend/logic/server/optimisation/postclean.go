package optimisation

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type cleanRequest struct {
	Host string `json:"host"`
}

// Enhanced response structure to handle all client response types
type ClientOptimizeResponse struct {
	Status  string `json:"status"`  // "success", "partial", "failed"
	Message string `json:"message"` // Detailed message from client
}

func PostClean(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			log.Printf("❌ [OPTIMIZE] Invalid method: %s", r.Method)
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req cleanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
			log.Printf("❌ [OPTIMIZE] Invalid request: %v", err)
			sendError(w, "Invalid request format or missing host", http.StatusBadRequest)
			return
		}

		log.Printf("🔍 [OPTIMIZE] Processing optimization request for host: %s", req.Host)

		// Lookup device and get access token
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			log.Printf("❌ [OPTIMIZE] Device not found: %s", req.Host)
			sendError(w, "Device not registered in system", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("❌ [OPTIMIZE] Database error: %v", err)
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Process clean request based on OS
		clientPayload := processCleanRequest(req, device.Os)

		jsonPayload, err := json.Marshal(clientPayload)
		if err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to marshal client payload: %v", err)
			sendError(w, "Failed to prepare client request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Build client URL for optimization endpoint
		clientURL := config.GetClientURL(req.Host, "/client/resource/optimize")
		log.Printf("🔍 [OPTIMIZE] Sending request to client: %s", clientURL)

		// Create request to client
		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to create client request: %v", err)
			sendError(w, "Failed to create request to client: "+err.Error(), http.StatusInternalServerError)
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// Add timeout for cleanup operations
		httpClient := &http.Client{
			Timeout: 120 * time.Second, // 2 minutes for cleanup operations
		}

		// Send request to client
		resp, err := httpClient.Do(clientReq)
		if err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to reach client: %v", err)
			sendError(w, "Failed to connect to client device: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		log.Printf("🔍 [OPTIMIZE] Client response status: %d", resp.StatusCode)

		// Handle client error responses
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("❌ [OPTIMIZE] Client returned error status: %d", resp.StatusCode)

			var clientError ErrorResponse
			if json.Unmarshal(body, &clientError) == nil && clientError.Status == "failed" {
				sendError(w, "Client error: "+clientError.Message, http.StatusBadGateway)
			} else {
				sendError(w, "Client device returned an error: "+string(body), http.StatusBadGateway)
			}
			return
		}

		// Read client response
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to read client response: %v", err)
			sendError(w, "Failed to read response from client: "+err.Error(), http.StatusBadGateway)
			return
		}

		log.Printf("🔍 [OPTIMIZE] Client response body: %s", string(responseBody))

		// Parse client response to handle different statuses
		var clientResp ClientOptimizeResponse
		if err := json.Unmarshal(responseBody, &clientResp); err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to parse client response: %v", err)
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processCleanResponse(clientResp, device.Os)

		// Handle different client response statuses (SPECIAL CASE: supports "partial")
		switch processedResp.Status {
		case "success":
			log.Printf("✅ [OPTIMIZE] Optimization completed successfully for %s", req.Host)
			sendGetSuccess(w, processedResp)

		case "partial":
			log.Printf("⚠️ [OPTIMIZE] Partial optimization completed for %s: %s", req.Host, processedResp.Message)
			// Special case: return partial status directly
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(processedResp)

		case "failed":
			log.Printf("❌ [OPTIMIZE] Optimization failed for %s: %s", req.Host, processedResp.Message)
			sendError(w, "System optimization failed: "+processedResp.Message, http.StatusInternalServerError)

		default:
			log.Printf("⚠️ [OPTIMIZE] Unknown status from client: %s", processedResp.Status)
			sendGetSuccess(w, processedResp) // Forward unknown status
		}

		log.Printf("✅ [OPTIMIZE] Response sent to frontend: %s", processedResp.Status)
	}
}

// Process clean request based on OS
func processCleanRequest(req cleanRequest, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsCleanRequest(req)
	}

	// Default Linux behavior - no host field needed for client
	return map[string]interface{}{}
}

// Process clean response based on OS
func processCleanResponse(resp ClientOptimizeResponse, osType string) ClientOptimizeResponse {
	if strings.ToLower(osType) == "windows" {
		return processWindowsCleanResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific clean request processing (placeholder for future differences)
func processWindowsCleanRequest(req cleanRequest) interface{} {
	// For now, return same format as Linux
	// Future: Windows might need different cleanup parameters
	return map[string]interface{}{}
}

// Windows-specific clean response processing (placeholder for future differences)
func processWindowsCleanResponse(resp ClientOptimizeResponse) ClientOptimizeResponse {
	// For now, return same format as Linux
	// Future: Windows might have different cleanup result structure
	return resp
}
