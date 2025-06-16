package server

import (
	serverdb "backend/db/gen/server"
	"backend/logic/server/health"
	"net/http"
)

func RegisterHealthRoutes(mux *http.ServeMux, queries *serverdb.Queries) {
	// GET-like operations via POST
	mux.HandleFunc("/api/server/health", health.GetHealth(queries))
}
