package config1

import (
	"encoding/json"
	"net/http"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

func HandleGetAllServers(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a GET request
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization (both admin and viewer can view devices)
		user, ok := config.GetUserFromContext(r)
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Get all devices from database
		devices, err := queries.GetAllServerDevices(r.Context())
		if err != nil {
			http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
			return
		}

		// Prepare response (exclude access tokens for security)
		var deviceList []map[string]interface{}
		for _, device := range devices {
			deviceList = append(deviceList, map[string]interface{}{
				"id":         device.ID,
				"ip":         device.Ip,
				"tag":        device.Tag,
				"os":         device.Os, // Handle sql.NullString
				"created_at": device.CreatedAt,
			})
		}

		// Success response
		response := map[string]interface{}{
			"status":  "success",
			"message": "Devices retrieved successfully",
			"devices": deviceList,
			"count":   len(deviceList),
			"user":    user.Username,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
