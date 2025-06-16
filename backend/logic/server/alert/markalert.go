package alert

import (
	"encoding/json"
	"net/http"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type MarkSeenRequest struct {
	AlertIDs []int32 `json:"alert_ids"`
}

type AlertActionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// HandleMarkAlertsAsSeen - Mark alerts as seen
func HandleMarkAlertsAsSeen(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization
		_, ok := config.GetUserFromContext(r)
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Parse JSON request body
		var req MarkSeenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.AlertIDs) == 0 {
			http.Error(w, "No alert IDs provided", http.StatusBadRequest)
			return
		}

		// Mark alerts as seen
		err := queries.MarkMultipleAlertsAsSeen(r.Context(), req.AlertIDs)
		if err != nil {
			http.Error(w, "Failed to mark alerts as seen", http.StatusInternalServerError)
			return
		}

		response := AlertActionResponse{
			Status:  "success",
			Message: "Alerts marked as seen",
			Count:   len(req.AlertIDs),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
