package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	"taskboard/graph"
	"taskboard/internal/auth"
	"taskboard/internal/cache"
	"taskboard/internal/config"
	"taskboard/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Database connection
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Successfully connected to database")

	// Redis connection
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("Warning: Redis not available: %v\n", err)
		log.Println("Continuing without cache...")
		redisCache = nil
	} else {
		log.Println("Successfully connected to Redis")
	}

	// JWT Manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTRefreshSecret)
	authMiddleware := auth.NewAuthMiddleware(jwtManager)

	// Repositories
	userRepo := repository.NewUserRepository(dbPool)
	taskRepo := repository.NewTaskRepository(dbPool)

	// GraphQL resolver
	resolver := graph.NewResolver(userRepo, taskRepo, redisCache, jwtManager)

	// GraphQL server
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// Add transports
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	})

	// Add extensions
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: nil,
	})

	// CORS configuration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            cfg.Env == "development",
	})

	// Routes
	mux := http.NewServeMux()

	// GraphQL playground (development only)
	if cfg.Env == "development" {
		mux.Handle("/", playground.Handler("GraphQL Playground", "/query"))
		log.Println("GraphQL Playground enabled at http://localhost:" + cfg.Port)
	}

	// GraphQL endpoint with auth middleware
	mux.Handle("/query", corsHandler.Handler(authMiddleware.Middleware(srv)))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// Check database
		if err := dbPool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready","error":"database unavailable"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server starting on http://localhost:%s", cfg.Port)
	log.Printf("Environment: %s", cfg.Env)
	log.Fatal(http.ListenAndServe(addr, mux))
}