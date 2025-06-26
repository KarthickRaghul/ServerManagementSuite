package alert

import (
	"encoding/json"
	"net/http"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

type MarkSeenRequest struct {
	AlertIDs []int32 `json:"alert_ids"`
}

// HandleMarkAlertsAsSeen - Mark alerts as seen
func HandleMarkAlertsAsSeen(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization
		_, ok := config.GetUserFromContext(r)
		if !ok {
			sendError(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Parse JSON request body
		var req MarkSeenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.AlertIDs) == 0 {
			sendError(w, "No alert IDs provided", http.StatusBadRequest)
			return
		}

		// Mark alerts as seen
		err := queries.MarkMultipleAlertsAsSeen(r.Context(), req.AlertIDs)
		if err != nil {
			sendError(w, "Failed to mark alerts as seen: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := AlertActionResponse{
			Status:  "success",
			Message: "Alerts marked as seen",
			Count:   len(req.AlertIDs),
		}

		// Send successful response
		sendGetSuccess(w, response)
	}
}
