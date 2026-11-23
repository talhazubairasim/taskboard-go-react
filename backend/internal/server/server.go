// Package server wires the REST API, the GraphQL API (HTTP and WebSocket),
// health checks and metrics into one http.Handler.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/coder/websocket"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/ast"

	"taskboard/graph"
	"taskboard/internal/auth"
	"taskboard/internal/cache"
	"taskboard/internal/metrics"
	"taskboard/internal/pubsub"
	"taskboard/internal/rest"
	"taskboard/internal/service"
)

// Deps are the parts the HTTP layer needs.
type Deps struct {
	Env            string
	AllowedOrigins []string
	JWT            *auth.JWTManager
	AuthSvc        *service.AuthService
	UserSvc        *service.UserService
	TaskSvc        *service.TaskService
	Broker         *pubsub.Broker
	// Ready reports whether dependencies (database, cache) are reachable.
	Ready func(ctx context.Context) error
	// Optional statistics shown on /metrics.
	CacheStats *cache.Stats
	APILatency *metrics.Recorder
	DBLatency  *metrics.Recorder
}

// New builds the root handler.
func New(d Deps) http.Handler {
	if d.APILatency == nil {
		d.APILatency = metrics.NewRecorder(1024)
	}
	mw := auth.NewAuthMiddleware(d.JWT)
	mux := http.NewServeMux()

	// GraphQL over HTTP and WebSocket
	gql := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		AuthSvc: d.AuthSvc, UserSvc: d.UserSvc, TaskSvc: d.TaskSvc, Broker: d.Broker,
	}}))
	gql.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Implementation:        transport.CoderWebsocketImplementation{AcceptOptions: acceptOptions(d)},
		// The frontend sends {"Authorization": "Bearer <token>"} in connection_init.
		InitFunc: func(ctx context.Context, p transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			header := p.Authorization()
			if header == "" {
				if v, ok := p["Authorization"].(string); ok {
					header = v
				}
			}
			if header == "" {
				return ctx, &p, nil
			}
			claims, err := mw.ClaimsFromHeader(header)
			if err != nil {
				return ctx, nil, errors.New("invalid or expired token")
			}
			return auth.ContextWithClaims(ctx, claims), &p, nil
		},
	})
	gql.AddTransport(transport.Options{})
	gql.AddTransport(transport.GET{})
	gql.AddTransport(transport.POST{})
	gql.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	gql.Use(extension.Introspection{})
	gql.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})
	gql.SetErrorPresenter(graph.ErrorPresenter)
	mux.Handle("/query", gql)

	if d.Env != "production" {
		mux.Handle("GET /{$}", playground.Handler("TaskBoard GraphQL Playground", "/query"))
	}

	// REST
	(&rest.Handler{Auth: d.AuthSvc, Users: d.UserSvc, Tasks: d.TaskSvc}).Register(mux)

	// Operations
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "timestamp": time.Now().UTC().Format(time.RFC3339)})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if d.Ready != nil {
			if err := d.Ready(ctx); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready", "error": err.Error()})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, snapshot(d))
	})

	h := mw.Middleware(mux)
	h = d.APILatency.Middleware(h)
	h = securityHeaders(h)
	return cors.New(cors.Options{
		AllowedOrigins:   d.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(h)
}

// acceptOptions controls which browser origins may open a WebSocket.
// Development accepts any origin; otherwise only CORS_ALLOWED_ORIGINS.
func acceptOptions(d Deps) websocket.AcceptOptions {
	if d.Env == "development" {
		return websocket.AcceptOptions{InsecureSkipVerify: true}
	}
	var patterns []string
	for _, o := range d.AllowedOrigins {
		if o == "*" {
			return websocket.AcceptOptions{InsecureSkipVerify: true}
		}
		if u, err := url.Parse(o); err == nil && u.Host != "" {
			patterns = append(patterns, u.Host)
		}
	}
	return websocket.AcceptOptions{OriginPatterns: patterns}
}

// Metrics is the body of GET /metrics.
type Metrics struct {
	Requests        int64   `json:"requests"`
	APILatencyP95Ms float64 `json:"apiLatencyP95Ms"`
	DBQueries       int64   `json:"dbQueries"`
	DBLatencyP95Ms  float64 `json:"dbLatencyP95Ms"`
	CacheHits       int64   `json:"cacheHits"`
	CacheMisses     int64   `json:"cacheMisses"`
	CacheHitRate    float64 `json:"cacheHitRate"`
	Subscribers     int     `json:"subscribers"`
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

func snapshot(d Deps) Metrics {
	m := Metrics{
		Requests:        d.APILatency.Count(),
		APILatencyP95Ms: ms(d.APILatency.Percentile(95)),
	}
	if d.DBLatency != nil {
		m.DBQueries = d.DBLatency.Count()
		m.DBLatencyP95Ms = ms(d.DBLatency.Percentile(95))
	}
	if d.CacheStats != nil {
		m.CacheHits, m.CacheMisses, m.CacheHitRate = d.CacheStats.Snapshot()
	}
	if d.Broker != nil {
		m.Subscribers = d.Broker.Subscribers()
	}
	return m
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
