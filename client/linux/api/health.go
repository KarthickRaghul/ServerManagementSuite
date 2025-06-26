package api

import (
	"github.com/kishore-001/ServerManagementSuite/linux/auth"
	"github.com/kishore-001/ServerManagementSuite/linux/logic/health"
	"net/http"
)

func RegisterHealthRoutes(mux *http.ServeMux) {
	mux.Handle(
		"/client/health",
		auth.TokenAuthMiddleware(http.HandlerFunc(health.HandleHealthConfig)),
	)
}
