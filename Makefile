# Makefile for Brokle AI Control Plane
#
# This Makefile provides automation for development, testing, building,
# and deployment of the Brokle platform.

.PHONY: help setup dev build test lint clean docker
.PHONY: dev-backend dev-frontend build-backend build-frontend
.PHONY: migrate-up migrate-down migrate-status seed seed-dev db-reset
.PHONY: docker-build docker-build-dev docker-dev docker-prod

# Default target
help: ## Show this help message
	@echo "Brokle AI Control Plane - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

##@ Development

setup: ## Setup development environment
	@echo "🚀 Setting up development environment..."
	@$(MAKE) install-deps
	@$(MAKE) submodule-init
	@$(MAKE) setup-databases
	@$(MAKE) migrate-up
	@$(MAKE) seed-dev
	@echo "✅ Development environment ready!"

install-deps: ## Install Go and Node.js dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	cd web && pnpm install

setup-databases: ## Start databases with Docker Compose
	@echo "🗄️ Starting databases..."
	docker compose up -d postgres clickhouse redis
	@echo "⏳ Waiting for databases to be ready..."
	@sleep 10

dev: ## Start full stack (server + worker)
	@echo "🔥 Starting full stack development..."
	@$(MAKE) -j2 dev-server dev-worker

dev-server: ## Start HTTP server with hot reload
	@echo "🔥 Starting HTTP server with hot reload..."
	air -c .air.toml

dev-worker: ## Start workers with hot reload
	@echo "🔥 Starting workers with hot reload..."
	air -c .air.worker.toml

dev-frontend: ## Start Next.js development server only
	@echo "⚛️ Starting Next.js development server..."
	cd web && pnpm run dev

##@ Building

build: build-server-oss build-worker-oss ## Build both server and worker (OSS)

build-server-oss: ## Build HTTP server (OSS version)
	@echo "🔨 Building HTTP server (OSS)..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/brokle-server cmd/server/main.go

build-server-enterprise: ## Build HTTP server (Enterprise version)
	@echo "🔨 Building HTTP server (Enterprise)..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags="enterprise" -ldflags="-w -s" -o bin/brokle-server-enterprise cmd/server/main.go

build-worker-oss: ## Build worker process (OSS version)
	@echo "🔨 Building worker process (OSS)..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/brokle-worker cmd/worker/main.go

build-worker-enterprise: ## Build worker process (Enterprise version)
	@echo "🔨 Building worker process (Enterprise)..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags="enterprise" -ldflags="-w -s" -o bin/brokle-worker-enterprise cmd/worker/main.go

build-frontend: ## Build Next.js for production
	@echo "🔨 Building Next.js frontend..."
	cd web && pnpm run build

build-all: build-server-oss build-worker-oss build-server-enterprise build-worker-enterprise ## Build all variants
	@echo "✅ All builds complete!"

build-dev: build-dev-server ## Build server for development (default)

build-dev-server: ## Build server for development (faster, with debug info)
	@echo "🔨 Building server for development..."
	mkdir -p bin
	go build -o bin/brokle-dev-server cmd/server/main.go

build-dev-worker: ## Build worker for development (faster, with debug info)
	@echo "🔨 Building worker for development..."
	mkdir -p bin
	go build -o bin/brokle-dev-worker cmd/worker/main.go

##@ Database Operations

migrate-up: ## Run all pending migrations
	@echo "📊 Running database migrations..."
	go run cmd/migrate/main.go up

migrate-down: ## Rollback one migration
	@echo "📊 Rolling back one migration..."
	go run cmd/migrate/main.go down

migrate-status: ## Show migration status
	@echo "📊 Migration status:"
	go run cmd/migrate/main.go status

migrate-reset: ## Reset all databases (WARNING: destroys data)
	@echo "⚠️ Resetting databases (this will destroy all data)..."
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		go run cmd/migrate/main.go postgres down -steps=999; \
		go run cmd/migrate/main.go clickhouse down -steps=999; \
		$(MAKE) migrate-up; \
	fi

seed: ## Seed databases with production data
	@echo "🌱 Seeding databases with production data..."
	go run cmd/migrate/main.go seed -env production

seed-dev: ## Seed databases with development data
	@echo "🌱 Seeding databases with development data..."
	go run cmd/migrate/main.go seed -env dev

db-reset: migrate-reset seed-dev ## Reset databases and seed with dev data

