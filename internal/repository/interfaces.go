package repository

import (
	"context"

	"github.com/wukong0111/go-banks/internal/models"
)

// BankFilters represents the search criteria for banks (domain boundary)
type BankFilters struct {
	Environment string
	Name        string
	API         string
	Country     string
	Page        int
	Limit       int
}

// BankRepository defines the methods that a bank repository must implement
type BankRepository interface {
	GetBanks(ctx context.Context, filters *BankFilters) ([]models.Bank, *models.Pagination, error)
	GetBankByID(ctx context.Context, bankID string) (*models.Bank, error)
	GetBankEnvironmentConfigs(ctx context.Context, bankID string, environment string) (map[string]*models.BankEnvironmentConfig, error)
}
