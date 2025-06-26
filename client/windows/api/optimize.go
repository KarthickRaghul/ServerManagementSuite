package api

import (
	"net/http"
	"github.com/kishore-001/ServerManagementSuite/windows/auth"
	"github.com/kishore-001/ServerManagementSuite/windows/logic/optimization"
)

func RegisterOptimizeRoutes(mux *http.ServeMux) {
	mux.Handle("/client/cleaninfo", auth.TokenAuthMiddleware(http.HandlerFunc(optimization.HandleFileInfo)))
	mux.Handle("/client/optimize", auth.TokenAuthMiddleware(http.HandlerFunc(optimization.HandleFileClean)))
	mux.Handle("/client/service", auth.TokenAuthMiddleware(http.HandlerFunc(optimization.HandleListService)))
	mux.Handle("/client/restartservice", auth.TokenAuthMiddleware(http.HandlerFunc(optimization.HandleRestartService)))
}
