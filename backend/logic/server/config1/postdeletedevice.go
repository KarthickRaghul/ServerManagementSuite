package config1

import (
	"encoding/json"
	"net/http"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

func HandleDeleteServer(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a DELETE request
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check admin authorization
		user, _ := config.GetUserFromContext(r)

		var req struct {
			IP string `json:"ip"`
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

		// Check if device exists before deletion
		_, err := queries.GetServerDeviceByIP(r.Context(), req.IP)
		if err != nil {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		// Delete the device
		err = queries.DeleteServerDevice(r.Context(), req.IP)
		if err != nil {
			http.Error(w, "Failed to delete device", http.StatusInternalServerError)
			return
		}

		// Success response
		response := map[string]interface{}{
			"status":     "success",
			"message":    "Server device deleted successfully",
			"deleted_ip": req.IP,
			"deleted_by": user.Username,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
