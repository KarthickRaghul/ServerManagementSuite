package server

import (
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
	"github.com/kishore-001/ServerManagementSuite/backend/logic/server/alert"
	"net/http"
)

// Register alert routes
func RegisterAlertRoutes(mux *http.ServeMux, queries *serverdb.Queries) {
	// Existing route
	mux.HandleFunc("/api/server/alerts", alert.HandleListAlerts(queries))

	// New routes for status management
	mux.HandleFunc("/api/server/alerts/markseen", alert.HandleMarkAlertsAsSeen(queries))
	mux.HandleFunc("/api/server/alerts/marksingleseen", alert.HandleMarkSingleAlertAsSeen(queries))
	mux.HandleFunc("/api/server/alerts/delete", alert.HandleDeleteAlerts(queries))
}
