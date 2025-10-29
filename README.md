# TaskBoard - Production-Ready Full Stack Application

A comprehensive task management platform built with Go, React, and modern web technologies.

## 🏗️ Architecture

### Tech Stack
- **Backend**: Go 1.21+ with gqlgen (GraphQL)
- **Frontend**: React 18 + TypeScript + Vite
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Authentication**: JWT with bcrypt password hashing
- **Container Orchestration**: Docker Compose + Kubernetes
- **Cloud**: Google Cloud Platform ready

### Key Features
- 🔐 Secure JWT authentication with refresh tokens
- 👥 User management and profiles
- 📋 Task CRUD with assignments
- 🔄 Real-time updates via GraphQL subscriptions
- 💾 Redis caching for performance
- 🧪 Comprehensive test coverage
- 🚀 Production-ready deployment configs

## 📁 Project Structure

```
taskboard-go-react/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   ├── middleware.go
│   │   │   └── password.go
│   │   ├── cache/
│   │   │   └── redis.go
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── database/
│   │   │   ├── migrations/
│   │   │   │   ├── 001_initial_schema.sql
│   │   │   │   └── 002_add_assignments.sql
│   │   │   └── postgres.go
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   └── task.go
│   │   └── repository/
│   │       ├── user_repository.go
│   │       └── task_repository.go
│   ├── graph/
│   │   ├── schema.graphqls
│   │   ├── schema.resolvers.go
│   │   └── resolver.go
│   ├── tests/
│   │   ├── auth_test.go
│   │   ├── integration_test.go
│   │   └── repository_test.go
│   ├── go.mod
│   ├── go.sum
│   ├── gqlgen.yml
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── auth/
│   │   │   │   ├── Login.tsx
│   │   │   │   ├── Register.tsx
│   │   │   │   └── ProtectedRoute.tsx
│   │   │   ├── tasks/
│   │   │   │   ├── TaskList.tsx
│   │   │   │   ├── TaskItem.tsx
│   │   │   │   ├── TaskForm.tsx
│   │   │   │   └── TaskAssignment.tsx
│   │   │   ├── users/
│   │   │   │   ├── UserProfile.tsx
│   │   │   │   └── UserList.tsx
│   │   │   └── layout/
│   │   │       ├── Header.tsx
│   │   │       └── Sidebar.tsx
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   ├── useTasks.ts
│   │   │   └── useRealtime.ts
│   │   ├── graphql/
│   │   │   ├── queries.ts
│   │   │   ├── mutations.ts
│   │   │   └── subscriptions.ts
│   │   ├── contexts/
│   │   │   └── AuthContext.tsx
│   │   ├── utils/
│   │   │   ├── apollo-client.ts
│   │   │   └── errorHandler.ts
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── tests/
│   │   ├── components/
│   │   └── hooks/
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── Dockerfile
├── k8s/
│   ├── namespace.yaml
│   ├── postgres-deployment.yaml
│   ├── redis-deployment.yaml
│   ├── backend-deployment.yaml
│   ├── frontend-deployment.yaml
│   ├── ingress.yaml
│   └── secrets.yaml
├── docker-compose.yml
├── docker-compose.prod.yml
├── .env.example
├── Makefile
└── README.md
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Make (optional)

### Environment Setup

```bash
# Copy environment file
cp .env.example .env

# Edit with your values
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/taskboard?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-super-secret-jwt-key-change-this
JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-this
```

### Development

```bash
# Using Make
make dev

# Or manually
docker-compose up --build

# Backend will be available at http://localhost:8080
# Frontend will be available at http://localhost:5173
```

### Running Tests

```bash
# Backend tests
cd backend
go test ./... -v -cover

# Frontend tests
cd frontend
npm test

