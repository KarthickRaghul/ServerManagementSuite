package alert

import (
	"encoding/json"
	"net/http"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type DeleteAlertsRequest struct {
	AlertIDs []int32 `json:"alert_ids"`
}

// HandleDeleteAlerts - Delete/resolve alerts
func HandleDeleteAlerts(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization
		user, ok := config.GetUserFromContext(r)
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Only admin can delete alerts
		if user.Role != "admin" {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Parse JSON request body
		var req DeleteAlertsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.AlertIDs) == 0 {
			http.Error(w, "No alert IDs provided", http.StatusBadRequest)
			return
		}

		// Delete alerts
		err := queries.DeleteMultipleAlerts(r.Context(), req.AlertIDs)
		if err != nil {
			http.Error(w, "Failed to delete alerts", http.StatusInternalServerError)
			return
		}

		response := AlertActionResponse{
			Status:  "success",
			Message: "Alerts resolved and deleted",
			Count:   len(req.AlertIDs),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
