package api

import (
	"github.com/kishore-001/ServerManagementSuite/linux/auth"
	"github.com/kishore-001/ServerManagementSuite/linux/logic/log"
	"net/http"
)

func RegisterLogRoutes(mux *http.ServeMux) {
	mux.Handle("/client/log",
		auth.TokenAuthMiddleware(http.HandlerFunc(log.HandleLog)),
	)
}
