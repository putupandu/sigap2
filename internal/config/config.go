package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPass       string
	DBName       string
	JWTSecret    string
	AppPort      string
	FonnteAPIKey string
}

var AppConfig Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	AppConfig = Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "3306"),
		DBUser:       getEnv("DB_USER", "sigap"),
		DBPass:       getEnv("DB_PASS", "sigap_secret"),
		DBName:       getEnv("DB_NAME", "sigap2"),
		JWTSecret:    getEnv("JWT_SECRET", "super-secret-jwt-key"),
		AppPort:      getEnv("PORT", getEnv("APP_PORT", "3000")),
		FonnteAPIKey: getEnv("FONNTE_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}