# Integration tests
make test-integration
```

## 📊 Database Migrations

Migrations are automatically applied on startup. Manual migration:

```bash
cd backend
go run cmd/migrate/main.go up
```

## 🔑 Authentication Flow

1. User registers with email/password
2. Password is hashed with bcrypt (cost 12)
3. User logs in and receives:
   - Access token (15 min expiry)
   - Refresh token (7 days expiry)
4. Access token used in Authorization header
5. Refresh token endpoint for new access tokens

## 📝 GraphQL API Examples

### Register User
```graphql
mutation {
  register(input: {
    email: "user@example.com"
    password: "SecurePass123!"
    name: "John Doe"
  }) {
    token
    refreshToken
    user {
      id
      email
      name
    }
  }
}
```

### Login
```graphql
mutation {
  login(email: "user@example.com", password: "SecurePass123!") {
    token
    refreshToken
    user {
      id
      email
      name
    }
  }
}
```

### Create Task
```graphql
mutation {
  createTask(input: {
    title: "Implement Redis caching"
    description: "Add Redis for task list caching"
    status: TODO
    priority: HIGH
  }) {
    id
    title
    description
    status
    priority
    createdBy {
      id
      name
    }
    createdAt
  }
}
```

### Assign Task
```graphql
mutation {
  assignTask(taskId: "123", userId: "456") {
    id
    assignedTo {
      id
      name
      email
    }
  }
}
```

### Get Tasks (with caching)
```graphql
query {
  tasks(filter: { status: TODO }) {
    id
    title
    description
    status
    priority
    assignedTo {
      id
      name
    }
    createdBy {
      id
      name
    }
    createdAt
    updatedAt
  }
}
```

### Real-time Task Updates
```graphql
subscription {
  taskUpdated {
    id
    title
    status
    updatedAt
  }
}
```

## 🎯 Performance Features

### Redis Caching Strategy
- Task lists cached for 5 minutes
- User profiles cached for 10 minutes
- Cache invalidation on mutations
- Cache key patterns: `tasks:all`, `user:{id}`

### Database Optimization
- Indexed columns: email, task status, created_at
- Connection pooling (max 25 connections)
- Prepared statements for common queries

## 🧪 Testing

### Backend Coverage
- Unit tests for auth package
- Repository tests with test database
- Integration tests for GraphQL resolvers
- Middleware tests

### Frontend Coverage
- Component tests with React Testing Library
- Hook tests
- GraphQL mock testing
- E2E tests with Playwright (optional)

## 🚢 Deployment

### Docker Compose Production

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Apply secrets (edit first!)
kubectl apply -f k8s/secrets.yaml

# Deploy services
kubectl apply -f k8s/

# Check status
kubectl get pods -n taskboard
```

### GCP Deployment

```bash
# Build and push images
gcloud builds submit --tag gcr.io/PROJECT_ID/taskboard-backend backend/
gcloud builds submit --tag gcr.io/PROJECT_ID/taskboard-frontend frontend/

# Deploy to Cloud Run
gcloud run deploy taskboard-backend \
  --image gcr.io/PROJECT_ID/taskboard-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## 📈 Monitoring & Observability

- Health check endpoints: `/health` and `/ready`
- Prometheus metrics exposed on `/metrics`
- Structured logging with zerolog
- Error tracking ready for Sentry integration

## 🔒 Security Features

- ✅ Bcrypt password hashing (cost 12)
- ✅ JWT with RS256 signing
- ✅ Refresh token rotation
- ✅ CORS configuration
- ✅ Rate limiting middleware
- ✅ SQL injection prevention (parameterized queries)
- ✅ XSS protection headers
- ✅ HTTPS only in production

## 🛠️ Development Tools

```bash
# Generate GraphQL code
make generate

# Run linters
make lint

# Format code
make format

# Run dev server with hot reload
make dev

# Database shell
make db-shell
```

## 📚 API Documentation

GraphQL Playground available at `http://localhost:8080/` in development mode.

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

MIT License - see LICENSE file for details

## 🙋 Support

For issues and questions, please open a GitHub issue.

---

Built with ❤️ using Go, React, and modern web technologies