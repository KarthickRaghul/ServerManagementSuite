package common

import (
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	"github.com/kishore-001/ServerManagementSuite/backend/logic/settings"
	"net/http"
)

func RegisterSettingsRoutes(mux *http.ServeMux, queries *generaldb.Queries) {
	mux.HandleFunc("/api/admin/settings/adduser", settings.HandleAddUser(queries))
	mux.HandleFunc("/api/admin/settings/removeuser", settings.HandleRemoveUser(queries))
	mux.HandleFunc("/api/admin/settings/listuser", settings.HandleListUsers(queries))

}
