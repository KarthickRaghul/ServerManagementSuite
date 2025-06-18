// config/env.go
package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type AppConfiguration struct {
	ClientPort     string
	ClientProtocol string
	DatabaseURL    string
	JWTSecret      string
	ServerPort     string
	LogLevel       string
}

var AppConfig *AppConfiguration

// LoadConfig loads environment variables from .env file
func LoadConfig() {
	// Get the working directory
	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Failed to get working directory: %v", err)
	}

	// Load .env file
	envPath := filepath.Join(rootPath, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️ Warning: .env file not found at %s, using OS environment variables", envPath)
	}

	// Initialize configuration
	AppConfig = &AppConfiguration{
		ClientPort:     getEnv("CLIENT_PORT", "8080"),
		ClientProtocol: getEnv("CLIENT_PROTOCOL", "http"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		ServerPort:     getEnv("SERVER_PORT", "8000"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if AppConfig.JWTSecret == "" {
		log.Fatal("❌ JWT_SECRET environment variable is required")
	}
	if AppConfig.DatabaseURL == "" {
		log.Fatal("❌ DATABASE_URL environment variable is required")
	}

	log.Printf("✅ Configuration loaded - Server Port: %s, Client Port: %s",
		AppConfig.ServerPort, AppConfig.ClientPort)
}

// getEnv gets environment variable with fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetClientURL builds the complete client URL
func GetClientURL(host string, endpoint string) string {
	return AppConfig.ClientProtocol + "://" + host + ":" + AppConfig.ClientPort + endpoint
}
