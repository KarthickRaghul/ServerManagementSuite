package api

import (
	"linux/auth"
	"linux/logic/config_2"
	"net/http"
)

func RegisterConfig2Routes(mux *http.ServeMux) {
	mux.Handle("/client/config2/network", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleNetworkConfig)))
	mux.Handle("/client/config2/route", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleRouteTable)))
	mux.Handle("/client/config2/firewall", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleFirewallRules)))
	mux.Handle("/client/config2/updatenetwork", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleUpdateNetworkConfig)))
	mux.Handle("/client/config2/updateinterface", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleUpdateInterface)))
	mux.Handle("/client/config2/restartinterface", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleRestartInterfaces)))
	mux.Handle("/client/config2/updateroute", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleUpdateRoute)))
	mux.Handle("/client/config2/updatefirewall", auth.TokenAuthMiddleware(http.HandlerFunc(config_2.HandleUpdateFirewall)))
}
