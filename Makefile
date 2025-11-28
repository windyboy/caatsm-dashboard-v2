.PHONY: help build test test-unit test-integration test-race lint dev dev-run \
        install install-deno frontend-install frontend-setup \
        frontend-dev frontend-build frontend-test frontend-test-unit \
        dev-up dev-down dev-logs migrate clean clean-all \
        docker-build docker-up docker-down generate-test-data \
        publish-stream publish-stream-fast publish-stream-slow

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo 'CAATSM Dashboard - Available Commands'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Core backend
build: ## Build the API server binary
	@go build -buildvcs=false -o bin/caatsm ./cmd/server

test: ## Run all Go tests with coverage
	@go test -cover ./...

test-unit: ## Run unit tests only (short)
	@go test -short ./...

test-integration: ## Run integration tests only
	@go test -tags=integration ./...

test-race: ## Run tests with race detector
	@go test -race ./...

lint: ## Run golangci-lint
	@golangci-lint run ./...

dev: ## Run backend with hot reload (air if available, else go run)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found, running go run ./cmd/server"; \
		go run ./cmd/server -config $${CAATSM_CONFIG:-config/config.local.toml}; \
	fi

dev-run: ## Run backend with local config
	@go run ./cmd/server -config $${CAATSM_CONFIG:-config/config.local.toml}

# Installation
install: ## Install all dependencies (Go modules, Deno, frontend)
	@echo "Installing Go dependencies..."
	@go mod download
	@go mod tidy
	@echo "Setting up frontend..."
	@$(MAKE) frontend-setup

install-deno: ## Install Deno if not available
	@if command -v deno >/dev/null 2>&1; then \
		echo "Deno is already installed: $$(deno --version)"; \
	else \
		echo "Installing Deno..."; \
		curl -fsSL https://deno.land/install.sh | sh; \
		echo "Deno installed. Please add Deno to your PATH or restart your shell."; \
	fi

frontend-install: ## Cache frontend dependencies with Deno (or npm fallback)
	@cd frontend && if command -v deno >/dev/null 2>&1; then \
		echo "Caching frontend dependencies with Deno..."; \
		deno task --quiet || deno cache deno.json || true; \
		echo "Frontend dependencies cached."; \
	else \
		echo "Deno not found, using npm..."; \
		npm install; \
	fi

frontend-setup: ## Full frontend setup (install Deno + cache dependencies)
	@echo "Setting up frontend..."
	@if ! command -v deno >/dev/null 2>&1; then \
		echo "Deno not found. Installing..."; \
		$(MAKE) install-deno || echo "Please install Deno manually: curl -fsSL https://deno.land/install.sh | sh"; \
	fi
	@$(MAKE) frontend-install

# Frontend
frontend-dev: ## Run frontend dev server (Deno, fallback to npm)
	@cd frontend && if command -v deno >/dev/null 2>&1; then \
		deno task dev; \
	else \
		echo "deno not found, using npm"; \
		npm run dev; \
	fi

frontend-build: ## Build frontend for production (Deno, fallback to npm)
	@cd frontend && if command -v deno >/dev/null 2>&1; then \
		deno task build; \
	else \
		echo "deno not found, using npm"; \
		npm run build; \
	fi

frontend-test: ## Run frontend E2E tests (Playwright)
	@cd frontend && if command -v deno >/dev/null 2>&1; then \
		deno task test; \
	else \
		npm test; \
	fi

frontend-test-unit: ## Run frontend unit tests (Vitest via Deno, fallback to npm)
	@cd frontend && if command -v deno >/dev/null 2>&1; then \
		deno task test:unit; \
	else \
		npm run test:unit; \
	fi

# Environment
dev-up: ## Start development dependencies (docker-compose.dev.yml)
	@docker compose -f docker-compose.dev.yml --env-file .env.local up -d

dev-down: ## Stop development dependencies
	@docker compose -f docker-compose.dev.yml --env-file .env.local down --remove-orphans

dev-logs: ## Tail dependency logs
	@docker compose -f docker-compose.dev.yml --env-file .env.local logs -f

migrate: ## Run database migrations with goose
	@goose -dir migrations postgres "$${CAATSM_DATABASE_DSN:-postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable}" up

# Utilities
generate-test-data: ## Generate sample telegram data (50 records)
	@go run ./cmd/generate-test-data -count 50

publish-stream: ## Publish live messages to NATS
	@go run ./cmd/publish-stream -config $${CAATSM_CONFIG:-config/config.local.toml} $${CLI_ARGS:-}

publish-stream-fast: ## Publish messages quickly (1/sec, 50 total)
	@CLI_ARGS="-interval 1s -total 50" $(MAKE) publish-stream

publish-stream-slow: ## Publish messages slowly (5s interval)
	@CLI_ARGS="-interval 5s" $(MAKE) publish-stream

# Docker
docker-build: ## Build Docker image
	@docker build -t caatsm-dashboard:latest .

docker-up: ## Start docker-compose stack
	@docker compose up -d --build

docker-down: ## Stop docker-compose stack
	@docker compose down --remove-orphans

# Cleanup
clean: ## Remove build artifacts and caches
	@rm -rf bin/ tmp/ .cache/go-build
	@rm -f coverage.out coverage.html
	@rm -rf frontend/.svelte-kit frontend/build frontend/.vite

clean-all: clean ## Clean everything including dependencies
	@rm -rf node_modules frontend/node_modules
