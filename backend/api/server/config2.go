package server

import (
	generaldb "backend/db/gen/server"
	"backend/logic/server/config2"
	"net/http"
)

func RegisterConfig2Routes(mux *http.ServeMux, queries *generaldb.Queries) {
	// GET-like operations via POST
	mux.HandleFunc("/api/admin/server/getfirewall", config2.HandleGetFirewall(queries))
	mux.HandleFunc("/api/admin/server/getnetworkbasics", config2.HandleGetNetworkBasics(queries))
	mux.HandleFunc("/api/admin/server/getroute", config2.HandleGetRouteTable(queries))

	// POST operations
	mux.HandleFunc("/api/admin/server/postinterface", config2.HandlePostInterface(queries))
	mux.HandleFunc("/api/admin/server/postnetwork", config2.HandlePostNetwork(queries))
	mux.HandleFunc("/api/admin/server/postrestartinterface", config2.HandlePostInterface1(queries))
	mux.HandleFunc("/api/admin/server/postupdatefirewall", config2.HandlePostUpdateFirewall(queries))
	mux.HandleFunc("/api/admin/server/postupdateroute", config2.HandlePostUpdateRouter(queries))
}
