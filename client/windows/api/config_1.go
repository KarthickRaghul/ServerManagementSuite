package api

import (
	"net/http"
	"windows/auth"
	"windows/logic/config_1"
)

func RegisterConfig1Routes(mux *http.ServeMux) {
	mux.Handle("/client/config-1/ssh", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandleSSHUpload)))
	mux.Handle("/client/config-1/pass", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandlePasswordChange)))
	mux.Handle("/client/config-1/basic", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandleBasicInfo)))
	mux.Handle("/client/config-1/cmd", auth.TokenAuthMiddleware(http.HandlerFunc(config_1.HandleCommandExec)))
}
