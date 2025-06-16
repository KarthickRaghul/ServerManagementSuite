package config

import (
	generaldb "backend/db/gen/general"
	serverdb "backend/db/gen/server"
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
	"path/filepath"
)

func GeneralQueries() *generaldb.Queries {
	// Try loading the .env file from project root
	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Failed to get working directory: %v", err)
	}

	envPath := filepath.Join(rootPath, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("❌ Error loading .env file at %s: %v", envPath, err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL not set in .env file")
	}

	dbConn, err := sql.Open("postgres", dbURL)
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
	// Try loading the .env file from project root
	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Failed to get working directory: %v", err)
	}

	envPath := filepath.Join(rootPath, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("❌ Error loading .env file at %s: %v", envPath, err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL not set in .env file")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("❌ Cannot reach DB: %v", err)
	}

	log.Println("✅ Connected to Sever database")
	return serverdb.New(dbConn)
}
