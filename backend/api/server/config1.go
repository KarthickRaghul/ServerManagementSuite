package server

import (
	serverdb "backend/db/gen/server"
	"backend/logic/server/config1"
	"net/http"
)

func RegisterConfig1Routes(mux *http.ServeMux, queries *serverdb.Queries) {
	mux.HandleFunc("/api/admin/server/config1/basic", config1.HandleBasic(queries))
	mux.HandleFunc("/api/admin/server/config1/basic_update", config1.HandleBasicChange(queries))
	mux.HandleFunc("/api/admin/server/config1/create", config1.HandleCreateServer(queries))
	mux.HandleFunc("/api/admin/server/config1/delete", config1.HandleDeleteServer(queries))
	mux.HandleFunc("/api/admin/server/config1/device", config1.HandleGetAllServers(queries))
	mux.HandleFunc("/api/admin/server/config1/cmd", config1.HandleCommand(queries))
	mux.HandleFunc("/api/admin/server/config1/pass", config1.HandlePasswordChange(queries))
	mux.HandleFunc("/api/admin/server/config1/ssh", config1.HandleSSHKeyManagement(queries))
	mux.HandleFunc("/api/admin/server/config1/overview", config1.HandleServerOverview(queries))

}
