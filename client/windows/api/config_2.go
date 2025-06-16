package api

import (
	"net/http"
	"windows/auth"
	"windows/logic/config_2"
)

func RegisterConfig2Routes(mux *http.ServeMux) {
	mux.Handle("/client/config-2/basic", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.NetworkConfigHandler)))
	mux.Handle("/client/config-2/route", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.RouteHandler)))
	mux.Handle("/client/config-2/firewall", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.GetWindowsFirewallRules)))
	mux.Handle("/client/config-2/updateinterface", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleUpdateInterface)))
	mux.Handle("/client/config-2/updatenetwork", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleUpdateNetworkConfig)))
	mux.Handle("/client/config-2/postrestartinterface", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleRestartInterfaces)))
}
