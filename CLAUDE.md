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

# Code Quality
make lint               # Run linter (must show 0 issues)
make test               # Run all tests
make build              # Build all binaries
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

### Linter Configuration & Code Quality

This project uses an enhanced golangci-lint configuration with multiple linters for optimal code quality:

**Enabled Linters:**
- `gosec` - Security checks
- `gocritic` - Performance and style improvements
- `perfsprint` - String formatting optimizations
- `staticcheck` - Advanced static analysis
- `prealloc` - Slice pre-allocation suggestions
- `ineffassign` - Ineffectual assignments
- `unused` - Unused code detection
- `revive` - Comprehensive linting rules
- `misspell` - Spelling corrections
- `unconvert` - Unnecessary type conversions
- `sloglint` - Structured logging improvements

**Code Optimization Patterns Implemented:**

1. **Error Handling Pattern**:
   ```go
   // ✅ Preferred: Use run() pattern for main functions
   func main() {
       if err := run(); err != nil {
           log.Fatal(err)  // Only one exit point
       }
   }
   
   func run() error {
       defer cleanup()  // Guaranteed execution
       return fmt.Errorf("descriptive error: %w", err)
   }
   ```

2. **Slice Operations**:
   ```go
   // ✅ Use slices package for containment checks
   return slices.Contains(slice, item)
   return slices.ContainsFunc(slice, predicate)
   
   // ❌ Avoid manual loops
   for _, item := range slice {
       if item == target { return true }
   }
   ```

3. **String Formatting**:
   ```go
   // ✅ Use concatenation for simple cases
   query := "SELECT * FROM " + table
   
   // ✅ Use errors.New for static messages
   return errors.New("static error message")
   
   // ❌ Avoid fmt.Sprintf for simple concatenation
   query := fmt.Sprintf("SELECT * FROM %s", table)
   ```

4. **Parameter Passing**:
   ```go
   // ✅ Pass large structs by pointer (>80 bytes)
   func Process(filters *BankFilters) error
   
   // ✅ Combine parameters of same type
   func GetConfig(ctx context.Context, bankID, environment string)
   ```

5. **HTTP Requests**:
   ```go
   // ✅ Use http.NoBody for empty request bodies
   req, _ := http.NewRequest("GET", url, http.NoBody)
   ```

6. **Control Flow**:
   ```go
   // ✅ Use switch for multiple string conditions
   switch {
   case strings.Contains(err, "not found"):
       return NotFoundError
   case strings.Contains(err, "invalid"):
       return ValidationError
   }
   ```

**Quality Standards:**
- Maintain 0 linter issues (run `make lint` before commits)
- All defer statements must be guaranteed to execute
- Use named return values for functions with multiple returns
- Pre-allocate slices when size is known: `make([]Type, 0, capacity)`
- Always run tests after code changes: `make test`

**Development Workflow:**
```bash
# Standard development cycle
make lint    # Must show 0 issues
make test    # All tests must pass
make build   # Verify clean compilation
```

## Important Notes

- Always run migrations before seeding: `make migrate-up && make seed`
- API follows OpenAPI 3.0.3 spec in docs/api-documentation.yml
- JWT authentication required for all endpoints except health checks
- Use `docker compose` not `docker-compose` (modern syntax)