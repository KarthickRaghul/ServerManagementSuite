// logic/server/optimisation/postclean.go
package optimisation

import (
	"bytes"
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

type cleanRequest struct {
	Host string `json:"host"`
}

// ✅ Enhanced response structure to handle all client response types
type ClientOptimizeResponse struct {
	Status  string `json:"status"`  // "success", "partial", "failure"
	Message string `json:"message"` // Detailed message from client
}

type BackendOptimizeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func PostClean(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			log.Printf("❌ [OPTIMIZE] Invalid method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Method not allowed",
			})
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to read request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Failed to read request body",
			})
			return
		}

		var req cleanRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil || req.Host == "" {
			log.Printf("❌ [OPTIMIZE] Invalid request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Invalid request format or missing host",
			})
			return
		}

		log.Printf("🔍 [OPTIMIZE] Processing optimization request for host: %s", req.Host)

		// Verify device exists and get access token
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			log.Printf("❌ [OPTIMIZE] Device not found: %s", req.Host)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Device not registered in system",
			})
			return
		} else if err != nil {
			log.Printf("❌ [OPTIMIZE] Database error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Database error occurred",
			})
			return
		}

		// ✅ Build client URL for optimization endpoint
		clientURL := config.GetClientURL(req.Host, "/client/optimize")
		log.Printf("🔍 [OPTIMIZE] Sending request to client: %s", clientURL)

		// Create request to client
		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to create client request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Failed to create request to client",
			})
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		// ✅ Add timeout for cleanup operations
		httpClient := &http.Client{
			Timeout: 120 * time.Second, // 2 minutes for cleanup operations
		}

		// Send request to client
		resp, err := httpClient.Do(clientReq)
		if err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to reach client: %v", err)
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Failed to connect to client device",
				Details: err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		// Read client response
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to read client response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Failed to read response from client",
			})
			return
		}

		log.Printf("🔍 [OPTIMIZE] Client response status: %d", resp.StatusCode)
		log.Printf("🔍 [OPTIMIZE] Client response body: %s", string(responseBody))

		// ✅ Handle different HTTP status codes from client
		if resp.StatusCode != http.StatusOK {
			log.Printf("❌ [OPTIMIZE] Client returned error status: %d", resp.StatusCode)
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(BackendOptimizeResponse{
				Status:  "failure",
				Message: "Client device returned an error",
				Details: string(responseBody),
			})
			return
		}

		// ✅ Parse client response to handle different statuses
		var clientResp ClientOptimizeResponse
		if err := json.Unmarshal(responseBody, &clientResp); err != nil {
			log.Printf("⚠️ [OPTIMIZE] Failed to parse client response, forwarding as-is: %v", err)
			// If we can't parse, forward the raw response
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
			return
		}

		// ✅ Handle different client response statuses
		var backendResp BackendOptimizeResponse
		var httpStatus int

		switch clientResp.Status {
		case "success":
			log.Printf("✅ [OPTIMIZE] Optimization completed successfully for %s", req.Host)
			httpStatus = http.StatusOK
			backendResp = BackendOptimizeResponse{
				Status:  "success",
				Message: "System optimization completed successfully",
				Details: clientResp.Message,
			}

		case "partial":
			log.Printf("⚠️ [OPTIMIZE] Partial optimization completed for %s: %s", req.Host, clientResp.Message)
			httpStatus = http.StatusOK // Still 200 OK, but with partial status
			backendResp = BackendOptimizeResponse{
				Status:  "partial",
				Message: "System optimization partially completed - some files could not be deleted",
				Details: clientResp.Message,
			}

		case "failure":
			log.Printf("❌ [OPTIMIZE] Optimization failed for %s: %s", req.Host, clientResp.Message)
			httpStatus = http.StatusInternalServerError
			backendResp = BackendOptimizeResponse{
				Status:  "failure",
				Message: "System optimization failed",
				Details: clientResp.Message,
			}

		default:
			log.Printf("⚠️ [OPTIMIZE] Unknown status from client: %s", clientResp.Status)
			httpStatus = http.StatusOK
			backendResp = BackendOptimizeResponse{
				Status:  clientResp.Status, // Forward unknown status
				Message: clientResp.Message,
				Details: "Unknown status received from client",
			}
		}

		// Send enhanced response to frontend
		w.WriteHeader(httpStatus)
		if err := json.NewEncoder(w).Encode(backendResp); err != nil {
			log.Printf("❌ [OPTIMIZE] Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}

		log.Printf("✅ [OPTIMIZE] Response sent to frontend: %s", backendResp.Status)
	}
}
