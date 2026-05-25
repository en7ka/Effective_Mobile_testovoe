package config

import (
	"os"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
}

func Load() Config {
	return Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
