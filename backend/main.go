// main.go
package main

import (
	"backend/api/common"
	"backend/api/server"
	"backend/config"
	"backend/routine"
	"log"
	"net/http"
)

func main() {
	// Load configuration first
	config.LoadConfig()

	// Create separate muxes for different protection levels
	publicMux := http.NewServeMux()
	protectedMux := http.NewServeMux()
	adminMux := http.NewServeMux()

	// This is necessary for Database
	generalqueries := config.GeneralQueries()
	serverqueries := config.ServerQueries()

	// starting the go routines
	healthMonitor := routine.NewHealthMonitor(serverqueries)
	healthMonitor.Start()

	// 🌐 Public routes (no authentication required)
	common.RegisterAuthRoutes(publicMux, generalqueries)

	// 🔒 Protected routes (authentication required)
	//  Server Protected Routes
	server.RegisterHealthRoutes(protectedMux, serverqueries)
	server.RegisterAlertRoutes(protectedMux, serverqueries)
	server.RegisterLogRoutes(protectedMux, serverqueries)

	// 👑 Admin-only routes
	//  Common Admin Routes
	common.RegisterSettingsRoutes(adminMux, generalqueries)

	//  Server Admin Routes
	server.RegisterConfig1Routes(adminMux, serverqueries)
	server.RegisterConfig2Routes(adminMux, serverqueries)
	server.RegisterOptimisation(adminMux, serverqueries)

	// Create main mux and apply appropriate middlewares
	mainMux := http.NewServeMux()

	// Mount with different middleware chains
	mainMux.Handle("/api/auth/", config.ApplyPublicMiddlewares(publicMux))
	mainMux.Handle("/api/server/", config.ApplyProtectedMiddlewares(protectedMux))
	mainMux.Handle("/api/network/", config.ApplyProtectedMiddlewares(protectedMux))
	mainMux.Handle("/api/admin/", config.ApplyAdminMiddlewares(adminMux))

	log.Printf("✅ SNSMS backend running on port %s...", config.AppConfig.ServerPort)
	if err := http.ListenAndServe("0.0.0.0:"+config.AppConfig.ServerPort, mainMux); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
