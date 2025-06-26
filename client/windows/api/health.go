package api

import (
	"net/http"
	"github.com/kishore-001/ServerManagementSuite/windows/auth"
	"github.com/kishore-001/ServerManagementSuite/windows/logic/health"
)

func RegisterHealthRoutes(mux *http.ServeMux) {
	mux.Handle("/client/health", auth.TokenAuthMiddleware(http.HandlerFunc(health.HandleHealthConfig)))
}
