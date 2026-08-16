package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL string
	HTTPPort    string
	JWTSecret   string
	TokenTTL    time.Duration
}

func Load() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:55432/cafe_scheduling?sslmode=disable"),
		HTTPPort:    getEnv("HTTP_PORT", "18101"),
		JWTSecret:   getEnv("JWT_SECRET", "cafe-scheduling-dev-secret"),
		TokenTTL:    getDurationEnv("TOKEN_TTL", 24*time.Hour),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
