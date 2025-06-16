package alert

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/config"
	serverdb "backend/db/gen/server"
)

// HandleMarkSingleAlertAsSeen - Mark single alert as seen (for convenience)
func HandleMarkSingleAlertAsSeen(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization
		_, ok := config.GetUserFromContext(r)
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Get alert ID from URL path or query parameter
		alertIDStr := r.URL.Query().Get("id")
		if alertIDStr == "" {
			http.Error(w, "Alert ID is required", http.StatusBadRequest)
			return
		}

		alertID, err := strconv.ParseInt(alertIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid alert ID", http.StatusBadRequest)
			return
		}

		// Mark alert as seen
		err = queries.MarkAlertAsSeen(r.Context(), int32(alertID))
		if err != nil {
			http.Error(w, "Failed to mark alert as seen", http.StatusInternalServerError)
			return
		}

		response := AlertActionResponse{
			Status:  "success",
			Message: "Alert marked as seen",
			Count:   1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
