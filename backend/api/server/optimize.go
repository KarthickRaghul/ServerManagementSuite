package server

import (
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
	"github.com/kishore-001/ServerManagementSuite/backend/logic/server/optimisation"
	"net/http"
)

func RegisterOptimisation(mux *http.ServeMux, queries *generaldb.Queries) {
	// GET-like operations via POST
	mux.HandleFunc("/api/admin/server/resource/cleaninfo", optimisation.GetCleanInfo(queries))
	mux.HandleFunc("/api/admin/server/resource/optimize", optimisation.PostClean(queries))
	mux.HandleFunc("/api/admin/server/resource/service", optimisation.GetServices(queries))
	mux.HandleFunc("/api/admin/server/resource/restartservice", optimisation.PostRestartService(queries))
}
