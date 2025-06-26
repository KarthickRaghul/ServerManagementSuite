package common

import (
	"github.com/kishore-001/ServerManagementSuite/backend/config"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
	"github.com/kishore-001/ServerManagementSuite/backend/logic/server/config1"
	"net/http"
)

func RegisterCheckRoutes(mux *http.ServeMux, queries *serverdb.Queries) {
	mux.HandleFunc("/api/server/check", config.HandleCheck(queries))
	mux.HandleFunc("/api/server/config1/device", config1.HandleGetAllServers(queries))
}
