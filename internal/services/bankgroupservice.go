package services

import (
	"context"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

// BankGroupService defines the interface for bank group related operations.
type BankGroupService interface {
	GetBankGroups(ctx context.Context) ([]models.BankGroup, error)
}

type bankGroupService struct {
	bankGroupRepo repository.BankGroupRepository
}

func NewBankGroupService(bankGroupRepo repository.BankGroupRepository) BankGroupService {
	return &bankGroupService{
		bankGroupRepo: bankGroupRepo,
	}
}

func (s *bankGroupService) GetBankGroups(ctx context.Context) ([]models.BankGroup, error) {
	return s.bankGroupRepo.GetBankGroups(ctx)
}
