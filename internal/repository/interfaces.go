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
	GetAvailableFilters(ctx context.Context) (*models.BankFilters, error)
}

// BankWriter defines the methods for creating and updating banks
type BankWriter interface {
	CreateBank(ctx context.Context, bank *models.Bank) error
	CreateBankWithEnvironments(ctx context.Context, bank *models.Bank, configs []*models.BankEnvironmentConfig) error
	UpdateBank(ctx context.Context, bank *models.Bank) error
	UpdateBankWithEnvironments(ctx context.Context, bank *models.Bank, configs []*models.BankEnvironmentConfig) error
}

// BankGroupRepository defines the methods that a bank group repository must implement
type BankGroupRepository interface {
	GetBankGroups(ctx context.Context) ([]models.BankGroup, error)
}

// BankGroupWriter defines the methods for creating and updating bank groups
type BankGroupWriter interface {
	CreateBankGroup(ctx context.Context, bankGroup *models.BankGroup) error
}
