package config

import (
	"os"
	"time"
)

type Config struct {
	Port             string
	DatabaseURL      string
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      mustEnv("DATABASE_URL"),
		JWTSecret:        mustEnv("JWT_SECRET"),
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}
