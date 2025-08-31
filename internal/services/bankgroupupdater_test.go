package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/models"
)

type MockBankGroupUpdaterWriter struct {
	mock.Mock
}

func (m *MockBankGroupUpdaterWriter) CreateBankGroup(ctx context.Context, bankGroup *models.BankGroup) error {
	args := m.Called(ctx, bankGroup)
	return args.Error(0)
}

func (m *MockBankGroupUpdaterWriter) UpdateBankGroup(ctx context.Context, bankGroup *models.BankGroup) error {
	args := m.Called(ctx, bankGroup)
	return args.Error(0)
}

type MockBankGroupUpdaterReader struct {
	mock.Mock
}

func (m *MockBankGroupUpdaterReader) GetBankGroups(ctx context.Context) ([]models.BankGroup, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.BankGroup), args.Error(1)
}

func TestBankGroupUpdaterService_UpdateBankGroup_Success(t *testing.T) {
	// Arrange
	mockWriter := new(MockBankGroupUpdaterWriter)
	mockReader := new(MockBankGroupUpdaterReader)
	service := NewBankGroupUpdaterService(mockWriter, mockReader)

	groupID := uuid.New()
	existingGroup := models.BankGroup{
		GroupID:     groupID,
		Name:        "Original Group",
		Description: stringPtr("Original description"),
		LogoURL:     stringPtr("https://example.com/old-logo.png"),
		Website:     stringPtr("https://old.example.com"),
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}

	request := &UpdateBankGroupRequest{
		Name:        stringPtr("Updated Group"),
		Description: stringPtr("Updated description"),
		LogoURL:     stringPtr("https://example.com/new-logo.png"),
	}

	mockReader.On("GetBankGroups", mock.Anything).Return([]models.BankGroup{existingGroup}, nil)
	mockWriter.On("UpdateBankGroup", mock.Anything, mock.MatchedBy(func(bg *models.BankGroup) bool {
		return bg.GroupID == groupID &&
			bg.Name == "Updated Group" &&
			bg.Description != nil && *bg.Description == "Updated description" &&
			bg.LogoURL != nil && *bg.LogoURL == "https://example.com/new-logo.png" &&
			bg.Website != nil && *bg.Website == "https://old.example.com" // Should preserve existing website
	})).Return(nil)

	// Act
	result, err := service.UpdateBankGroup(context.Background(), groupID.String(), request)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, groupID, result.GroupID)
	assert.Equal(t, "Updated Group", result.Name)
	assert.Equal(t, "Updated description", *result.Description)
	assert.Equal(t, "https://example.com/new-logo.png", *result.LogoURL)
	assert.Equal(t, "https://old.example.com", *result.Website) // Should preserve existing
	mockWriter.AssertExpectations(t)
	mockReader.AssertExpectations(t)
}

func TestBankGroupUpdaterService_UpdateBankGroup_PartialUpdate(t *testing.T) {
	// Arrange
	mockWriter := new(MockBankGroupUpdaterWriter)
	mockReader := new(MockBankGroupUpdaterReader)
	service := NewBankGroupUpdaterService(mockWriter, mockReader)

	groupID := uuid.New()
	existingGroup := models.BankGroup{
		GroupID:     groupID,
		Name:        "Original Group",
		Description: stringPtr("Original description"),
		LogoURL:     stringPtr("https://example.com/old-logo.png"),
		Website:     stringPtr("https://old.example.com"),
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}

	// Only update name
	request := &UpdateBankGroupRequest{
		Name: stringPtr("Updated Group Only"),
	}

	mockReader.On("GetBankGroups", mock.Anything).Return([]models.BankGroup{existingGroup}, nil)
	mockWriter.On("UpdateBankGroup", mock.Anything, mock.MatchedBy(func(bg *models.BankGroup) bool {
		return bg.GroupID == groupID &&
			bg.Name == "Updated Group Only" &&
			bg.Description != nil && *bg.Description == "Original description" && // Should preserve
			bg.LogoURL != nil && *bg.LogoURL == "https://example.com/old-logo.png" && // Should preserve
			bg.Website != nil && *bg.Website == "https://old.example.com" // Should preserve
	})).Return(nil)

	// Act
	result, err := service.UpdateBankGroup(context.Background(), groupID.String(), request)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Group Only", result.Name)
	assert.Equal(t, "Original description", *result.Description) // Preserved
	mockWriter.AssertExpectations(t)
	mockReader.AssertExpectations(t)
}

