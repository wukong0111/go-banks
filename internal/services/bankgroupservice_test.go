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

type MockBankGroupRepository struct {
	mock.Mock
}

func (m *MockBankGroupRepository) GetBankGroups(ctx context.Context) ([]models.BankGroup, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.BankGroup), args.Error(1)
}

func TestBankGroupService_GetBankGroups_Success(t *testing.T) {
	mockRepo := new(MockBankGroupRepository)
	service := NewBankGroupService(mockRepo)
	ctx := context.Background()

	// Mock data
	desc := "Test description"
	logo := "https://example.com/logo.png"
	site := "https://example.com"

	expectedGroups := []models.BankGroup{
		{
			GroupID:     uuid.New(),
			Name:        "Test Group 1",
			Description: &desc,
			LogoURL:     &logo,
			Website:     &site,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			GroupID:     uuid.New(),
			Name:        "Test Group 2",
			Description: nil,
			LogoURL:     nil,
			Website:     nil,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	// Setup mock expectations
	mockRepo.On("GetBankGroups", ctx).Return(expectedGroups, nil)

	// Execute
	result, err := service.GetBankGroups(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedGroups, result)
	mockRepo.AssertExpectations(t)
}

func TestBankGroupService_GetBankGroups_EmptyResult(t *testing.T) {
	mockRepo := new(MockBankGroupRepository)
	service := NewBankGroupService(mockRepo)
	ctx := context.Background()

	// Setup mock expectations
	mockRepo.On("GetBankGroups", ctx).Return([]models.BankGroup{}, nil)

	// Execute
	result, err := service.GetBankGroups(ctx)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestBankGroupService_GetBankGroups_RepositoryError(t *testing.T) {
	mockRepo := new(MockBankGroupRepository)
	service := NewBankGroupService(mockRepo)
	ctx := context.Background()

	expectedError := errors.New("database connection failed")

	// Setup mock expectations
	mockRepo.On("GetBankGroups", ctx).Return([]models.BankGroup{}, expectedError)

	// Execute
	result, err := service.GetBankGroups(ctx)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
