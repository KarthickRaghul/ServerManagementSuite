package alert

import (
	"encoding/json"
	"net/http"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

type DeleteAlertsRequest struct {
	AlertIDs []int32 `json:"alert_ids"`
}

type AlertActionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// Standard response structures
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HandleDeleteAlerts - Delete/resolve alerts
func HandleDeleteAlerts(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow DELETE
		if r.Method != http.MethodDelete {
			sendError(w, "Only DELETE method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user authorization
		user, ok := config.GetUserFromContext(r)
		if !ok {
			sendError(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Only admin can delete alerts
		if user.Role != "admin" {
			sendError(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Parse JSON request body
		var req DeleteAlertsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.AlertIDs) == 0 {
			sendError(w, "No alert IDs provided", http.StatusBadRequest)
			return
		}

		// Delete alerts
		err := queries.DeleteMultipleAlerts(r.Context(), req.AlertIDs)
		if err != nil {
			sendError(w, "Failed to delete alerts: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := AlertActionResponse{
			Status:  "success",
			Message: "Alerts resolved and deleted",
			Count:   len(req.AlertIDs),
		}

		// Send successful response
		sendDeleteSuccess(w, response)
	}
}

// Standard response functions
func sendDeleteSuccess(w http.ResponseWriter, data interface{}) {
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
