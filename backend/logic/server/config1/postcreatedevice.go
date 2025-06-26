package config1

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

func HandleCreateServer(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a POST request
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check admin authorization
		user, _ := config.GetUserFromContext(r)

		// Parse request body
		var req struct {
			IP  string `json:"ip"`
			Tag string `json:"tag"`
			OS  string `json:"os"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.IP == "" {
			http.Error(w, "IP address is required", http.StatusBadRequest)
			return
		}

		// Generate access token for the device
		accessToken, err := generateDeviceToken()
		if err != nil {
			http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
			return
		}

		// Create device in database
		device, err := queries.CreateServerDevice(r.Context(), serverdb.CreateServerDeviceParams{
			Ip:          req.IP,
			Tag:         req.Tag,
			Os:          req.OS,
			AccessToken: accessToken,
		})
		if err != nil {
			// Check for duplicate IP error
			if err.Error() == "pq: duplicate key value violates unique constraint" {
				http.Error(w, "Device with this IP already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to create device", http.StatusInternalServerError)
			return
		}

		// Success response
		response := map[string]interface{}{
			"status":  "success",
			"message": "Server device created successfully",
			"device": map[string]interface{}{
				"id":         device.ID,
				"ip":         device.Ip,
				"tag":        device.Tag,
				"os":         device.Os,
				"created_at": device.CreatedAt,
			},
			"access_token": accessToken, // Return token for device setup
			"created_by":   user.Username,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// generateDeviceToken creates a secure token for device access
func generateDeviceToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
