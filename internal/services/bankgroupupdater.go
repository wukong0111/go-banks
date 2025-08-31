package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

type UpdateBankGroupRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	Website     *string `json:"website,omitempty"`
}

type BankGroupUpdater interface {
	UpdateBankGroup(ctx context.Context, groupID string, request *UpdateBankGroupRequest) (*models.BankGroup, error)
}

type BankGroupUpdaterService struct {
	writer repository.BankGroupWriter
	reader repository.BankGroupRepository
}

func NewBankGroupUpdaterService(writer repository.BankGroupWriter, reader repository.BankGroupRepository) *BankGroupUpdaterService {
	return &BankGroupUpdaterService{
		writer: writer,
		reader: reader,
	}
}

func (s *BankGroupUpdaterService) UpdateBankGroup(ctx context.Context, groupID string, request *UpdateBankGroupRequest) (*models.BankGroup, error) {
	// Parse and validate group ID
	groupUUID, err := s.validateGroupID(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group_id: %w", err)
	}

	// Fetch existing bank group
	existingGroups, err := s.reader.GetBankGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bank groups: %w", err)
	}

	var existingGroup *models.BankGroup
	for i := range existingGroups {
		if existingGroups[i].GroupID == *groupUUID {
			existingGroup = &existingGroups[i]
			break
		}
	}

	if existingGroup == nil {
		return nil, errors.New("bank group not found")
	}

	// Build updated bank group from request
	updatedGroup, err := s.requestToBankGroup(existingGroup, request)
	if err != nil {
		return nil, err
	}

	// Update bank group in database
	if err := s.writer.UpdateBankGroup(ctx, updatedGroup); err != nil {
		return nil, fmt.Errorf("failed to update bank group: %w", err)
	}

	return updatedGroup, nil
}

func (s *BankGroupUpdaterService) validateGroupID(groupID string) (*uuid.UUID, error) {
	trimmedID := strings.TrimSpace(groupID)
	if trimmedID == "" {
		return nil, errors.New("group_id cannot be empty")
	}

	parsed, err := uuid.Parse(trimmedID)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	return &parsed, nil
}

func (s *BankGroupUpdaterService) requestToBankGroup(existing *models.BankGroup, request *UpdateBankGroupRequest) (*models.BankGroup, error) {
	// Start with existing bank group data
	updated := &models.BankGroup{
		GroupID:     existing.GroupID,
		Name:        existing.Name,
		Description: existing.Description,
		LogoURL:     existing.LogoURL,
		Website:     existing.Website,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   existing.UpdatedAt, // Will be updated by database
	}

	// Apply updates only for provided fields
	if request.Name != nil {
		trimmedName := strings.TrimSpace(*request.Name)
		if trimmedName == "" {
			return nil, errors.New("name cannot be empty")
		}
		updated.Name = trimmedName
	}

	if request.Description != nil {
		updated.Description = request.Description
	}

	if request.LogoURL != nil {
		updated.LogoURL = request.LogoURL
	}

	if request.Website != nil {
		updated.Website = request.Website
	}

	return updated, nil
}
