package services

import (
	"context"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

type BankService struct {
	bankRepo *repository.BankRepository
}

func NewBankService(bankRepo *repository.BankRepository) *BankService {
	return &BankService{
		bankRepo: bankRepo,
	}
}

func (s *BankService) GetBanks(ctx context.Context, filters repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
	return s.bankRepo.GetBanks(ctx, filters)
}
