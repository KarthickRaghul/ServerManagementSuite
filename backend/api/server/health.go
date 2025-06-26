package server

import (
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
	"github.com/kishore-001/ServerManagementSuite/backend/logic/server/health"
	"net/http"
)

func RegisterHealthRoutes(mux *http.ServeMux, queries *serverdb.Queries) {
	// GET-like operations via POST
	mux.HandleFunc("/api/server/health", health.GetHealth(queries))
}
