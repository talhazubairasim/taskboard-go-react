# TaskBoard — Go + GraphQL + React + PostgreSQL + Docker + K8s + JWT 

Minimal full-stack demo intended to be finished in ~2 days:
- Backend: Go (gqlgen) + pgx + JWT (demo)
- Frontend: React + TypeScript (Vite) + Apollo Client
- DB: PostgreSQL
- Containerized with Docker Compose; optional K8s manifests for demo

## Quick start (docker-compose)
1. From repo root:
docker-compose up --build

2. Backend GraphQL Playground: http://localhost:8080/
3. Frontend: http://localhost:5173/

## Notes
- Run `cd backend && go install github.com/99designs/gqlgen@latest && go mod tidy && gqlgen generate` if you want full gqlgen codegen to run.
- JWT in this scaffold accepts any username/password for demo. Replace with real authentication if using in production.


Final tips / next actions

Copy these files into your taskboard-go-react repository (exact paths).

In backend run:

go install github.com/99designs/gqlgen@latest
go mod tidy
gqlgen generate


(This will create graph/generated.go and graph/models_gen.go required by gqlgen.)

From project root:

docker-compose up --build


Open http://localhost:8080/ for GraphQL playground. Example queries:

mutation { createTask(title: "Hello from Go", description: "demo") { id title } }
query { tasks { id title status } }


Frontend will call the GraphQL endpoint at http://localhost:8080/query (CORS may require adjustments if you run across issues; using the dev containers as above should work).# taskboard-go-react 
