package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address     string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		Address:     getEnv("HTTP_ADDRESS", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
