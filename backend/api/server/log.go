package server

import (
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
	"github.com/kishore-001/ServerManagementSuite/backend/logic/server/log"
	"net/http"
)

func RegisterLogRoutes(mux *http.ServeMux, queries *serverdb.Queries) {
	// GET-like operations via POST
	mux.HandleFunc("/api/server/log", log.GetLog(queries))
}
