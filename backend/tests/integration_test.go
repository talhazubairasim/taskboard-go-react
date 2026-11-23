// Integration tests run the real server against PostgreSQL and Redis.
//
//	TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/taskboard_test?sslmode=disable \
//	REDIS_TEST_ADDR=localhost:6379 go test ./tests/... -v
//
// They are skipped when TEST_DATABASE_URL is not set.
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"taskboard/internal/auth"
	"taskboard/internal/cache"
	"taskboard/internal/database"
	"taskboard/internal/metrics"
	"taskboard/internal/pubsub"
	"taskboard/internal/repository"
	"taskboard/internal/server"
	"taskboard/internal/service"
)

const password = "Str0ng!Pass"

type stack struct {
	t   *testing.T
	url string
}

// newStack wires the server exactly as cmd/server does.
func newStack(t *testing.T) *stack {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	auth.BcryptCost = bcrypt.MinCost
	ctx := context.Background()

	dbLatency := metrics.NewRecorder(4096)
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.Tracer = metrics.QueryTracer{R: dbLatency}
	pool, err := database.ConnectConfig(ctx, cfg, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := database.MigrateUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE tasks, users CASCADE"); err != nil {
		t.Fatal(err)
	}

	var c service.Cache
	var stats *cache.Stats
	redisAddr := os.Getenv("REDIS_TEST_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if rc, err := cache.NewRedisCache(redisAddr, "", 14); err == nil {
		_ = rc.DeletePattern(ctx, "*")
		t.Cleanup(func() { _ = rc.Close() })
		c, stats = rc, rc.Stats
	} else {
		t.Logf("running without Redis: %v", err)
	}

	jwt := auth.NewJWTManager("integration-secret", "integration-refresh")
	users := repository.NewUserRepository(pool)
	tasks := repository.NewTaskRepository(pool)
	broker := pubsub.NewBroker()
	srv := httptest.NewServer(server.New(server.Deps{
		Env: "test", JWT: jwt, AllowedOrigins: []string{"*"},
		AuthSvc: service.NewAuthService(users, jwt),
		UserSvc: service.NewUserService(users, c, time.Minute),
		TaskSvc: service.NewTaskService(tasks, users, c, broker, time.Minute),
		Broker:  broker, CacheStats: stats, DBLatency: dbLatency,
		Ready: func(ctx context.Context) error { return pool.Ping(ctx) },
	}))
	t.Cleanup(srv.Close)
	return &stack{t: t, url: srv.URL}
}

func (s *stack) call(method, path, token string, body, out interface{}) int {
	s.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, s.url+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.t.Fatal(err)
	}
	defer res.Body.Close()
	if out != nil {
		_ = json.NewDecoder(res.Body).Decode(out)
	}
	return res.StatusCode
}

type session struct {
	Token string `json:"token"`
	User  struct {
		ID string `json:"id"`
	} `json:"user"`
}

func (s *stack) register(email string) session {
	var a session
	if code := s.call("POST", "/api/auth/register", "", map[string]string{"email": email, "password": password, "name": email}, &a); code != 201 {
		s.t.Fatalf("register: %d", code)
	}
	return a
}

type task struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	AssignedTo *struct {
		ID string `json:"id"`
	} `json:"assignedTo"`
}

