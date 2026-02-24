package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string

	// Google Sheets
	SpreadsheetID string
	SheetName     string

	// JWT
	JWTSecret     string
	JWTExpiration int // hours

	// Frontend
	FrontendURL string
}

var AppConfig *Config

func Load() {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		// Server
		Port: getEnv("PORT", "8080"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "weeklyds"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Google OAuth
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:  getEnv("GOOGLE_REDIRECT_URI", "http://localhost:8080/api/v1/auth/callback"),

		// Google Sheets
		SpreadsheetID: getEnv("SPREADSHEET_ID", "1Xlb_RMkAEShWkqMAuXoJX6Nb9wig4MUCy7dZTULK56s"),
		SheetName:     getEnv("SHEET_NAME", "DashboardTemplate"),

		// JWT
		JWTSecret:     getEnv("JWT_SECRET", "weekly-dashboard-secret-key-change-in-production"),
		JWTExpiration: getEnvInt("JWT_EXPIRATION_HOURS", 24),

		// Frontend
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}

	log.Println("Configuration loaded successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
