package api

import (
	"github.com/kishore-001/ServerManagementSuite/linux/auth"
	"github.com/kishore-001/ServerManagementSuite/linux/logic/optimization"
	"net/http"
)

func RegisterOptimizeRoutes(mux *http.ServeMux) {
	mux.Handle("/client/resource/cleaninfo", auth.TokenAuthMiddleware(http.HandlerFunc(optimization.HandleFileInfo)))
	mux.Handle("/client/resource/optimize", auth.TokenAuthMiddleware(http.HandlerFunc(optimization.HandleFileClean)))
	mux.Handle("/client/resource/service", auth.TokenAuthMiddleware(http.HandlerFunc(optimization.HandleListService)))
	mux.Handle("/client/resource/restartservice", auth.TokenAuthMiddleware(http.HandlerFunc(optimization.HandleRestartService)))
}
