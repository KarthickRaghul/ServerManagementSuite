package common

import (
	"backend/config"
	serverdb "backend/db/gen/server"
	"backend/logic/server/config1"
	"net/http"
)

func RegisterCheckRoutes(mux *http.ServeMux, queries *serverdb.Queries) {
	mux.HandleFunc("/api/server/check", config.HandleCheck(queries))
	mux.HandleFunc("/api/server/config1/device", config1.HandleGetAllServers(queries))
}