create-migration: ## Create new migration (usage: make create-migration DB=postgres NAME=add_users_table)
	@if [ -z "$(DB)" ] || [ -z "$(NAME)" ]; then \
		echo "Usage: make create-migration DB=postgres|clickhouse NAME=migration_name"; \
		exit 1; \
	fi
	go run cmd/migrate/main.go create --db=$(DB) --name=$(NAME)

##@ Testing

test: ## Run all tests
	@echo "🧪 Running all tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	go test -v -short ./...

test-integration: ## Run integration tests only
	@echo "🧪 Running integration tests..."
	go test -v -tags=integration ./tests/integration/...

test-e2e: ## Run end-to-end tests
	@echo "🧪 Running E2E tests..."
	cd tests/e2e && pnpm test

test-load: ## Run load tests
	@echo "🧪 Running load tests..."
	cd tests/load && go test -v ./...

##@ Code Quality

lint: lint-go lint-frontend ## Run all linters

lint-go: ## Run Go linter
	@echo "🔍 Running Go linter..."
	golangci-lint run --config .golangci.yml

lint-frontend: ## Run frontend linter
	@echo "🔍 Running frontend linter..."
	cd web && pnpm run lint

fmt: ## Format Go code
	@echo "💅 Formatting Go code..."
	go fmt ./...
	goimports -w .

fmt-frontend: ## Format frontend code
	@echo "💅 Formatting frontend code..."
	cd web && pnpm run format

security-scan: ## Run security scans
	@echo "🔒 Running security scans..."
	gosec ./...

##@ Docker

docker-build-server: ## Build server Docker image
	@echo "🐳 Building server Docker image..."
	docker build -t brokle/server:latest .

docker-build-worker: ## Build worker Docker image
	@echo "🐳 Building worker Docker image..."
	docker build -f Dockerfile.worker -t brokle/worker:latest .

docker-build: docker-build-server docker-build-worker ## Build all Docker images
	@echo "🐳 Building dashboard Docker image..."
	docker build -f web/Dockerfile -t brokle/dashboard:latest ./web
	@echo "✅ All Docker images built!"

docker-up: ## Start all services with docker compose
	@echo "🐳 Starting all services..."
	docker compose up -d --build

docker-prod: ## Start production environment with scaling
	@echo "🐳 Starting production environment (3 backends + 10 workers)..."
	docker compose -f docker compose.yml -f docker compose.prod.yml up -d --build

docker-down: ## Stop all Docker containers
	@echo "🐳 Stopping Docker containers..."
	docker compose down

docker-stop: docker-down ## Alias for docker-down

docker-clean: ## Clean up Docker resources
	@echo "🐳 Cleaning up Docker resources..."
	docker compose down -v --remove-orphans
	docker system prune -f

docker-logs-backend: ## Show backend logs
	@echo "📋 Backend logs:"
	docker compose logs -f backend

docker-logs-worker: ## Show worker logs
	@echo "📋 Worker logs:"
	docker compose logs -f worker

docker-logs: ## Show all logs
	docker compose logs -f

##@ Health & Status

health: ## Check health of all services
	@echo "🏥 Checking service health..."
	@echo "API Server:"
	@curl -f http://localhost:8080/health || echo "❌ API Server not responding"
	@echo "Next.js Dashboard:"
	@curl -f http://localhost:3000 || echo "❌ Dashboard not responding"
	@echo "PostgreSQL:"
	@docker exec -it $$(docker compose ps -q postgres) pg_isready -U brokle || echo "❌ PostgreSQL not ready"
	@echo "ClickHouse:"
	@docker exec -it $$(docker compose ps -q clickhouse) clickhouse-client --query "SELECT 1" || echo "❌ ClickHouse not ready"
	@echo "Redis:"
	@docker exec -it $$(docker compose ps -q redis) redis-cli ping || echo "❌ Redis not ready"

status: ## Show status of all services
	@echo "📊 Service Status:"
	@docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

logs: ## Show logs for all services
	docker compose logs -f

logs-backend: ## Show backend logs
	docker compose logs -f backend

logs-frontend: ## Show frontend logs
	docker compose logs -f frontend

logs-db: ## Show database logs
	docker compose logs -f postgres clickhouse redis

##@ Deployment

deploy-staging: ## Deploy to staging environment
	@echo "🚀 Deploying to staging..."
	./scripts/deploy/deploy-staging.sh

