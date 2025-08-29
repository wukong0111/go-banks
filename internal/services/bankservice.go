package services

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

type BankService struct {
	bankRepo repository.BankRepository
}

func NewBankService(bankRepo repository.BankRepository) *BankService {
	return &BankService{
		bankRepo: bankRepo,
	}
}

func (s *BankService) GetBanks(ctx context.Context, filters *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
	// Apply business rules and validation
	s.normalizeFilters(filters)

	// Delegate to repository
	return s.bankRepo.GetBanks(ctx, filters)
}

// normalizeFilters applies business rules to filter parameters
func (s *BankService) normalizeFilters(filters *repository.BankFilters) {
	const (
		DefaultPage        = 1
		DefaultLimit       = 20
		DefaultEnvironment = "all"
		MaxLimit           = 100
		MinLimit           = 1
	)

	// Normalize environment - default to "all" if empty
	if filters.Environment == "" {
		filters.Environment = DefaultEnvironment
	}

	// Normalize page - must be at least 1
	if filters.Page < 1 {
		filters.Page = DefaultPage
	}

	// Normalize limit - apply defaults and max limits
	if filters.Limit < MinLimit {
		filters.Limit = DefaultLimit
	} else if filters.Limit > MaxLimit {
		filters.Limit = MaxLimit
	}
}

func (s *BankService) GetBankDetails(ctx context.Context, bankID, environment string) (models.BankDetails, error) {
	// If specific environment is requested, validate it first
	if environment != "" {
		if !s.isValidEnvironment(environment) {
			return nil, fmt.Errorf("invalid environment: %s", environment)
		}
	}

	// Get the bank by ID first
	bank, err := s.bankRepo.GetBankByID(ctx, bankID)
	if err != nil {
		if err == pgx.ErrNoRows || strings.Contains(err.Error(), "no rows in result set") {
			return nil, errors.New("bank not found")
		}
		return nil, fmt.Errorf("failed to get bank: %w", err)
	}

	// Get environment configurations
	envConfigs, err := s.bankRepo.GetBankEnvironmentConfigs(ctx, bankID, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment configs: %w", err)
	}

	// If specific environment is requested
	if environment != "" {

		config, exists := envConfigs[environment]
		if !exists {
			return nil, fmt.Errorf("environment configuration not found for %s", environment)
		}

		return &models.BankWithEnvironment{
			Bank:              *bank,
			EnvironmentConfig: config,
		}, nil
	}

	// Return all environments
	return &models.BankWithEnvironments{
		Bank:               *bank,
		EnvironmentConfigs: envConfigs,
	}, nil
}

// isValidEnvironment validates if the provided environment is valid
func (s *BankService) isValidEnvironment(env string) bool {
	validEnvironments := []string{"sandbox", "production", "uat", "test"}
	return slices.Contains(validEnvironments, env)
}
