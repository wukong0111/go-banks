package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for database/sql

	"github.com/wukong0111/go-banks/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	command := os.Args[1]

	switch command {
	case "run":
		if err := runSeeders(ctx, cfg.Database, false); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		fmt.Println("✅ Seeders executed successfully")

	case "debug":
		if err := runSeeders(ctx, cfg.Database, true); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		fmt.Println("✅ Seeders executed successfully (debug mode)")

	case "reset":
		if err := resetAndSeed(ctx, cfg.Database); err != nil {
			log.Fatalf("Reset and seed failed: %v", err)
		}
		fmt.Println("✅ Database reset and seeded successfully")

	case "status":
		if err := showStatus(ctx, cfg.Database); err != nil {
			log.Fatalf("Failed to show status: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runSeeders(ctx context.Context, cfg *config.DatabaseConfig, debug bool) error {
	db, err := sql.Open("pgx", cfg.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close database connection: %v", closeErr)
		}
	}()

	// Verify database connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Check if migrations are applied
	if err := checkMigrationsApplied(ctx, db); err != nil {
		return fmt.Errorf("migrations check failed: %w", err)
	}

	// Get seeder files
	seederFiles, err := getSeederFiles()
	if err != nil {
		return fmt.Errorf("failed to get seeder files: %w", err)
	}

	// Execute seeders in order
	for _, file := range seederFiles {
		fmt.Printf("📦 Executing seeder: %s\n", file)
		if err := executeSQLFile(ctx, db, file, debug); err != nil {
			return fmt.Errorf("failed to execute seeder %s: %w", file, err)
		}
	}

	return nil
}

func resetAndSeed(ctx context.Context, cfg *config.DatabaseConfig) error {
	db, err := sql.Open("pgx", cfg.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close database connection: %v", closeErr)
		}
	}()

	// Verify database connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Check if migrations are applied
	if err := checkMigrationsApplied(ctx, db); err != nil {
		return fmt.Errorf("migrations check failed: %w", err)
	}

	fmt.Println("🗑️  Cleaning existing seed data...")

	// Clean tables in reverse order to respect foreign keys
	cleanQueries := []string{
		"TRUNCATE TABLE bank_environment_configs CASCADE",
		"TRUNCATE TABLE banks CASCADE",
		"TRUNCATE TABLE bank_groups CASCADE",
	}

	for _, query := range cleanQueries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to clean table: %w", err)
		}
	}

	// Run seeders
	return runSeeders(ctx, cfg, false)
}

func showStatus(ctx context.Context, cfg *config.DatabaseConfig) error {
	db, err := sql.Open("pgx", cfg.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close database connection: %v", closeErr)
		}
	}()

	fmt.Println("📊 Seeder Status:")
	fmt.Printf("  Database: %s\n", cfg.Name)
	fmt.Printf("  Host: %s:%d\n", cfg.Host, cfg.Port)

	// Count records in each table
	tables := []string{"bank_groups", "banks", "bank_environment_configs"}

	for _, table := range tables {
		var count int
		// #nosec G201 - table names are hardcoded, no injection risk
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			return fmt.Errorf("failed to count records in %s: %w", table, err)
		}
		fmt.Printf("  %s: %d records\n", table, count)
	}

	return nil
}

func checkMigrationsApplied(ctx context.Context, db *sql.DB) error {
	// Check if schema_migrations table exists
	var exists bool
	query := `SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'schema_migrations'
    )`

	if err := db.QueryRowContext(ctx, query).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !exists {
		return fmt.Errorf("migrations have not been applied yet. Run 'make migrate-up' first")
	}

	return nil
}

func getSeederFiles() ([]string, error) {
	seedersDir := "seeders"

	files, err := filepath.Glob(filepath.Join(seedersDir, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to find seeder files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no seeder files found in %s directory", seedersDir)
	}

	// Sort files to ensure correct execution order
	sort.Strings(files)

	return files, nil
}

func executeSQLFile(ctx context.Context, db *sql.DB, filePath string, debug bool) error {
	// #nosec G304 - filePath comes from controlled seeder directory
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	if debug {
		fmt.Printf("🔍 File content length: %d bytes\n", len(content))
	}

	// Remove comments and empty lines first
	lines := strings.Split(string(content), "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comment lines
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	cleanContent := strings.Join(cleanLines, "\n")

	if debug {
		fmt.Printf("🔍 Clean content:\n%s\n", cleanContent)
	}

	// For multi-line INSERT statements, execute the entire content as one statement
	// if it contains INSERT, otherwise split by semicolon
	if strings.Contains(strings.ToUpper(cleanContent), "INSERT") {
		// Execute as single statement
		result, err := db.ExecContext(ctx, cleanContent)
		if err != nil {
			if debug {
				fmt.Printf("❌ Error executing statement: %v\n", err)
			}
			return fmt.Errorf("failed to execute INSERT statement: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if debug || rowsAffected > 0 {
			fmt.Printf("✅ INSERT executed successfully - %d rows affected\n", rowsAffected)
		}

		if rowsAffected == 0 {
			fmt.Printf("⚠️  INSERT executed but no rows were affected (possible conflict)\n")
		}
	} else {
		// Split by semicolon for other types of statements
		statements := strings.Split(cleanContent, ";")

		for i, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if debug {
				fmt.Printf("🔍 Executing statement %d: %s\n", i+1, stmt)
			}

			result, err := db.ExecContext(ctx, stmt)
			if err != nil {
				if debug {
					fmt.Printf("❌ Error executing statement %d: %v\n", i+1, err)
				}
				return fmt.Errorf("failed to execute statement %d: %w", i+1, err)
			}

			rowsAffected, _ := result.RowsAffected()
			if debug {
				fmt.Printf("✅ Statement %d executed successfully - %d rows affected\n", i+1, rowsAffected)
			}
		}
	}

	return nil
}

func printUsage() {
	fmt.Println("Bank Service Seeder Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  seed <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  run     Execute all seeders")
	fmt.Println("  debug   Execute all seeders with verbose logging")
	fmt.Println("  reset   Clean existing data and re-run seeders")
	fmt.Println("  status  Show seeding status and record counts")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  seed run")
	fmt.Println("  seed debug")
	fmt.Println("  seed reset")
	fmt.Println("  seed status")
}
