// config/db.go
package config

import (
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func GeneralQueries() *generaldb.Queries {
	// Use the loaded configuration
	if AppConfig == nil {
		log.Fatal("❌ Configuration not loaded. Call LoadConfig() first.")
	}

	dbConn, err := sql.Open("postgres", AppConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("❌ Cannot reach DB: %v", err)
	}

	log.Println("✅ Connected to General database")
	return generaldb.New(dbConn)
}

func ServerQueries() *serverdb.Queries {
	// Use the loaded configuration
	if AppConfig == nil {
		log.Fatal("❌ Configuration not loaded. Call LoadConfig() first.")
	}

	dbConn, err := sql.Open("postgres", AppConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("❌ Cannot reach DB: %v", err)
	}

	log.Println("✅ Connected to Server database")
	return serverdb.New(dbConn)
}
