package common

import (
	"backend/config"
	serverdb "backend/db/gen/server"
	"net/http"
)

func RegisterCheckRoutes(mux *http.ServeMux, queries *serverdb.Queries) {
	mux.HandleFunc("/api/server/check", config.HandleCheck(queries))
}
