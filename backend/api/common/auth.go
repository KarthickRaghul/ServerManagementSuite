package common

import (
	"github.com/kishore-001/ServerManagementSuite/backend/auth"
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	"net/http"
)

func RegisterAuthRoutes(mux *http.ServeMux, queries *generaldb.Queries) {
	mux.HandleFunc("/api/auth/login", auth.HandleLogin(queries))
	mux.HandleFunc("/api/auth/refresh", auth.HandleRefresh(queries))
	mux.HandleFunc("/api/auth/verify", auth.HandleVerify(queries))

}
