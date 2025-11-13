.PHONY: help build test lint run dev clean docker-build docker-up docker-down migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/caatsm ./cmd/server

test: ## Run all tests (unit + integration)
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-unit: ## Run unit tests only
	go test -v -race -coverprofile=coverage.out -short ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests only
	go test -v -race -tags=integration ./...

race: ## Run tests with race detector
	go test -race ./...

lint: ## Run linter
	golangci-lint run ./...

run: build ## Build and run the application
	./bin/caatsm

dev: ## Run the application with hot reload
	air

clean: ## Clean build artifacts
	rm -rf bin/ tmp/ coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t caatsm-dashboard:latest .

docker-up: ## Start Docker Compose stack
	docker compose up -d

docker-down: ## Stop Docker Compose stack
	docker compose down

migrate: ## Run database migrations
	@echo "Running database migrations..."
	@echo "Please implement migration script in scripts/migrate.sh"

generate: ## Generate templ components
	go run github.com/a-h/templ/cmd/templ@v0.3.960 generate

tidy: ## Run go mod tidy
	go mod tidy

unocss: ## Build UnoCSS (watch mode)
	npm run unocss:dev

unocss-build: ## Build UnoCSS once
	npm run unocss:build

install-deps: ## Install npm dependencies
	npm install
