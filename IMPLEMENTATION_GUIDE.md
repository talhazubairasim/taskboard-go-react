# TaskBoard Implementation Guide

This guide explains how TaskBoard is put together and how to run, test and deploy it. For a quick start, see the [README](README.md).

## Architecture

```
React (Vite, Apollo Client, Tailwind)
   │  HTTP: GraphQL /query and REST /api/*        WebSocket: GraphQL subscriptions /query
   ▼
Go server (net/http)
   ├── auth middleware (JWT access tokens)
   ├── graph/       GraphQL resolvers (gqlgen)   ┐
   ├── internal/rest  REST handlers               ┘ both call the same services
   ├── internal/service  business rules: validation, permissions, caching, events
   ├── internal/repository  SQL (pgx)          → PostgreSQL
   ├── internal/cache       read-through cache → Redis
   └── internal/pubsub      in-process broker  → subscriptions
```

The clean-architecture rule is simple: handlers and resolvers only translate between the transport and the service layer; services own every rule; repositories only run SQL. Because the REST API and the GraphQL API call the same services, they always behave the same way.

### Backend packages

| Package | Responsibility |
|---|---|
| `cmd/server` | Wires everything together, applies migrations on startup, graceful shutdown |
| `cmd/migrate` | `up` / `down` migration CLI (`make migrate-up`, `make migrate-down`) |
| `internal/config` | Environment variables (see `.env.example`) |
| `internal/database` | Connection with retry, embedded SQL migrations tracked in `schema_migrations` |
| `internal/models` | Domain types shared by all layers |
| `internal/repository` | `UserRepository`, `TaskRepository` (pgx) and error translation |
| `internal/service` | `AuthService`, `UserService`, `TaskService` |
| `internal/auth` | JWT access/refresh tokens, bcrypt passwords, auth middleware |
| `internal/cache` | Redis JSON cache with hit/miss statistics |
| `internal/pubsub` | Fan-out broker for task events |
| `internal/rest` | REST handlers under `/api` |
| `graph` | GraphQL schema, generated code and resolvers |
| `internal/server` | HTTP routing, CORS, security headers, `/health`, `/ready`, `/metrics` |
| `internal/metrics` | Latency recorders for API requests and SQL queries |
| `internal/memstore` | In-memory stores used by tests |

### Permissions

- Anyone signed in can see the board and the team, and create tasks.
- The **creator** or the **current assignee** of a task can edit it, move it between columns and (re)assign it.
- Only the **creator** can delete a task.

### Authentication

- `register` / `login` return a 15-minute access token and a 7-day refresh token (lifetimes are configurable with `JWT_ACCESS_EXPIRY` and `JWT_REFRESH_EXPIRY`).
- Passwords need 8+ characters with upper- and lower-case letters, a number and a special character. They are hashed with bcrypt.
- Requests send `Authorization: Bearer <access token>`. WebSocket connections send the same value as `Authorization` in the `connection_init` payload.
- When a request fails with `UNAUTHENTICATED` (or HTTP 401 for an expired token), the frontend exchanges the refresh token once and retries. If that fails, the user is signed out.

### Caching

`TaskService` and `UserService` read through Redis: single tasks (`task:<id>`), task lists per filter (`tasks:<filter>`) and users (`user:<id>`). Every write invalidates the affected keys. If Redis is unavailable, the server logs a warning and runs without a cache.

### Real-time updates

Every task change publishes an event on the in-process broker. The `taskCreated`, `taskUpdated(taskId)` and `taskDeleted` subscriptions stream these events over WebSocket (`graphql-transport-ws` protocol). The frontend's `useRealtime` hook applies them to the Apollo cache, so every open board updates without reloading.

> The broker is in-process. When you run more than one backend replica (the Kubernetes manifests use 2–10), a client only receives events for changes made on the replica it is connected to. To fan events out across replicas, replace `internal/pubsub` with Redis Pub/Sub. The `Publisher` interface in `internal/service` is the only thing that needs a new implementation.

## Running locally

### With Docker (recommended)

```bash
cp .env.example .env
docker-compose up --build
```

- Frontend: http://localhost:5173
- GraphQL Playground: http://localhost:8080
- REST API: http://localhost:8080/api

### Without Docker

You need Go 1.22+, Node.js 18+ (20 recommended), PostgreSQL 15+ and Redis 7.

```bash
createdb taskboard
cd backend && go run cmd/server/main.go     # migrations run automatically
cd frontend && npm install && npm run dev
```

## API

### GraphQL

The full schema is in [`backend/graph/schema.graphqls`](backend/graph/schema.graphqls). Errors carry `extensions.code`: `UNAUTHENTICATED`, `FORBIDDEN`, `NOT_FOUND`, `BAD_USER_INPUT`, `CONFLICT` or `INTERNAL_SERVER_ERROR`.

