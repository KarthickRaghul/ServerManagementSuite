package server

import (
	serverdb "backend/db/gen/server"
	"backend/logic/server/log"
	"net/http"
)

func RegisterLogRoutes(mux *http.ServeMux, queries *serverdb.Queries) {
	// GET-like operations via POST
	mux.HandleFunc("/api/server/log", log.GetLog(queries))
}
