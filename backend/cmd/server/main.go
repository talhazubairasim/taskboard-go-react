// Command server runs the TaskBoard API (REST + GraphQL + WebSocket subscriptions).
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"taskboard/internal/auth"
	"taskboard/internal/cache"
	"taskboard/internal/config"
	"taskboard/internal/database"
	"taskboard/internal/metrics"
	"taskboard/internal/pubsub"
	"taskboard/internal/repository"
	"taskboard/internal/server"
	"taskboard/internal/service"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbLatency := metrics.NewRecorder(4096)
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("invalid DATABASE_URL: %v", err)
	}
	poolCfg.ConnConfig.Tracer = metrics.QueryTracer{R: dbLatency}
	pool, err := database.ConnectConfig(ctx, poolCfg, 30)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	log.Println("connected to database")

	n, err := database.MigrateUp(ctx, pool)
	if err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Printf("applied %d migration(s)", n)

	var c service.Cache
	var stats *cache.Stats
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("warning: Redis unavailable, continuing without cache: %v", err)
	} else {
		defer redisCache.Close()
		c, stats = redisCache, redisCache.Stats
		log.Println("connected to Redis")
	}

	jwt := auth.NewJWTManagerWithExpiry(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	users := repository.NewUserRepository(pool)
	tasks := repository.NewTaskRepository(pool)
	broker := pubsub.NewBroker()

	h := server.New(server.Deps{
		Env:            cfg.Env,
		AllowedOrigins: cfg.AllowedOrigins,
		JWT:            jwt,
		AuthSvc:        service.NewAuthService(users, jwt),
		UserSvc:        service.NewUserService(users, c, cfg.CacheUserTTL),
		TaskSvc:        service.NewTaskService(tasks, users, c, broker, cfg.CacheTasksTTL),
		Broker:         broker,
		CacheStats:     stats,
		DBLatency:      dbLatency,
		Ready: func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				return errors.New("database unavailable")
			}
			if redisCache != nil {
				if err := redisCache.Ping(ctx); err != nil {
					return errors.New("cache unavailable")
				}
			}
			return nil
		},
	})

	srv := &http.Server{
		Addr:              cfg.Host + ":" + cfg.Port,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Printf("TaskBoard API listening on http://localhost:%s (env=%s)", cfg.Port, cfg.Env)
		if cfg.Env != "production" {
			log.Printf("GraphQL Playground: http://localhost:%s/", cfg.Port)
		}
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
