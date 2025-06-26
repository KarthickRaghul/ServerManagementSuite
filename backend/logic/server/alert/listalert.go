package alert

import (
	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
	"encoding/json"
	"net/http"
)

func HandleListAlerts(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check user context
		if _, ok := config.GetUserFromContext(r); !ok {
			sendError(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Parse request body
		var req struct {
			Host       string `json:"host,omitempty"`
			Limit      int    `json:"limit,omitempty"`
			OnlyUnseen bool   `json:"only_unseen,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Set default limit
		if req.Limit <= 0 {
			req.Limit = 100
		}

		var alerts []serverdb.Alert
		var err error

		if req.OnlyUnseen {
			if req.Host != "" {
				alerts, err = queries.GetUnseenAlertsByHost(r.Context(), req.Host)
			} else {
				alerts, err = queries.GetUnseenAlerts(r.Context(), int32(req.Limit))
			}
		} else {
			if req.Host != "" {
				alerts, err = queries.GetAlertsByHost(r.Context(), req.Host)
			} else {
				alerts, err = queries.GetAllAlerts(r.Context(), int32(req.Limit))
			}
		}

		if err != nil {
			sendError(w, "Failed to fetch alerts: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Send successful response
		response := map[string]interface{}{
			"status": "success",
			"alerts": alerts,
			"count":  len(alerts),
		}

		sendGetSuccess(w, response)
	}
}

// sendGetSuccess sends successful GET response with data
func sendGetSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
