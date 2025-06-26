package alert

import (
	"net/http"
	"strconv"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

// HandleMarkSingleAlertAsSeen - Mark single alert as seen (for convenience)
func HandleMarkSingleAlertAsSeen(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow PUT
		if r.Method != http.MethodPut {
			sendError(w, "Only PUT method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization
		_, ok := config.GetUserFromContext(r)
		if !ok {
			sendError(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Get alert ID from URL query parameter
		alertIDStr := r.URL.Query().Get("id")
		if alertIDStr == "" {
			sendError(w, "Alert ID is required", http.StatusBadRequest)
			return
		}

		alertID, err := strconv.ParseInt(alertIDStr, 10, 32)
		if err != nil {
			sendError(w, "Invalid alert ID: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Mark alert as seen
		err = queries.MarkAlertAsSeen(r.Context(), int32(alertID))
		if err != nil {
			sendError(w, "Failed to mark alert as seen: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := AlertActionResponse{
			Status:  "success",
			Message: "Alert marked as seen",
			Count:   1,
		}

		// Send successful response
		sendGetSuccess(w, response)
	}
}

