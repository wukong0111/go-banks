# GEMINI.md

## Project Overview

This is a Go project that provides a bank service API. It uses the Gin web framework, PostgreSQL for the database, and JWT for authentication. The project is containerized using Docker and managed with a `Makefile`.

**Key Technologies:**

*   **Go:** The primary programming language.
*   **Gin:** A web framework for building APIs.
*   **PostgreSQL:** The relational database.
*   **Docker:** For containerization of services.
*   **JWT:** For securing the API with JSON Web Tokens.
*   **golang-migrate:** For database migrations.

**Architecture:**

The project follows a standard Go project layout:

*   `cmd/`: Contains the main applications for the API, database migrations, and seeding.
*   `internal/`: Contains the core business logic, including handlers, models, repositories, and services.
*   `migrations/`: Contains the SQL migration files.
*   `seeders/`: Contains SQL files for seeding the database with initial data.
*   `pkg/`: (Not present, but a common Go pattern) would contain reusable code.

## Building and Running

The project uses a `Makefile` to simplify the development workflow.

**Prerequisites:**

*   Go (version 1.22 or higher)
*   Docker and Docker Compose
*   `make`

**Key Commands:**

*   **Start the development environment:**
    ```bash
    make dev
    ```
    This command starts the development server with live reload.

*   **Build the application:**
    ```bash
    make build
    ```
    This command compiles the application and creates a binary in the `bin/` directory.

*   **Run the application:**
    ```bash
    make run
    ```
    This command runs the built application.

*   **Run tests:**
    ```bash
    make test
    ```
    This command runs all the tests in the project.

*   **Database Migrations:**
    *   `make migrate-up`: Apply all pending migrations.
    *   `make migrate-down`: Roll back all migrations.
    *   `make migrate-status`: Show the migration status.

*   **Database Seeding:**
    *   `make seed`: Run all seeders.

## Development Conventions

*   **Linting:** The project uses `golangci-lint` for code linting. Run `make lint` to check the code and `make lint-fix` to automatically fix issues. The linting rules are defined in `.golangci.yml`.
*   **Formatting:** The project uses `goimports` for code formatting. Run `make format` to format the code.
*   **Testing:** The project uses the standard Go testing library and the `testify` package for assertions. Test files are located next to the source files with a `_test.go` suffix.
*   **Authentication:** The API is secured with JWT. The `Makefile` provides commands to generate JWT tokens with different permissions:
    *   `make token`: Generate a token with default permissions.
    *   `make token-read`: Generate a token with read-only permissions.
    *   `make token-write`: Generate a token with write-only permissions.
    *   `make token-admin`: Generate a token with admin permissions.
