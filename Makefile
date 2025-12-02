.PHONY: help \
        backend-build backend-test backend-test-unit backend-test-integration backend-test-race \
        backend-lint backend-check-layers backend-dev backend-dev-run backend-migrate \
        backend-generate-test-data backend-publish-stream backend-publish-stream-fast backend-publish-stream-slow \
        install install-deno frontend-install frontend-setup \
        frontend-dev frontend-build frontend-test frontend-test-unit frontend-generate-client \
        dev-up dev-down dev-logs dev-config clean clean-all \
        docker-build docker-up docker-down

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo 'CAATSM Dashboard - Available Commands'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Backend - Build
backend-build: ## Build the API server binary
	@go build -buildvcs=false -o bin/caatsm ./cmd/server

# Backend - Testing
backend-test: ## Run all Go tests with coverage
	@go test -cover ./...

backend-test-unit: ## Run unit tests only (short)
	@go test -short ./...

backend-test-integration: ## Run integration tests only
	@go test -tags=integration ./...

backend-test-race: ## Run tests with race detector
	@go test -race ./...

# Backend - Code Quality
backend-check-layers: ## Check architecture layer import boundaries
	@./scripts/check-layer-imports.sh

backend-lint: backend-check-layers ## Run golangci-lint and layer checks
	@golangci-lint run ./...

# Backend - Development
backend-dev: ## Run backend with hot reload (air if available, else go run)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found, running go run ./cmd/server"; \
		go run ./cmd/server -config $${CAATSM_CONFIG:-config/config.local.toml}; \
	fi

backend-dev-run: ## Run backend server directly (no hot reload)
	@go run ./cmd/server -config $${CAATSM_CONFIG:-config/config.local.toml}

# Backend - Database
backend-migrate: ## Run database migrations with goose
	@goose -dir migrations postgres "$${CAATSM_DATABASE_DSN:-postgres://caatsm:caatsm@localhost:5432/caatsm?sslmode=disable}" up

# Backend - Utilities
backend-generate-test-data: ## Generate sample telegram data (50 records)
	@go run ./cmd/generate-test-data -count 50

backend-publish-stream: ## Publish live messages to NATS
	@go run ./cmd/publish-stream -config $${CAATSM_CONFIG:-config/config.local.toml} $${CLI_ARGS:-}

backend-publish-stream-fast: ## Publish messages quickly (1/sec, 50 total)
	@CLI_ARGS="-interval 1s -total 50" $(MAKE) backend-publish-stream

backend-publish-stream-slow: ## Publish messages slowly (5s interval)
	@CLI_ARGS="-interval 5s" $(MAKE) backend-publish-stream

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

frontend-install: ## Install frontend dependencies (npm + Deno cache)
	@cd frontend && echo "Installing frontend dependencies..." && \
		npm install --no-save && \
		if command -v deno >/dev/null 2>&1; then \
			echo "Caching TypeScript files with Deno..."; \
			deno cache vite.config.ts svelte.config.js vitest.config.ts playwright.config.ts src/**/*.ts 2>/dev/null || true; \
			echo "Dependencies installed and cached."; \
		else \
			echo "Dependencies installed (Deno not available for caching)."; \
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

frontend-generate-client: ## Generate TypeScript API client from OpenAPI spec
	@echo "Generating TypeScript client from OpenAPI spec..."
	@npm run generate:api-client
	@echo "TypeScript client generated at frontend/src/lib/api/generated"

# Environment
dev-up: ## Start development dependencies (docker-compose.dev.yml)
	@docker compose -f docker-compose.dev.yml --env-file .env.local up -d

dev-down: ## Stop development dependencies
	@docker compose -f docker-compose.dev.yml --env-file .env.local down --remove-orphans

dev-logs: ## Tail dependency logs
	@docker compose -f docker-compose.dev.yml --env-file .env.local logs -f

dev-config: ## Create .env.local from example if missing
	@if [ ! -f .env.local ]; then \
		cp env.local.example .env.local; \
		echo "Created .env.local (edit with your settings)"; \
	else \
		echo ".env.local already exists"; \
	fi


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