func TestTaskIntegration(t *testing.T) {
	s := newStack(t)
	alice := s.register("alice@integration.test")
	bob := s.register("bob@integration.test")

	// REST create
	var created task
	if code := s.call("POST", "/api/tasks", alice.Token, map[string]interface{}{
		"title": "Integration Test Task", "description": "end to end", "priority": "HIGH", "assignedToId": bob.User.ID,
	}, &created); code != http.StatusCreated {
		t.Fatalf("create: %d", code)
	}
	if created.AssignedTo == nil || created.AssignedTo.ID != bob.User.ID {
		t.Fatalf("assignee not set: %+v", created)
	}

	// GraphQL sees the same data
	var gql struct {
		Data struct {
			Task struct {
				Title     string `json:"title"`
				CreatedBy struct {
					ID string `json:"id"`
				} `json:"createdBy"`
			} `json:"task"`
		} `json:"data"`
	}
	s.call("POST", "/query", alice.Token, map[string]string{"query": `{ task(id:"` + created.ID + `"){ title createdBy{ id } } }`}, &gql)
	if gql.Data.Task.Title != "Integration Test Task" || gql.Data.Task.CreatedBy.ID != alice.User.ID {
		t.Fatalf("graphql task: %+v", gql)
	}

	// the assignee moves it across the board
	for _, st := range []string{"IN_PROGRESS", "REVIEW", "DONE"} {
		var moved task
		if code := s.call("PATCH", "/api/tasks/"+created.ID, bob.Token, map[string]string{"status": st}, &moved); code != 200 || moved.Status != st {
			t.Fatalf("move to %s: %d %+v", st, code, moved)
		}
	}
	var done []task
	s.call("GET", "/api/tasks?status=DONE", alice.Token, nil, &done)
	if len(done) != 1 {
		t.Fatalf("filter DONE: %d", len(done))
	}

	// delete
	if code := s.call("DELETE", "/api/tasks/"+created.ID, bob.Token, nil, nil); code != http.StatusForbidden {
		t.Fatalf("assignee delete: %d", code)
	}
	if code := s.call("DELETE", "/api/tasks/"+created.ID, alice.Token, nil, nil); code != http.StatusNoContent {
		t.Fatalf("delete: %d", code)
	}
	if code := s.call("GET", "/api/tasks/"+created.ID, alice.Token, nil, nil); code != http.StatusNotFound {
		t.Fatalf("get deleted: %d", code)
	}
	if code := s.call("GET", "/ready", "", nil, nil); code != 200 {
		t.Fatalf("ready: %d", code)
	}
}

// TestPerformanceTargets checks the numbers in the README's Performance table
// on a local stack: API P95 < 100ms, cache hit rate > 80%, DB query P95 < 50ms
// and WebSocket event latency < 10ms.
func TestPerformanceTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in -short mode")
	}
	s := newStack(t)
	a := s.register("perf@integration.test")
	for i := 0; i < 20; i++ {
		s.call("POST", "/api/tasks", a.Token, map[string]string{"title": "perf task"}, nil)
	}
	// typical read-heavy board traffic
	for i := 0; i < 300; i++ {
		s.call("GET", "/api/tasks", a.Token, nil, nil)
		if i%10 == 0 {
			s.call("POST", "/query", a.Token, map[string]string{"query": "{ tasks { id title createdBy { name } } }"}, nil)
		}
	}

	// WebSocket latency: time from the mutation returning to the event arriving
	wsURL := "ws" + strings.TrimPrefix(s.url, "http") + "/query"
	conn, _, err := (&websocket.Dialer{Subprotocols: []string{"graphql-transport-ws"}}).Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.WriteJSON(map[string]interface{}{"type": "connection_init", "payload": map[string]string{"Authorization": "Bearer " + a.Token}})
	var msg map[string]interface{}
	_ = conn.ReadJSON(&msg)
	_ = conn.WriteJSON(map[string]interface{}{"id": "1", "type": "subscribe", "payload": map[string]string{"query": "subscription { taskCreated { id } }"}})
	time.Sleep(100 * time.Millisecond)

	wsLatency := metrics.NewRecorder(64)
	for i := 0; i < 20; i++ {
		var created task
		s.call("POST", "/api/tasks", a.Token, map[string]string{"title": "ws"}, &created)
		sent := time.Now()
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			if err := conn.ReadJSON(&msg); err != nil {
				t.Fatal(err)
			}
			if msg["type"] == "next" {
				break
			}
		}
		wsLatency.Observe(time.Since(sent))
	}

	var m server.Metrics
	s.call("GET", "/metrics", "", nil, &m)
	wsP95 := float64(wsLatency.Percentile(95).Microseconds()) / 1000
	t.Logf("API P95 %.2fms over %d requests | DB P95 %.2fms over %d queries | cache hit rate %.1f%% | WebSocket P95 %.2fms",
		m.APILatencyP95Ms, m.Requests, m.DBLatencyP95Ms, m.DBQueries, m.CacheHitRate*100, wsP95)

	if m.APILatencyP95Ms >= 100 {
		t.Errorf("API P95 %.2fms, target < 100ms", m.APILatencyP95Ms)
	}
	if m.DBLatencyP95Ms >= 50 {
		t.Errorf("DB P95 %.2fms, target < 50ms", m.DBLatencyP95Ms)
	}
	if m.CacheHits+m.CacheMisses > 0 && m.CacheHitRate <= 0.8 {
		t.Errorf("cache hit rate %.1f%%, target > 80%%", m.CacheHitRate*100)
	}
	if wsP95 >= 10 {
		t.Errorf("WebSocket P95 %.2fms, target < 10ms", wsP95)
	}
}
