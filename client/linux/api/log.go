package api

import (
	"linux/auth"
	"linux/logic/log"
	"net/http"
)

func RegisterLogRoutes(mux *http.ServeMux) {
	mux.Handle("/client/log",
		auth.TokenAuthMiddleware(http.HandlerFunc(log.HandleLog)),
	)
}
