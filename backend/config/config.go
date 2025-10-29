package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Database
	DatabaseURL string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret        string
	JWTRefreshSecret string

	// Server
	Port string
	Host string
	Env  string

	// CORS
	AllowedOrigins []string
}

func LoadConfig() *Config {
	cfg := &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/taskboard?sslmode=disable"),
		RedisAddr:        getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getEnvAsInt("REDIS_DB", 0),
		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-this-in-production"),
		Port:             getEnv("PORT", "8080"),
		Host:             getEnv("HOST", "0.0.0.0"),
		Env:              getEnv("ENV", "development"),
		AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}