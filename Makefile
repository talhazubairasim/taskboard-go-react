.PHONY: help dev build test test-backend test-frontend test-integration coverage clean generate lint format \
        docker-up docker-down docker-clean db-shell redis-shell logs logs-backend logs-frontend install setup \
        migrate-up migrate-down prod

TEST_DATABASE_URL ?= postgresql://postgres:postgres@localhost:5432/taskboard_test?sslmode=disable
REDIS_TEST_ADDR   ?= localhost:6379

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development environment
	docker compose up --build

build: ## Build all services
	docker compose build

test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests (needs Postgres + Redis: `make docker-up` or local)
	cd backend && TEST_DATABASE_URL="$(TEST_DATABASE_URL)" REDIS_TEST_ADDR="$(REDIS_TEST_ADDR)" go test ./... -v -cover

test-frontend: ## Run frontend tests
	cd frontend && npm test -- --run

test-integration: ## Run integration tests against Postgres + Redis
	cd backend && TEST_DATABASE_URL="$(TEST_DATABASE_URL)" REDIS_TEST_ADDR="$(REDIS_TEST_ADDR)" go test ./tests/... -v -count=1

coverage: ## Backend + frontend coverage (fails under 95%)
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" REDIS_TEST_ADDR="$(REDIS_TEST_ADDR)" ./scripts/coverage.sh
	cd frontend && npm run coverage

generate: ## Generate GraphQL code
	cd backend && go run github.com/99designs/gqlgen generate

lint: ## Run linters
	cd backend && go vet ./... && gofmt -l . | (! grep .)
	cd frontend && npm run lint

format: ## Format code
	cd backend && gofmt -w .
	cd frontend && npm run format

docker-up: ## Start Docker containers
	docker compose up -d

docker-down: ## Stop Docker containers
	docker compose down

docker-clean: ## Remove Docker containers and volumes
	docker compose down -v

db-shell: ## Open PostgreSQL shell
	docker compose exec postgres psql -U postgres -d taskboard

redis-shell: ## Open Redis CLI
	docker compose exec redis redis-cli

logs: ## Show logs
	docker compose logs -f

logs-backend: ## Show backend logs
	docker compose logs -f backend

logs-frontend: ## Show frontend logs
	docker compose logs -f frontend

clean: ## Clean build artifacts
	rm -rf backend/coverage*.out frontend/dist frontend/coverage frontend/node_modules

install: ## Install dependencies
	cd backend && go mod download
	cd frontend && npm install

setup: install generate ## Setup project

migrate-up: ## Run database migrations
	cd backend && go run cmd/migrate/main.go up

migrate-down: ## Rollback the latest database migration
	cd backend && go run cmd/migrate/main.go down

prod: ## Start production environment
	docker compose -f docker-compose.prod.yml up -d --build
