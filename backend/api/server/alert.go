package server

import (
	serverdb "backend/db/gen/server"
	"backend/logic/server/alert"
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
