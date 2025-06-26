package optimisation

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type hostExtract struct {
	Host string `json:"host"`
}

// Standard response structures
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func GetCleanInfo(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req hostExtract
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
			sendError(w, "Invalid request body: host is required", http.StatusBadRequest)
			return
		}

		// Lookup device and get access token
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			sendError(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/resource/cleaninfo")

		clientReq, err := http.NewRequest("GET", clientURL, nil)
		if err != nil {
			sendError(w, "Failed to create request to client: "+err.Error(), http.StatusInternalServerError)
			return
		}
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(clientReq)
		if err != nil {
			sendError(w, "Failed to reach client: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

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
		processedResp := processCleanInfoResponse(clientResp, device.Os)

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

// Process clean info response based on OS
func processCleanInfoResponse(resp interface{}, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsCleanInfoResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific clean info response processing (placeholder for future differences)
func processWindowsCleanInfoResponse(resp interface{}) interface{} {
	// For now, return same format as Linux
	// Future: Windows might have different directory structure or cleanup info
	return resp
}
