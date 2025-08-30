package database

import (
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // PostgreSQL driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"       // File source driver for migrations

	"github.com/wukong0111/go-banks/internal/config"
)

type MigrationService struct {
	migrate *migrate.Migrate
}

func NewMigrationService(cfg *config.DatabaseConfig) (*MigrationService, error) {
	sourceURL := "file://migrations"
	databaseURL := cfg.ConnectionString()

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		slog.Error("failed to create migration service",
			"error", err.Error(),
			"source_url", sourceURL,
			"database_host", cfg.Host,
			"database_port", cfg.Port,
			"database_name", cfg.Name,
		)
		return nil, err
	}

	return &MigrationService{
		migrate: m,
	}, nil
}

func (ms *MigrationService) Up() error {
	currentVersion, dirty, versionErr := ms.migrate.Version()
	
	err := ms.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		slog.Error("migration up failed",
			"error", err.Error(),
			"current_version", func() interface{} {
				if versionErr != nil {
					return "unknown"
				}
				return currentVersion
			}(),
			"dirty", dirty,
		)
		return err
	}
	
	if err == nil {
		newVersion, _, _ := ms.migrate.Version()
		slog.Info("migration up completed successfully",
			"from_version", func() interface{} {
				if versionErr != nil {
					return "unknown"
				}
				return currentVersion
			}(),
			"to_version", newVersion,
		)
	}
	
	return nil
}

func (ms *MigrationService) Down() error {
	currentVersion, dirty, versionErr := ms.migrate.Version()
	
	err := ms.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		slog.Error("migration down failed",
			"error", err.Error(),
			"current_version", func() interface{} {
				if versionErr != nil {
					return "unknown"
				}
				return currentVersion
			}(),
			"dirty", dirty,
		)
		return err
	}
	
	if err == nil {
		newVersion, _, _ := ms.migrate.Version()
		slog.Info("migration down completed successfully",
			"from_version", func() interface{} {
				if versionErr != nil {
					return "unknown"
				}
				return currentVersion
			}(),
			"to_version", newVersion,
		)
	}
	
	return nil
}

func (ms *MigrationService) Steps(n int) error {
	currentVersion, dirty, versionErr := ms.migrate.Version()
	
	err := ms.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		slog.Error("migration steps failed",
			"error", err.Error(),
			"steps", n,
			"current_version", func() interface{} {
				if versionErr != nil {
					return "unknown"
				}
				return currentVersion
			}(),
			"dirty", dirty,
		)
		return err
	}
	
	if err == nil {
		newVersion, _, _ := ms.migrate.Version()
		slog.Info("migration steps completed successfully",
			"steps", n,
			"from_version", func() interface{} {
				if versionErr != nil {
					return "unknown"
				}
				return currentVersion
			}(),
			"to_version", newVersion,
		)
	}
	
	return nil
}

func (ms *MigrationService) Version() (version uint, dirty bool, err error) {
	return ms.migrate.Version()
}

func (ms *MigrationService) Close() error {
	sourceErr, databaseErr := ms.migrate.Close()
	if sourceErr != nil {
		slog.Error("failed to close migration source",
			"error", sourceErr.Error(),
		)
		return sourceErr
	}
	if databaseErr != nil {
		slog.Error("failed to close migration database connection",
			"error", databaseErr.Error(),
		)
		return databaseErr
	}
	
	slog.Debug("migration service closed successfully")
	return nil
}
