.PHONY: help build test test-unit test-integration test-race lint dev clean \
        frontend-dev frontend-build frontend-test frontend-test-unit \
        dev-up dev-down migrate install

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo 'CAATSM Dashboard - Available Commands'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Common Commands:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'For more commands, run: task --list'

# ============================================
# Core Development Commands (Quick Access)
# ============================================

build: ## Build the application
	@task build

test: ## Run all tests with coverage
	@task test

test-unit: ## Run unit tests only
	@task test:unit

test-integration: ## Run integration tests only
	@task test:integration

test-race: ## Run tests with race detector
	@task test:race

test-pretty: ## Run tests with gotestsum (prettier output)
	@task test:pretty

test-pretty-unit: ## Run unit tests with gotestsum
	@task test:pretty:unit

lint: ## Run linter (golangci-lint)
	@task lint

dev: ## Run backend with hot reload (air)
	@task dev

clean: ## Clean build artifacts
	@task clean

# ============================================
# Frontend Commands
# ============================================

frontend-dev: ## Run frontend dev server (Deno/npm)
	@task frontend:dev

frontend-build: ## Build frontend for production
	@task frontend:build

frontend-test: ## Run frontend E2E tests (Playwright)
	@task frontend:test

frontend-test-unit: ## Run frontend unit tests (Vitest)
	@task frontend:test:unit

# ============================================
# Development Environment
# ============================================

dev-up: ## Start development dependencies (Docker)
	@task dev:up

dev-down: ## Stop development dependencies
	@task dev:down

migrate: ## Run database migrations
	@task migrate

# ============================================
# Setup Commands
# ============================================

install: ## Install all dependencies (Go, npm, Deno)
	@echo "Installing dependencies..."
	@task install-deps
	@task frontend:setup
	@echo ""
	@echo "✓ Installation complete!"
	@echo "Next steps:"
	@echo "  1. Create config: task dev:config"
	@echo "  2. Start services: make dev-up"
	@echo "  3. Run migrations: make migrate"
	@echo "  4. Start backend: make dev"
	@echo "  5. Start frontend: make frontend-dev"

# ============================================
# Legacy/Convenience Commands
# ============================================

tidy: ## Run go mod tidy
	@go mod tidy

run: build ## Build and run the application
	@./bin/caatsm

docker-build: ## Build Docker image
	@task docker:build

docker-up: ## Start Docker Compose stack
	@task docker:up

docker-down: ## Stop Docker Compose stack
	@task docker:down
