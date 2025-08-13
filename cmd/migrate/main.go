package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/wukong0111/go-banks/internal/config"
	"github.com/wukong0111/go-banks/internal/database"
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

	command := os.Args[1]

	switch command {
	case "up":
		if err := runUp(cfg.Database); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("✅ Migrations applied successfully")

	case "down":
		if err := runDown(cfg.Database); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("✅ Migrations rolled back successfully")

	case "steps":
		if len(os.Args) < 3 {
			fmt.Println("Usage: migrate steps <number>")
			fmt.Println("Example: migrate steps 1 (up 1 step)")
			fmt.Println("Example: migrate steps -1 (down 1 step)")
			os.Exit(1)
		}

		n, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid step number: %v", err)
		}

		if err := runSteps(cfg.Database, n); err != nil {
			log.Fatalf("Migration steps failed: %v", err)
		}
		fmt.Printf("✅ Migration steps (%d) completed successfully\n", n)

	case "version":
		if err := showVersion(cfg.Database); err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}

	case "status":
		if err := showStatus(cfg.Database); err != nil {
			log.Fatalf("Failed to get status: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runUp(cfg *config.DatabaseConfig) error {
	ms, err := database.NewMigrationService(cfg)
	if err != nil {
		return err
	}
	defer ms.Close()

	return ms.Up()
}

func runDown(cfg *config.DatabaseConfig) error {
	ms, err := database.NewMigrationService(cfg)
	if err != nil {
		return err
	}
	defer ms.Close()

	return ms.Down()
}

func runSteps(cfg *config.DatabaseConfig, n int) error {
	ms, err := database.NewMigrationService(cfg)
	if err != nil {
		return err
	}
	defer ms.Close()

	return ms.Steps(n)
}

func showVersion(cfg *config.DatabaseConfig) error {
	ms, err := database.NewMigrationService(cfg)
	if err != nil {
		return err
	}
	defer ms.Close()

	version, dirty, err := ms.Version()
	if err != nil {
		return err
	}

	if version == 0 {
		fmt.Println("📋 Database version: None (no migrations applied)")
	} else {
		status := "clean"
		if dirty {
			status = "dirty"
		}
		fmt.Printf("📋 Database version: %d (%s)\n", version, status)
	}

	return nil
}

func showStatus(cfg *config.DatabaseConfig) error {
	ms, err := database.NewMigrationService(cfg)
	if err != nil {
		return err
	}
	defer ms.Close()

	version, dirty, err := ms.Version()
	if err != nil {
		return err
	}

	fmt.Println("📊 Migration Status:")
	fmt.Printf("  Database: %s\n", cfg.Name)
	fmt.Printf("  Host: %s:%d\n", cfg.Host, cfg.Port)

	if version == 0 {
		fmt.Println("  Current Version: None (no migrations applied)")
	} else {
		fmt.Printf("  Current Version: %d\n", version)
	}

	if dirty {
		fmt.Println("  Status: ⚠️  DIRTY - Migration failed or was interrupted")
		fmt.Println("  Action: Fix the issue and run 'migrate down' or 'migrate up' to resolve")
	} else {
		fmt.Println("  Status: ✅ CLEAN - All migrations applied successfully")
	}

	return nil
}

func printUsage() {
	fmt.Println("Bank Service Migration Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  migrate <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up       Apply all pending migrations")
	fmt.Println("  down     Roll back all migrations")
	fmt.Println("  steps N  Apply N migration steps (use negative for rollback)")
	fmt.Println("  version  Show current migration version")
	fmt.Println("  status   Show detailed migration status")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  migrate up")
	fmt.Println("  migrate down")
	fmt.Println("  migrate steps 1")
	fmt.Println("  migrate steps -1")
	fmt.Println("  migrate version")
	fmt.Println("  migrate status")
}