func TestBankGroupUpdaterService_UpdateBankGroup_InvalidGroupID(t *testing.T) {
	// Arrange
	mockWriter := new(MockBankGroupUpdaterWriter)
	mockReader := new(MockBankGroupUpdaterReader)
	service := NewBankGroupUpdaterService(mockWriter, mockReader)

	request := &UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	// Act & Assert
	result, err := service.UpdateBankGroup(context.Background(), "invalid-uuid", request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid group_id")
	assert.Contains(t, err.Error(), "invalid UUID format")
}

func TestBankGroupUpdaterService_UpdateBankGroup_EmptyGroupID(t *testing.T) {
	// Arrange
	mockWriter := new(MockBankGroupUpdaterWriter)
	mockReader := new(MockBankGroupUpdaterReader)
	service := NewBankGroupUpdaterService(mockWriter, mockReader)

	request := &UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	// Act & Assert
	result, err := service.UpdateBankGroup(context.Background(), "", request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid group_id")
	assert.Contains(t, err.Error(), "group_id cannot be empty")
}

func TestBankGroupUpdaterService_UpdateBankGroup_GroupNotFound(t *testing.T) {
	// Arrange
	mockWriter := new(MockBankGroupUpdaterWriter)
	mockReader := new(MockBankGroupUpdaterReader)
	service := NewBankGroupUpdaterService(mockWriter, mockReader)

	groupID := uuid.New()
	request := &UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	// Return empty list (group not found)
	mockReader.On("GetBankGroups", mock.Anything).Return([]models.BankGroup{}, nil)

	// Act
	result, err := service.UpdateBankGroup(context.Background(), groupID.String(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "bank group not found")
	mockReader.AssertExpectations(t)
}

func TestBankGroupUpdaterService_UpdateBankGroup_EmptyName(t *testing.T) {
	// Arrange
	mockWriter := new(MockBankGroupUpdaterWriter)
	mockReader := new(MockBankGroupUpdaterReader)
	service := NewBankGroupUpdaterService(mockWriter, mockReader)

	groupID := uuid.New()
	existingGroup := models.BankGroup{
		GroupID:     groupID,
		Name:        "Original Group",
		Description: stringPtr("Original description"),
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}

	request := &UpdateBankGroupRequest{
		Name: stringPtr("   "), // Whitespace only
	}

	mockReader.On("GetBankGroups", mock.Anything).Return([]models.BankGroup{existingGroup}, nil)

	// Act
	result, err := service.UpdateBankGroup(context.Background(), groupID.String(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "name cannot be empty")
	mockReader.AssertExpectations(t)
}

func TestBankGroupUpdaterService_UpdateBankGroup_ReaderError(t *testing.T) {
	// Arrange
	mockWriter := new(MockBankGroupUpdaterWriter)
	mockReader := new(MockBankGroupUpdaterReader)
	service := NewBankGroupUpdaterService(mockWriter, mockReader)

	groupID := uuid.New()
	request := &UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	mockReader.On("GetBankGroups", mock.Anything).Return([]models.BankGroup{}, errors.New("database connection failed"))

	// Act
	result, err := service.UpdateBankGroup(context.Background(), groupID.String(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch bank groups")
	mockReader.AssertExpectations(t)
}

func TestBankGroupUpdaterService_UpdateBankGroup_WriterError(t *testing.T) {
	// Arrange
	mockWriter := new(MockBankGroupUpdaterWriter)
	mockReader := new(MockBankGroupUpdaterReader)
	service := NewBankGroupUpdaterService(mockWriter, mockReader)

	groupID := uuid.New()
	existingGroup := models.BankGroup{
		GroupID:   groupID,
		Name:      "Original Group",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}

	request := &UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	mockReader.On("GetBankGroups", mock.Anything).Return([]models.BankGroup{existingGroup}, nil)
	mockWriter.On("UpdateBankGroup", mock.Anything, mock.Anything).Return(errors.New("database write failed"))

	// Act
	result, err := service.UpdateBankGroup(context.Background(), groupID.String(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update bank group")
	mockWriter.AssertExpectations(t)
	mockReader.AssertExpectations(t)
}
