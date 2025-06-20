// config/env.go
package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfiguration struct {
	ClientPort     string
	ClientProtocol string
	DatabaseURL    string
	JWTSecret      string
	ServerPort     string
	LogLevel       string

	// ✅ Add SMTP configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

var AppConfig *AppConfiguration

func LoadConfig() {
	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Failed to get working directory: %v", err)
	}

	envPath := filepath.Join(rootPath, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️ Warning: .env file not found at %s, using OS environment variables", envPath)
	}

	// Parse SMTP port
	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		log.Printf("⚠️ Invalid SMTP_PORT, using default 587")
		smtpPort = 587
	}

	AppConfig = &AppConfiguration{
		ClientPort:     getEnv("CLIENT_PORT", "8080"),
		ClientProtocol: getEnv("CLIENT_PROTOCOL", "http"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		ServerPort:     getEnv("SERVER_PORT", "8000"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),

		// ✅ SMTP configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     smtpPort,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
	}

	// Validate required fields
	if AppConfig.JWTSecret == "" {
		log.Fatal("❌ JWT_SECRET environment variable is required")
	}
	if AppConfig.DatabaseURL == "" {
		log.Fatal("❌ DATABASE_URL environment variable is required")
	}
	if AppConfig.SMTPUsername == "" || AppConfig.SMTPPassword == "" {
		log.Printf("⚠️ Warning: SMTP credentials not configured, email alerts will be disabled")
	}

	log.Printf("✅ Configuration loaded - Server Port: %s, Client Port: %s, SMTP: %s:%d",
		AppConfig.ServerPort, AppConfig.ClientPort, AppConfig.SMTPHost, AppConfig.SMTPPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetClientURL(host string, endpoint string) string {
	return AppConfig.ClientProtocol + "://" + host + ":" + AppConfig.ClientPort + endpoint
}