deploy-prod: ## Deploy to production environment
	@echo "🚀 Deploying to production..."
	./scripts/deploy/deploy-prod.sh

k8s-apply: ## Apply Kubernetes manifests
	@echo "☸️ Applying Kubernetes manifests..."
	kubectl apply -f deployments/kubernetes/

k8s-delete: ## Delete Kubernetes resources
	@echo "☸️ Deleting Kubernetes resources..."
	kubectl delete -f deployments/kubernetes/

helm-install: ## Install with Helm
	@echo "⛵ Installing with Helm..."
	helm install brokle deployments/helm/brokle/

helm-upgrade: ## Upgrade with Helm
	@echo "⛵ Upgrading with Helm..."
	helm upgrade brokle deployments/helm/brokle/

##@ SDK Management

submodule-init: ## Initialize all submodules (included in setup)
	@echo "📦 Initializing SDK submodules..."
	git submodule update --init --recursive

submodule-update: ## Update submodules to latest commits
	@echo "🔄 Updating SDK submodules..."
	git submodule update --recursive --remote

submodule-sync: ## Sync submodule URLs after remote changes
	@echo "🔄 Syncing submodule URLs..."
	git submodule sync --recursive

submodule-status: ## Show status of all submodules
	@echo "📊 SDK Submodule Status:"
	@git submodule status --recursive

submodule-clean: ## Clean submodule working directories
	@echo "🧹 Cleaning SDK submodules..."
	git submodule foreach --recursive git clean -fd
	git submodule foreach --recursive git reset --hard

##@ Utilities

clean: ## Clean build artifacts and caches
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/.next/
	rm -rf web/node_modules/
	rm -rf web/.pnpm-store/
	rm -f coverage.out coverage.html
	go clean -cache
	go clean -modcache

clean-builds: ## Clean only build artifacts (keep dependencies)
	@echo "🧹 Cleaning build artifacts only..."
	rm -rf bin/
	rm -rf web/.next/
	rm -f coverage.out coverage.html

fresh-start: clean setup ## Clean everything and start fresh
	@echo "🆕 Fresh start complete!"

docs-serve: ## Serve documentation locally
	@echo "📚 Serving documentation..."
	cd docs && python3 -m http.server 8000

docs-generate: ## Generate API documentation
	@echo "📚 Generating API documentation..."
	swag init -g cmd/server/main.go --output docs/swagger

shell-backend: ## Get shell access to backend container
	docker compose exec backend sh

shell-db: ## Get shell access to PostgreSQL
	docker compose exec postgres psql -U brokle -d brokle

shell-redis: ## Get shell access to Redis
	docker compose exec redis redis-cli

shell-clickhouse: ## Get shell access to ClickHouse
	docker compose exec clickhouse clickhouse-client

##@ Monitoring

metrics: ## Show Prometheus metrics
	@echo "📊 Prometheus metrics:"
	curl -s http://localhost:9090/metrics

grafana: ## Open Grafana dashboard
	@echo "📊 Opening Grafana dashboard..."
	open http://localhost:3000

prometheus: ## Open Prometheus UI
	@echo "📊 Opening Prometheus UI..."
	open http://localhost:9090

##@ Environment Variables

env-check: ## Check required environment variables
	@echo "🔍 Checking environment variables..."
	@./scripts/check-env.sh

env-example: ## Generate .env.example file
	@echo "📝 Generating .env.example..."
	@./scripts/generate-env-example.sh

##@ Release

release: ## Create a new release
	@echo "🏷️ Creating new release..."
	@./scripts/release.sh

changelog: ## Generate changelog
	@echo "📝 Generating changelog..."
	@git log --pretty=format:"- %s" $(shell git describe --tags --abbrev=0)..HEAD

##@ Development Helpers

watch: ## Watch for changes and restart server
	@echo "👀 Watching for changes..."
	air -c .air.toml

hot-reload: watch ## Alias for watch

install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

##@ Information

version: ## Show version information
	@echo "Brokle AI Control Plane"
	@echo "Version: $(shell git describe --tags --always --dirty)"
	@echo "Commit: $(shell git rev-parse HEAD)"
	@echo "Build Date: $(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
	@echo "Go Version: $(shell go version)"
	@echo "Node Version: $(shell node --version)"

info: version ## Show project information

##@ Default

.DEFAULT_GOAL := help