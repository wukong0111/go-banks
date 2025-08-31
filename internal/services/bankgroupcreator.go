package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

// CreateBankGroupRequest represents the request to create a bank group
type CreateBankGroupRequest struct {
	GroupID     string  `json:"group_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	Website     *string `json:"website,omitempty"`
}

// BankGroupCreator defines the interface for creating bank groups
type BankGroupCreator interface {
	CreateBankGroup(ctx context.Context, request *CreateBankGroupRequest) (*models.BankGroup, error)
}

// BankGroupCreatorService implements BankGroupCreator
type BankGroupCreatorService struct {
	writer repository.BankGroupWriter
}

// NewBankGroupCreatorService creates a new BankGroupCreatorService
func NewBankGroupCreatorService(writer repository.BankGroupWriter) BankGroupCreator {
	return &BankGroupCreatorService{
		writer: writer,
	}
}

// CreateBankGroup creates a new bank group with validation and business logic
func (s *BankGroupCreatorService) CreateBankGroup(ctx context.Context, request *CreateBankGroupRequest) (*models.BankGroup, error) {
	// Validate and parse UUID
	groupUUID, err := s.validateGroupID(request.GroupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group_id: %w", err)
	}

	// Validate name is not empty after trimming
	name := strings.TrimSpace(request.Name)
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	// Build bank group model
	bankGroup := &models.BankGroup{
		GroupID:     *groupUUID,
		Name:        name,
		Description: s.trimStringPtr(request.Description),
		LogoURL:     s.trimStringPtr(request.LogoURL),
		Website:     s.trimStringPtr(request.Website),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create bank group in repository
	if err := s.writer.CreateBankGroup(ctx, bankGroup); err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("bank group with ID '%s' already exists", request.GroupID)
		}
		return nil, fmt.Errorf("failed to create bank group: %w", err)
	}

	return bankGroup, nil
}

// validateGroupID validates and parses the group ID as UUID
func (s *BankGroupCreatorService) validateGroupID(groupID string) (*uuid.UUID, error) {
	trimmedID := strings.TrimSpace(groupID)
	if trimmedID == "" {
		return nil, errors.New("group_id cannot be empty")
	}

	parsedUUID, err := uuid.Parse(trimmedID)
	if err != nil {
		return nil, fmt.Errorf("group_id must be a valid UUID: %w", err)
	}

	return &parsedUUID, nil
}

// trimStringPtr trims whitespace from string pointer and returns nil if empty
func (s *BankGroupCreatorService) trimStringPtr(str *string) *string {
	if str == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*str)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
