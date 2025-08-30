# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🎯 Quick Start

**Get up and running fast:**
```bash
# Start development environment
docker compose up -d     # Start PostgreSQL container
make migrate-up          # Apply migrations
make seed                # Load test data
make dev                 # Start with live reload

# Before any commit
make lint               # Must show 0 issues
make test               # All tests must pass
make build              # Verify clean compilation
```

**Most used commands:**
- `make help` - See all available commands
- `make db-reset` - Reset database completely
- `make migrate-status` - Check migration status

## 🏗️ Architecture Overview

**Go 1.25.0 Banking Service API** with PostgreSQL database. Uses pgx v5 for database connectivity and Gin for HTTP routing.

**Core Structure:**
- `cmd/` - Entry points (api, migrate, seed, token)
- `internal/` - Private application code
  - `auth/` - JWT authentication service
  - `config/` - Environment-based configuration
  - `handlers/` - HTTP request handlers  
  - `middleware/` - Auth and other middleware
  - `models/` - Database models with pgx v5 types
  - `repository/` - Data access layer
  - `services/` - Business logic layer
- `migrations/` - Database schema migrations
- `seeders/` - SQL data seeding files
- `docs/` - OpenAPI 3.0.3 specification

**Database Architecture:**
- PostgreSQL with JSONB for flexible data (bank_codes, keywords, attributes)
- Environment-based configurations: sandbox, production, uat, test
- Primary tables: bank_groups, banks, bank_environment_configs

## 💻 Development

### Environment Setup

**Prerequisites:**
- Go 1.25.0
- Docker & Docker Compose
- PostgreSQL (via Docker)

**Configuration:**
```bash
# Environment variables with defaults
PORT=8080                                    # API server port
DB_HOST=localhost, DB_PORT=5432             # Database connection
DB_USER=postgres, DB_PASSWORD=password      # Database credentials  
DB_NAME=bankdb                              # Database name
JWT_SECRET=your-super-secret-jwt-key        # JWT signing secret
JWT_EXPIRY=24h                              # Token expiration
```

### Development Workflow

**Standard development cycle:**
```bash
# 1. Start services
docker compose up -d && make migrate-up && make seed

# 2. Develop with live reload
make dev

# 3. Before committing
make lint    # Must show 0 issues
make test    # All tests must pass  
make build   # Verify compilation

# 4. Database operations
make migrate-up      # Apply new migrations
make seed           # Load fresh test data
make db-reset       # Complete reset when needed
```

## 📝 Code Standards

### Go 1.25.0 Modern Features

**Required version:** Go 1.25.0 - Use latest language features:
- **Range over integers**: `for i := range 10 { ... }` instead of `for i := 0; i < 10; i++`
- **Enhanced type inference**: Better generic type deduction
- **Improved iterators**: Use new iteration patterns when applicable
- **Modern syntax**: `any` instead of `interface{}` (Go 1.18+)
- **Slices package**: `slices.Contains()` over manual loops (Go 1.21+)

### Type Safety Guidelines

**Core principle: Maximize type safety while using modern Go syntax**

1. **Syntax modernization** (always do this):
   ```go
   // ✅ Modern Go syntax
   func process(data any) error           // Use 'any' not 'interface{}'
   
   // ❌ Deprecated syntax  
   func process(data interface{}) error   // Don't use 'interface{}'
   ```

2. **Type safety optimization** (prefer when possible):
   ```go
   // ✅ Preferred: Specific interfaces for polymorphism
   type BankDetails interface {
       GetBank() *Bank
       GetType() BankDetailsType
   }
   func GetBankDetails(...) (BankDetails, error)
   
   // ✅ Acceptable: 'any' with modern syntax when needed
   func GetBankDetails(...) (any, error)
   
   // ❌ Avoid: Old interface{} syntax
   func GetBankDetails(...) (interface{}, error)
   ```

3. **Generic types for reusability**:
   ```go
   // ✅ Preferred: Type-safe generics
   type APIResponse[T any] struct {
       Success bool `json:"success"`
       Data    T    `json:"data,omitempty"`
       Error   *string `json:"error,omitempty"`
   }
   
   // ✅ Acceptable: any with type switches
   func handleResponse(data any) error {
       switch v := data.(type) {
       case string: return handleString(v)
       case int: return handleInt(v)
       default: return fmt.Errorf("unsupported type: %T", v)
       }
   }
   ```

**When `any` is perfectly acceptable:**
- JSON unmarshaling: `map[string]any`
- Database JSONB fields: `map[string]any` 
- Third-party library interfaces requiring it
- Generic containers with proper type constraints

### Code Patterns

1. **Error Handling Pattern**:
   ```go
   // ✅ Use run() pattern for main functions
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
   // ✅ Use slices package (Go 1.21+)
   return slices.Contains(validEnvs, env)
   return slices.ContainsFunc(items, predicate)
   
   // ✅ Range over integers (Go 1.23+)  
   for i := range 10 {
       process(items[i])
   }
   ```

