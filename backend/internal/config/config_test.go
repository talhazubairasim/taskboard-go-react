package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	for _, k := range []string{"DATABASE_URL", "REDIS_URL", "PORT", "ENV", "CORS_ALLOWED_ORIGINS", "JWT_ACCESS_EXPIRY", "JWT_REFRESH_EXPIRY", "REDIS_DB"} {
		t.Setenv(k, "")
	}
	cfg := Load()
	if cfg.Port != "8080" || cfg.Env != "development" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Fatalf("redis addr = %q", cfg.RedisAddr)
	}
	if cfg.JWTAccessExpiry != 15*time.Minute || cfg.JWTRefreshExpiry != 7*24*time.Hour {
		t.Fatalf("unexpected expiry defaults")
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("origins = %v", cfg.AllowedOrigins)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://cache:6380/")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("PORT", "9000")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://a.com, http://b.com")
	t.Setenv("JWT_ACCESS_EXPIRY", "30m")
	t.Setenv("JWT_REFRESH_EXPIRY", "3d")
	t.Setenv("CACHE_TASKS_TTL", "60")
	cfg := Load()
	if cfg.RedisAddr != "cache:6380" || cfg.RedisDB != 2 || cfg.Port != "9000" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if cfg.AllowedOrigins[1] != "http://b.com" {
		t.Fatalf("origins not trimmed: %v", cfg.AllowedOrigins)
	}
	if cfg.JWTAccessExpiry != 30*time.Minute || cfg.JWTRefreshExpiry != 72*time.Hour {
		t.Fatalf("expiry parse failed: %v %v", cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	}
	if cfg.CacheTasksTTL != time.Minute {
		t.Fatalf("ttl = %v", cfg.CacheTasksTTL)
	}
}

func TestBadDurationsFallBack(t *testing.T) {
	t.Setenv("JWT_ACCESS_EXPIRY", "soon")
	t.Setenv("JWT_REFRESH_EXPIRY", "xd")
	cfg := Load()
	if cfg.JWTAccessExpiry != 15*time.Minute || cfg.JWTRefreshExpiry != 7*24*time.Hour {
		t.Fatalf("bad values should fall back to defaults")
	}
}
