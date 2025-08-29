package database

import (
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
		return nil, err
	}

	return &MigrationService{
		migrate: m,
	}, nil
}

func (ms *MigrationService) Up() error {
	err := ms.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (ms *MigrationService) Down() error {
	err := ms.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (ms *MigrationService) Steps(n int) error {
	err := ms.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (ms *MigrationService) Version() (version uint, dirty bool, err error) {
	return ms.migrate.Version()
}

func (ms *MigrationService) Close() error {
	sourceErr, databaseErr := ms.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	if databaseErr != nil {
		return databaseErr
	}
	return nil
}