3. **String Operations**:
   ```go
   // ✅ Simple concatenation
   query := "SELECT * FROM " + table
   
   // ✅ Static error messages
   return errors.New("bank not found")
   
   // ✅ Complex formatting only when needed
   return fmt.Errorf("failed to process bank %s: %w", bankID, err)
   ```

4. **HTTP Patterns**:
   ```go
   // ✅ Use http.NoBody for empty requests
   req, _ := http.NewRequest("GET", url, http.NoBody)
   
   // ✅ Proper error handling
   switch {
   case strings.Contains(err, "not found"):
       return NotFoundError
   case strings.Contains(err, "invalid"):
       return ValidationError
   }
   ```

### Testing Standards

**Framework:** testify (consistent across all tests)
```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

// ✅ Use require for critical assertions
require.NoError(t, err)
require.NotNil(t, result)

// ✅ Use assert for value checks
assert.Equal(t, expected, actual)
assert.Contains(t, slice, item)

// ✅ Use testify/mock for mocking
type MockRepository struct {
    mock.Mock
}
```

## 🔧 Tools & Configuration

### Linter Configuration

**golangci-lint** with strict quality standards - **0 issues tolerated**

**Enabled linters:**
- `gosec` - Security vulnerability detection
- `gocritic` - Performance and style improvements  
- `staticcheck` - Advanced static analysis
- `revive` - Comprehensive Go linting
- `perfsprint` - String formatting optimizations
- `prealloc` - Slice pre-allocation suggestions
- `ineffassign` - Ineffectual assignment detection
- `unused` - Dead code elimination
- `misspell` - Spelling corrections

**Quality enforcement:**
```bash
make lint    # Must return "0 issues"
make test    # All tests must pass
make build   # Clean compilation required
```

**⚠️ CRITICAL LINTER POLICY:**
- **NEVER modify `.golangci.yml` to suppress warnings by exclusion**
- **NEVER add file exclusions or disable linters to silence issues**
- **ALWAYS fix the underlying code issue causing the warning**
- **Only acceptable config changes:**
  - Adjusting thresholds for legitimate edge cases (e.g., slog.Record size)
  - Enabling new linters to improve code quality
  - Adding comments to document why specific thresholds are needed

**Rationale:** Suppressing linter warnings hides real issues and degrades code quality over time. Every warning should be addressed by improving the code, not by silencing the tool.

## 📚 Domain Knowledge

### Banking Domain Concepts

**Environments:** 
- `sandbox` - Development/testing environment
- `production` - Live banking operations  
- `uat` - User acceptance testing
- `test` - Automated testing environment

**Core Models:**
- `Bank` - Financial institution with country, API type, environments
- `BankGroup` - Collection of related banks
- `BankEnvironmentConfig` - Environment-specific bank configuration
- `BankDetails` - Polymorphic response interface (single vs multiple environments)

**Data Patterns:**
```go
// JSONB fields for flexible data
type Bank struct {
    BankCodes   map[string]any `json:"bank_codes,omitempty"`
    Keywords    pgtype.Array[string] `json:"keywords"`
    Attributes  map[string]any `json:"attributes,omitempty"`
}

// Generic API responses
type APIResponse[T any] struct {
    Success    bool               `json:"success"`
    Data       T                  `json:"data,omitempty"`
    Error      *string            `json:"error,omitempty"`
    Pagination *Pagination        `json:"pagination,omitempty"`
}
```

### API Design Principles

**OpenAPI 3.0.3 Specification:** All endpoints documented in `docs/api-documentation.yml`

**Authentication:** JWT required for all endpoints except health checks
- Permissions: `banks:read`, `banks:write`
- Token validation via middleware
- Claims include subject and permissions array

**Response Patterns:**
- Consistent `APIResponse[T]` wrapper
- Proper HTTP status codes
- Structured error messages
- Pagination for list endpoints

## ⚠️ Important Notes

**Critical workflows:**
- **Always** run migrations before seeding: `make migrate-up && make seed`
- **Never** commit with linter issues: `make lint` must show `0 issues`
- **Always** test before committing: `make test` must pass

**Environment specifics:**
- Use `docker compose` not `docker-compose` (modern Docker syntax)
- PostgreSQL via Docker container (not local installation)
- Go 1.25.0 required - leverage modern language features

**Database operations:**
- Migrations are versioned and sequential (001, 002, 003, 004)
- Seeds depend on migrations being current
- Use `make db-reset` for clean slate during development

**Server Operations:**
- **Graceful shutdown**: Server handles SIGTERM/SIGINT properly for zero-downtime deployments
- **Health checks**: `/health` for liveness probe, `/ready` for readiness probe with DB connectivity check
- **Timeouts configured**: 15s read/write, 60s idle, 30s graceful shutdown
- **Production ready**: Compatible with Kubernetes, Docker Swarm, and other container orchestrators