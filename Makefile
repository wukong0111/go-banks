# Bank Service Makefile
# Commands to manage the bank service development environment

.PHONY: help clean build test format format-check dev run
.DEFAULT_GOAL := help

# Using standard compose.yml file

# Development Commands
dev: ## Start development server with live reload
	wgo run cmd/api/main.go

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: build ## Run the built application
	./bin/api

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
	docker compose exec db psql -U postgres -d bankdb -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "✅ Database data cleaned. Run 'make migrate-up' to restore schema."

db-reset: ## Complete reset with volumes (removes everything)
	@echo "🛑 Stopping all services and removing volumes..."
	docker compose down -v
	@echo "🗄️  Starting all services..."
	docker compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@echo "✅ Complete reset done. Run 'make migrate-up' to apply migrations."
