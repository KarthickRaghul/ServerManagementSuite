package server

import (
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
	"github.com/kishore-001/ServerManagementSuite/backend/logic/server/config2"
	"net/http"
)

func RegisterConfig2Routes(mux *http.ServeMux, queries *generaldb.Queries) {
	// GET-like operations via POST
	mux.HandleFunc("/api/admin/server/config2/getfirewall", config2.HandleGetFirewall(queries))
	mux.HandleFunc("/api/admin/server/config2/getnetworkbasics", config2.HandleGetNetworkBasics(queries))
	mux.HandleFunc("/api/admin/server/config2/getroute", config2.HandleGetRouteTable(queries))

	// POST operations
	mux.HandleFunc("/api/admin/server/config2/postinterface", config2.HandlePostInterface(queries))
	mux.HandleFunc("/api/admin/server/config2/postnetwork", config2.HandlePostNetwork(queries))
	mux.HandleFunc("/api/admin/server/config2/postrestartinterface", config2.HandlePostInterface1(queries))
	mux.HandleFunc("/api/admin/server/config2/postupdatefirewall", config2.HandlePostUpdateFirewall(queries))
	mux.HandleFunc("/api/admin/server/config2/postupdateroute", config2.HandlePostUpdateRouter(queries))
}
