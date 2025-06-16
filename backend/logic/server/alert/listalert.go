// list_alerts.go
package alert

import (
	"backend/config"
	serverdb "backend/db/gen/server"
	"encoding/json"
	"net/http"
)

func HandleListAlerts(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if _, ok := config.GetUserFromContext(r); !ok {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}
		var req struct {
			Host       string `json:"host,omitempty"`
			Limit      int    `json:"limit,omitempty"`
			OnlyUnseen bool   `json:"only_unseen,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
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
			http.Error(w, "Failed to fetch alerts", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"alerts": alerts,
			"count":  len(alerts),
		})
	}
}
