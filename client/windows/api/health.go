package api

import (
	"net/http"
	"windows/auth"
	"windows/logic/health"
)

func RegisterHealthRoutes(mux *http.ServeMux) {
	mux.Handle("/client/health", auth.TokenAuthMiddleware(http.HandlerFunc(health.HandleHealthConfig)))
}
