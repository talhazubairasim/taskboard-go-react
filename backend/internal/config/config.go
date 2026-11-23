// Package config loads runtime configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime settings for the TaskBoard backend.
type Config struct {
	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret        string
	JWTRefreshSecret string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	Port string
	Host string
	Env  string

	AllowedOrigins []string

	CacheTasksTTL time.Duration
	CacheUserTTL  time.Duration
}

// Load reads configuration from the environment, falling back to development defaults.
func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/taskboard?sslmode=disable"),
		RedisAddr:        RedisAddr(getEnv("REDIS_URL", "localhost:6379")),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getEnvAsInt("REDIS_DB", 0),
		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-this-in-production"),
		JWTAccessExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		Port:             getEnv("PORT", "8080"),
		Host:             getEnv("HOST", "0.0.0.0"),
		Env:              getEnv("ENV", "development"),
		AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		CacheTasksTTL:    time.Duration(getEnvAsInt("CACHE_TASKS_TTL", 300)) * time.Second,
		CacheUserTTL:     time.Duration(getEnvAsInt("CACHE_USER_TTL", 600)) * time.Second,
	}
}

// RedisAddr accepts either "host:port" or "redis://host:port" and returns "host:port".
func RedisAddr(v string) string {
	v = strings.TrimPrefix(v, "redis://")
	return strings.TrimSuffix(v, "/")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsDuration accepts Go durations ("15m") and a day suffix ("7d").
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	if strings.HasSuffix(v, "d") {
		if days, err := strconv.Atoi(strings.TrimSuffix(v, "d")); err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
		return defaultValue
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	parts := strings.Split(v, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
