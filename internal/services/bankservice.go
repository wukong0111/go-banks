package services

import (
	"context"

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

func (s *BankService) GetBanks(ctx context.Context, filters repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
	// Apply business rules and validation
	normalizedFilters := s.normalizeFilters(filters)

	// Delegate to repository
	return s.bankRepo.GetBanks(ctx, normalizedFilters)
}

// normalizeFilters applies business rules to filter parameters
func (s *BankService) normalizeFilters(filters repository.BankFilters) repository.BankFilters {
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

	return filters
}
