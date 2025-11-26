.PHONY: help build test lint run dev clean docker-build docker-up docker-down migrate install-deno frontend-setup frontend-install frontend-dev frontend-build frontend-test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/caatsm ./cmd/server

test: ## Run all tests (unit + integration)
	go test -v -race -coverprofile=coverage.out -timeout=5m ./...
	go tool cover -html=coverage.out -o coverage.html

test-unit: ## Run unit tests only
	go test -v -race -coverprofile=coverage.out -short -timeout=2m ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests only
	go test -v -race -tags=integration -timeout=10m ./...

race: ## Run tests with race detector
	go test -race ./...

lint: ## Run linter
	golangci-lint run -buildvcs=false ./...

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

migrate: ## Run database migrations (uses Taskfile)
	@echo "Running database migrations..."
	@echo "Use 'task migrate' for full migration support (goose or psql)"
	@task migrate


tidy: ## Run go mod tidy
	go mod tidy

unocss: ## Build UnoCSS (watch mode)
	npm run unocss:dev

unocss-build: ## Build UnoCSS once
	npm run unocss:build

install-deps: ## Install npm dependencies
	npm install

install-deno: ## Install Deno if not already available
	@if ! command -v deno > /dev/null 2>&1; then \
		echo "Installing Deno..."; \
		curl -fsSL https://deno.land/install.sh | sh || exit 1; \
		echo "Deno installed successfully. Please restart your shell or source your profile."; \
		echo "To add Deno to PATH, run: export PATH=\"$$HOME/.deno/bin:$$PATH\""; \
	else \
		echo "Deno is already installed: $$(deno --version)"; \
	fi

frontend-install: ## Pre-cache frontend dependencies with Deno
	@cd frontend && deno cache src/**/*.ts src/**/*.svelte

frontend-setup: ## Full frontend setup: install Deno, cache dependencies, and verify setup
	@echo "Setting up frontend development environment..."
	@$(MAKE) install-deno
	@if command -v deno > /dev/null 2>&1; then \
		echo "Deno found, caching frontend dependencies..."; \
		cd frontend && deno cache src/**/*.ts src/**/*.svelte || true; \
		echo "Frontend dependencies cached successfully"; \
	else \
		echo "Warning: Deno not found in PATH. Please restart your shell or run:"; \
		echo "  export PATH=\"$$HOME/.deno/bin:$$PATH\""; \
		echo "Then run 'make frontend-install' manually"; \
		echo ""; \
		echo "Falling back to npm installation..."; \
		cd frontend && if [ ! -d node_modules ]; then \
			npm install || exit 1; \
			echo "npm dependencies installed"; \
		else \
			echo "npm dependencies already installed"; \
		fi; \
	fi
	@echo ""
	@echo "Frontend setup complete!"
	@echo "To start development: make frontend-dev"

frontend-dev: ## Run frontend development server (uses Deno, fallback to npm)
	@cd frontend && if command -v deno > /dev/null 2>&1; then \
		echo "Using Deno to run frontend dev server..."; \
		deno task dev; \
	else \
		echo "Deno not found, using npm..."; \
		npm run dev; \
	fi

frontend-build: ## Build frontend for production (uses Deno, fallback to npm)
	@cd frontend && if command -v deno > /dev/null 2>&1; then \
		echo "Using Deno to build frontend..."; \
		deno task build; \
	else \
		echo "Deno not found, using npm..."; \
		npm run build; \
	fi

frontend-test: ## Run frontend E2E tests (Playwright)
	@cd frontend && if command -v deno > /dev/null 2>&1; then \
		echo "Using Deno to run frontend E2E tests..."; \
		deno run -A npm:@playwright/test test; \
	else \
		echo "Deno not found, using npm..."; \
		npm test; \
	fi

frontend-test-unit: ## Run frontend unit tests (Vitest)
	@cd frontend && if command -v deno > /dev/null 2>&1; then \
		echo "Using Deno to run frontend unit tests..."; \
		deno task test:unit; \
	else \
		echo "Deno not found, using npm..."; \
		npm run test:unit; \
	fi