### REST

All endpoints take and return JSON. Everything except the auth endpoints needs a bearer token.

| Method | Path | Description |
|---|---|---|
| POST | `/api/auth/register` | `{email, password, name}` → `{token, refreshToken, user}` |
| POST | `/api/auth/login` | `{email, password}` → `{token, refreshToken, user}` |
| POST | `/api/auth/refresh` | `{refreshToken}` → new token pair |
| GET | `/api/me` | The signed-in user |
| PATCH | `/api/me` | `{name?, avatar?}` |
| GET | `/api/users` | All users |
| GET | `/api/users/{id}` | One user |
| GET | `/api/tasks` | Filters: `status`, `priority`, `assignedToId`, `createdById` |
| POST | `/api/tasks` | `{title, description?, status?, priority?, assignedToId?, dueDate?}` |
| GET | `/api/tasks/mine` | Tasks you created |
| GET | `/api/tasks/assigned` | Tasks assigned to you |
| GET | `/api/tasks/{id}` | One task |
| PATCH / PUT | `/api/tasks/{id}` | Partial update. `null` clears `description`, `assignedToId` or `dueDate` |
| DELETE | `/api/tasks/{id}` | Delete (creator only) |
| POST | `/api/tasks/{id}/assign` | `{userId}` |
| POST | `/api/tasks/{id}/unassign` | Clear the assignee |

Errors look like `{"error": "message", "code": "NOT_FOUND"}`, with the matching HTTP status.

### Operations

| Path | Purpose |
|---|---|
| `GET /health` | Liveness |
| `GET /ready` | Readiness: checks PostgreSQL and Redis |
| `GET /metrics` | Request count, API P95, SQL query count and P95, cache hits/misses/hit rate, active subscribers |

## Testing

```bash
make docker-up            # Postgres + Redis for the database tests
make test                 # backend + frontend
make test-integration     # full stack against Postgres + Redis
make coverage             # coverage report; fails under 95%
```

- Backend unit tests use `internal/memstore`, so business rules are tested without a database. Repository, migration and cache tests use real PostgreSQL and Redis, and are skipped when `TEST_DATABASE_URL` is not set.
- `scripts/coverage.sh` measures coverage of hand-written code. It excludes gqlgen's generated files, the in-memory test stores and the `main` packages.
- Frontend tests use Vitest, Testing Library and Apollo's `MockedProvider`. Coverage thresholds are set in `vite.config.ts`.
- CI (`.github/workflows/ci.yml`) runs all of this, plus a Docker smoke test, on every push.

## Performance

`backend/tests/integration_test.go` → `TestPerformanceTargets` exercises a local stack with read-heavy board traffic, then checks the README targets using the server's own `/metrics`:

| Metric | Target | How it is measured |
|---|---|---|
| API response time | < 100 ms P95 | Middleware timing of every HTTP request |
| Cache hit rate | > 80% | Redis cache hits ÷ lookups |
| Database query time | < 50 ms P95 | pgx query tracer |
| WebSocket latency | < 10 ms P95 | From a mutation completing to the subscriber receiving the event |

These are local measurements. Check `/metrics` on your own deployment to see real-world numbers.

## Deployment

### Docker Compose (single server)

```bash
export DOMAIN=taskboard.example.com ACME_EMAIL=you@example.com \
       DB_PASSWORD=... JWT_SECRET=... JWT_REFRESH_SECRET=...
docker-compose -f docker-compose.prod.yml up -d --build
```

Traefik obtains a Let's Encrypt certificate and routes `/query` and `/api` to the backend and everything else to the frontend.

### Kubernetes

1. Build and push the images (see Google Cloud below), then set the image names in `k8s/*-deployment.yaml`.
2. Fill in `k8s/secrets.yaml`, or create the secret with `kubectl create secret`.
3. Set your domain in `k8s/ingress.yaml`.

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/
kubectl get pods -n taskboard
```

### Google Cloud

```bash
gcloud builds submit --tag gcr.io/PROJECT_ID/taskboard-backend backend/
gcloud builds submit --tag gcr.io/PROJECT_ID/taskboard-frontend frontend/
gcloud run deploy taskboard-backend --image gcr.io/PROJECT_ID/taskboard-backend \
  --platform managed --region us-central1 \
  --set-env-vars ENV=production,DATABASE_URL=...,REDIS_URL=...,JWT_SECRET=...,JWT_REFRESH_SECRET=...
```

Cloud Run needs a reachable PostgreSQL (e.g. Cloud SQL) and Redis (e.g. Memorystore with a VPC connector). The frontend image's API URLs are build arguments (`VITE_GRAPHQL_HTTP_URL`, `VITE_GRAPHQL_WS_URL`). For Cloud Build, set them with a `cloudbuild.yaml` or build locally with `docker build --build-arg ...`.
