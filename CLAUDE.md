# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture

Go banking service API with PostgreSQL database. Uses pgx v5 for database connectivity and Gin for HTTP routing.

**Core Structure:**
- `cmd/` - Entry points for different binaries (api, migrate, seed)
- `internal/` - Private application code
  - `config/` - Environment-based configuration loading
  - `models/` - Database models using pgx types for PostgreSQL
  - `database/` - Migration service wrapper around golang-migrate
  - `handlers/`, `repository/`, `services/` - API layers (currently empty)
- `migrations/` - Database schema migrations (001-004 covering bank_groups, banks, bank_environment_configs, triggers)
- `seeders/` - SQL data seeding files
- `docs/` - API documentation (OpenAPI spec in api-documentation.yml)

**Database:**
- PostgreSQL with JSONB fields for bank_codes, keywords, attributes, status codes
- Environment-based configurations (sandbox, production, uat, test)
- Primary tables: bank_groups, banks, bank_environment_configs

## Development Commands

**Essential Commands:**
```bash
# Start development environment
docker compose up -d     # Start PostgreSQL container
make migrate-up          # Apply migrations
make seed                # Load test data
make dev                 # Start with live reload

# Database management
make migrate-status      # Check migration status
make db-reset           # Complete database reset
```

For all available commands see `Makefile` or run `make help`

## Configuration

Environment variables with defaults:
- `PORT=8080` - API server port
- `DB_HOST=localhost`, `DB_PORT=5432`, `DB_USER=postgres`, `DB_PASSWORD=password`, `DB_NAME=bankdb`
- `JWT_SECRET=your-super-secret-jwt-key`, `JWT_EXPIRY=24h`

## Database Models

Models use pgx v5 types:
- `pgtype.Array[string]` for PostgreSQL arrays 
- `map[string]any` for JSONB fields
- Standard Go types with proper JSON/DB tags

Environment type enum: sandbox, production, uat, test

## Code Style Guidelines

- **Modern Go syntax**: Always use `any` instead of `interface{}` (Go 1.18+)
- **No deprecated types**: Avoid legacy syntax that has modern equivalents
- **Use slices package**: Prefer `slices.Contains()` over manual loops for slice operations (Go 1.21+)

## Important Notes

- Always run migrations before seeding: `make migrate-up && make seed`
- API follows OpenAPI 3.0.3 spec in docs/api-documentation.yml
- JWT authentication required for all endpoints except health checks
- Use `docker compose` not `docker-compose` (modern syntax)