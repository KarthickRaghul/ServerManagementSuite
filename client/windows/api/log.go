package api

import (
	"net/http"
	"windows/auth"
	log "windows/logic/logs"
)

func RegisterLogRoutes(mux *http.ServeMux) {
	mux.Handle("/client/log", auth.TokenAuthMiddleware(http.HandlerFunc(log.HandleLog)))
}
