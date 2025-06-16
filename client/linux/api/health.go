package api

import (
	"linux/auth"
	"linux/logic/health"
	"net/http"
)

func RegisterHealthRoutes(mux *http.ServeMux) {
	mux.Handle(
		"/client/health",
		auth.TokenAuthMiddleware(http.HandlerFunc(health.HandleHealthConfig)),
	)
}
