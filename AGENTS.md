# Agent Guidelines for go-banks

## Build/Test/Lint Commands
```bash
make lint          # Run golangci-lint (MUST show 0 issues before commit)
make test          # Run all tests with verbose output
make test-short    # Run tests without verbose output
go test -v ./internal/services -run TestBankService  # Run single test by name
go test -v ./internal/services/bankservice_test.go   # Run specific test file
make build         # Build application binary
make dev           # Start with live reload using wgo
```

## Code Style Guidelines
- **Go 1.25.0**: Use modern syntax - `any` not `interface{}`, range over integers `for i := range 10`
- **Imports**: Group stdlib, external deps, internal packages (separated by blank lines)
- **Testing**: Use testify/assert, testify/mock, testify/require for all tests
- **Error handling**: Return `fmt.Errorf("context: %w", err)` for wrapping, `errors.New()` for simple errors
- **Linting**: Zero tolerance - fix all issues, NEVER modify .golangci.yml to suppress warnings
- **Types**: Prefer specific interfaces over `any`, use generics `APIResponse[T]` for type safety
- **Database**: Use pgx v5 types (`pgtype.Array[string]`), JSONB as `map[string]any`
- **HTTP**: Gin framework patterns, consistent `APIResponse[T]` wrapper, proper status codes
- **Naming**: Go conventions - exported funcs/types CamelCase, unexported camelCase, consts UPPER_SNAKE