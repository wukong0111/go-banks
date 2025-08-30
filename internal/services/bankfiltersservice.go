package services

import (
	"context"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

// BankFiltersService defines the interface for bank filters operations
type BankFiltersService interface {
	GetAvailableFilters(ctx context.Context) (*models.BankFilters, error)
}

type bankFiltersService struct {
	bankRepo repository.BankRepository
}

func NewBankFiltersService(bankRepo repository.BankRepository) BankFiltersService {
	return &bankFiltersService{
		bankRepo: bankRepo,
	}
}

func (s *bankFiltersService) GetAvailableFilters(ctx context.Context) (*models.BankFilters, error) {
	return s.bankRepo.GetAvailableFilters(ctx)
}
