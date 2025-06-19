package api

import (
	"net/http"
	"windows/auth"
	"windows/logic/config_1"
)

func RegisterConfig1Routes(mux *http.ServeMux) {
	mux.Handle("/client/config1/ssh", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandleSSHUpload)))
	mux.Handle("/client/config1/pass", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandlePasswordChange)))
	mux.Handle("/client/config1/basic", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandleBasicInfo)))
	mux.Handle("/client/config1/cmd", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandleCommandExec)))
	mux.Handle("/client/config1/basic_update", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandleBasicUpdate)))
	mux.Handle("/client/config1/uptime", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandleOverview)))
}
