.PHONY: help dev build test clean generate lint format docker-up docker-down db-shell

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development environment
	docker-compose up --build

build: ## Build all services
	docker-compose build

test: ## Run all tests
	@echo "Running backend tests..."
	cd backend && go test ./... -v -cover
	@echo "Running frontend tests..."
	cd frontend && npm test

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	cd backend && go test ./tests/integration_test.go -v

generate: ## Generate GraphQL code
	cd backend && gqlgen generate

lint: ## Run linters
	@echo "Linting backend..."
	cd backend && golangci-lint run
	@echo "Linting frontend..."
	cd frontend && npm run lint

format: ## Format code
	@echo "Formatting backend..."
	cd backend && gofmt -w .
	@echo "Formatting frontend..."
	cd frontend && npm run format

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-clean: ## Remove Docker containers and volumes
	docker-compose down -v
	docker system prune -f

db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U postgres -d taskboard

redis-shell: ## Open Redis CLI
	docker-compose exec redis redis-cli

logs: ## Show logs
	docker-compose logs -f

logs-backend: ## Show backend logs
	docker-compose logs -f backend

logs-frontend: ## Show frontend logs
	docker-compose logs -f frontend

clean: ## Clean build artifacts
	rm -rf backend/main
	rm -rf frontend/dist
	rm -rf frontend/node_modules

install: ## Install dependencies
	cd backend && go mod download
	cd frontend && npm install

setup: install generate ## Setup project

migrate-up: ## Run database migrations
	cd backend && go run cmd/migrate/main.go up

migrate-down: ## Rollback database migrations
	cd backend && go run cmd/migrate/main.go down

prod: ## Start production environment
	docker-compose -f docker-compose.prod.yml up -d