package api

import (
	"net/http"
	"github.com/kishore-001/ServerManagementSuite/windows/auth"
	log "github.com/kishore-001/ServerManagementSuite/windows/logic/logs"
)

func RegisterLogRoutes(mux *http.ServeMux) {
	mux.Handle("/client/log", auth.TokenAuthMiddleware(http.HandlerFunc(log.HandleLog)))
}
