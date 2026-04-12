package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	JWTSecret       string
	TokenTTL        time.Duration
	WorkerInterval  time.Duration
	MigrationsPath  string
	CORSEnabled     bool
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        env("HTTP_ADDR", ":8081"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       env("JWT_SECRET", "change-me-please"),
		TokenTTL:        envDuration("TOKEN_TTL", 24*time.Hour),
		WorkerInterval:  envDuration("WORKER_INTERVAL", 30*time.Second),
		MigrationsPath:  env("MIGRATIONS_PATH", "migrations"),
		CORSEnabled:     envBool("CORS_ENABLED", true),
		ShutdownTimeout: envDuration("SHUTDOWN_TIMEOUT", 5*time.Second),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
