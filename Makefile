# Bank Service Makefile
# Commands to manage the bank service development environment

.PHONY: help clean build test test-short test-coverage format format-check dev run lint lint-fix token token-read token-write token-admin
.DEFAULT_GOAL := help

# Using standard compose.yml file

# Development Commands
dev: ## Start development server with live reload
	wgo run cmd/api/main.go

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: build ## Run the built application
	./bin/api

# Code Quality Commands
test: ## Run all tests with verbose output
	go test ./... -v

test-short: ## Run tests without verbose output
	go test ./...

test-coverage: ## Run tests with coverage report
	go test ./... -cover

lint: ## Run golangci-lint to check code quality
	golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix

format: ## Format code with goimports
	goimports -w .

# Help command
help: ## Show this help message
	@echo "Bank Service - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Database Migration Commands
migrate-up: ## Apply all pending migrations
	go run cmd/migrate/main.go up

migrate-down: ## Roll back all migrations
	go run cmd/migrate/main.go down

migrate-status: ## Show migration status
	go run cmd/migrate/main.go status

migrate-version: ## Show current migration version
	go run cmd/migrate/main.go version

migrate-steps: ## Apply N migration steps (use: make migrate-steps N=1 or N=-1)
	@if [ -z "$(N)" ]; then \
		echo "Usage: make migrate-steps N=<number>"; \
		echo "Example: make migrate-steps N=1 (apply 1 migration)"; \
		echo "Example: make migrate-steps N=-1 (rollback 1 migration)"; \
		exit 1; \
	fi
	docker compose exec app go run cmd/migrate/main.go steps $(N)

# Database Seeding Commands
seed: ## Run all seeders
	go run cmd/seed/main.go run

seed-debug: ## Run all seeders with verbose logging
	go run cmd/seed/main.go debug

seed-reset: ## Reset data and run seeders
	go run cmd/seed/main.go reset

seed-status: ## Show seeding status and record counts
	go run cmd/seed/main.go status

# Database Management
db-clean-data: ## Clean only database data (keeps containers running)
	@echo "🧹 Cleaning database data..."
	docker compose exec postgres psql -U postgres -d bankdb -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "✅ Database data cleaned. Run 'make migrate-up' to restore schema."

db-reset: ## Complete reset with volumes (removes everything)
	@echo "🛑 Stopping all services and removing volumes..."
	docker compose down -v
	@echo "🗄️  Starting all services..."
	docker compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@echo "✅ Complete reset done. Run 'make migrate-up' to apply migrations."

# JWT Token Commands
token: ## Generate JWT token with default permissions (banks:read)
	@echo "🔑 Generating JWT token..."
	@go run cmd/token/main.go

token-read: ## Generate JWT token with read permissions only
	@echo "🔑 Generating JWT token with read permissions..."
	@go run cmd/token/main.go -permissions banks:read

token-write: ## Generate JWT token with write permissions only
	@echo "🔑 Generating JWT token with write permissions..."
	@go run cmd/token/main.go -permissions banks:write

token-admin: ## Generate JWT token with all permissions (read + write)
	@echo "🔑 Generating JWT token with admin permissions..."
	@go run cmd/token/main.go -permissions banks:read,banks:write

token-custom: ## Generate JWT token with custom settings (use: make token-custom PERMISSIONS=banks:read EXPIRY=1h)
	@if [ -z "$(PERMISSIONS)" ]; then \
		echo "Usage: make token-custom PERMISSIONS=<perms> [EXPIRY=<duration>]"; \
		echo "Example: make token-custom PERMISSIONS=banks:read,banks:write EXPIRY=2h"; \
		exit 1; \
	fi
	@echo "🔑 Generating custom JWT token..."
	@if [ -n "$(EXPIRY)" ]; then \
		go run cmd/token/main.go -permissions $(PERMISSIONS) -expiry $(EXPIRY); \
	else \
		go run cmd/token/main.go -permissions $(PERMISSIONS); \
	fi
