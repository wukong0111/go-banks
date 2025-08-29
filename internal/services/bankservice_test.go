package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

// MockBankRepository implements the BankRepository interface for testing with testify/mock
type MockBankRepository struct {
	mock.Mock
}

func (m *MockBankRepository) GetBanks(ctx context.Context, filters *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
	args := m.Called(ctx, filters)
	var banks []models.Bank
	var pagination *models.Pagination

	if args.Get(0) != nil {
		banks = args.Get(0).([]models.Bank)
	}
	if args.Get(1) != nil {
		pagination = args.Get(1).(*models.Pagination)
	}

	return banks, pagination, args.Error(2)
}

func (m *MockBankRepository) GetBankByID(ctx context.Context, bankID string) (*models.Bank, error) {
	args := m.Called(ctx, bankID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bank), args.Error(1)
}

func (m *MockBankRepository) GetBankEnvironmentConfigs(ctx context.Context, bankID, environment string) (map[string]*models.BankEnvironmentConfig, error) {
	args := m.Called(ctx, bankID, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*models.BankEnvironmentConfig), args.Error(1)
}

func TestBankService_GetBanks(t *testing.T) {
	// Test data
	expectedBanks := []models.Bank{
		{
			BankID:  "test-bank-1",
			Name:    "Test Bank 1",
			Country: "ES",
		},
		{
			BankID:  "test-bank-2",
			Name:    "Test Bank 2",
			Country: "FR",
		},
	}
	expectedPagination := &models.Pagination{
		Page:       1,
		Limit:      20,
		Total:      2,
		TotalPages: 1,
	}

	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBanks", mock.Anything, mock.Anything).Return(expectedBanks, expectedPagination, nil)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Test filters
	filters := &repository.BankFilters{
		Environment: "sandbox",
		Page:        1,
		Limit:       20,
	}

	// Call the method
	banks, pagination, err := service.GetBanks(context.Background(), filters)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, banks, len(expectedBanks))
	assert.Equal(t, expectedBanks[0].BankID, banks[0].BankID)
	assert.Equal(t, expectedPagination.Total, pagination.Total)

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBanks_RepositoryError(t *testing.T) {
	// Create mock repository that returns error
	mockRepo := new(MockBankRepository)
	expectedErr := errors.New("database connection failed")
	mockRepo.On("GetBanks", mock.Anything, mock.Anything).Return(nil, nil, expectedErr)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Test filters
	filters := &repository.BankFilters{
		Environment: "sandbox",
		Page:        1,
		Limit:       20,
	}

	// Call the method
	banks, pagination, err := service.GetBanks(context.Background(), filters)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, banks)
	assert.Nil(t, pagination)
	assert.Contains(t, err.Error(), "database connection failed")

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBanks_EmptyResults(t *testing.T) {
	// Test data - empty results
	expectedBanks := []models.Bank{}
	expectedPagination := &models.Pagination{
		Page:       1,
		Limit:      20,
		Total:      0,
		TotalPages: 0,
	}

	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBanks", mock.Anything, mock.Anything).Return(expectedBanks, expectedPagination, nil)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Test filters
	filters := &repository.BankFilters{
		Environment: "production",
		Page:        1,
		Limit:       20,
	}

	// Call the method
	banks, pagination, err := service.GetBanks(context.Background(), filters)

	// Assertions
	require.NoError(t, err)
	assert.Empty(t, banks)
	assert.Equal(t, 0, pagination.Total)
	assert.Equal(t, 0, pagination.TotalPages)

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBanks_FilterValidation(t *testing.T) {
	// Test data
	expectedBanks := []models.Bank{
		{
			BankID:  "test-bank-1",
			Name:    "Test Bank",
			Country: "ES",
		},
	}
	expectedPagination := &models.Pagination{
		Page:       1,
		Limit:      20,
		Total:      1,
		TotalPages: 1,
	}

	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBanks", mock.Anything, mock.MatchedBy(func(filters *repository.BankFilters) bool {
		return filters.Name == "Test Bank" &&
			filters.Environment == "production" &&
			filters.Country == "ES"
	})).Return(expectedBanks, expectedPagination, nil)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Test filters with specific values
	filters := &repository.BankFilters{
		Name:        "Test Bank",
		Environment: "production",
		Country:     "ES",
		Page:        1,
		Limit:       20,
	}

	// Call the method
	banks, pagination, err := service.GetBanks(context.Background(), filters)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, banks, 1)
	assert.Equal(t, "test-bank-1", banks[0].BankID)
	assert.Equal(t, 1, pagination.Total)

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBanks_BusinessRulesValidation(t *testing.T) {
	// Test data
	expectedBanks := []models.Bank{
		{BankID: "bank-1", Name: "Bank 1", Country: "ES"},
		{BankID: "bank-2", Name: "Bank 2", Country: "FR"},
	}
	expectedPagination := &models.Pagination{
		Page:       2,
		Limit:      10,
		Total:      15,
		TotalPages: 2,
	}

	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBanks", mock.Anything, mock.MatchedBy(func(filters *repository.BankFilters) bool {
		return filters.Page == 2 && filters.Limit == 10
	})).Return(expectedBanks, expectedPagination, nil)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Test filters with page 2
	filters := &repository.BankFilters{
		Page:  2,
		Limit: 10,
	}

	// Call the method
	banks, pagination, err := service.GetBanks(context.Background(), filters)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, banks, 2)
	assert.Equal(t, 2, pagination.Page)
	assert.Equal(t, 15, pagination.Total)

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBanks_MaxLimitEnforcement(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBanks", mock.Anything, mock.MatchedBy(func(filters *repository.BankFilters) bool {
		// Service should enforce max limit of 100
		return filters.Limit <= 100
	})).Return([]models.Bank{}, &models.Pagination{Page: 1, Limit: 20, Total: 0, TotalPages: 0}, nil)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Test with large limit
	filters := &repository.BankFilters{
		Page:  1,
		Limit: 1000, // This should be capped to 100 by business rules
	}

	// Call the method
	banks, pagination, err := service.GetBanks(context.Background(), filters)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, banks)
	assert.NotNil(t, pagination)

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBankDetails_AllEnvironments(t *testing.T) {
	// Test data
	expectedBank := &models.Bank{
		BankID:  "test-bank",
		Name:    "Test Bank",
		Country: "ES",
	}
	expectedConfigs := map[string]*models.BankEnvironmentConfig{
		"sandbox": {
			BankID:      "test-bank",
			Environment: models.EnvironmentSandbox,
			Enabled:     true,
		},
		"production": {
			BankID:      "test-bank",
			Environment: models.EnvironmentProduction,
			Enabled:     true,
		},
	}

	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBankByID", mock.Anything, "test-bank").Return(expectedBank, nil)
	mockRepo.On("GetBankEnvironmentConfigs", mock.Anything, "test-bank", "").Return(expectedConfigs, nil)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Call the method - no environment specified (get all)
	bankDetails, err := service.GetBankDetails(context.Background(), "test-bank", "")

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, bankDetails)
	bankWithEnvs := bankDetails.(*models.BankWithEnvironments)
	assert.Equal(t, expectedBank.BankID, bankWithEnvs.BankID)
	assert.Len(t, bankWithEnvs.EnvironmentConfigs, 2)
	assert.Contains(t, bankWithEnvs.EnvironmentConfigs, "sandbox")
	assert.Contains(t, bankWithEnvs.EnvironmentConfigs, "production")

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBankDetails_SpecificEnvironment(t *testing.T) {
	// Test data
	expectedBank := &models.Bank{
		BankID:  "test-bank",
		Name:    "Test Bank",
		Country: "ES",
	}
	expectedConfig := &models.BankEnvironmentConfig{
		BankID:      "test-bank",
		Environment: models.EnvironmentSandbox,
		Enabled:     true,
	}
	expectedConfigs := map[string]*models.BankEnvironmentConfig{
		"sandbox": expectedConfig,
	}

	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBankByID", mock.Anything, "test-bank").Return(expectedBank, nil)
	mockRepo.On("GetBankEnvironmentConfigs", mock.Anything, "test-bank", "sandbox").Return(expectedConfigs, nil)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Call the method - specific environment
	bankDetails, err := service.GetBankDetails(context.Background(), "test-bank", "sandbox")

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, bankDetails)
	bankWithEnv := bankDetails.(*models.BankWithEnvironment)
	assert.Equal(t, expectedBank.BankID, bankWithEnv.BankID)
	assert.NotNil(t, bankWithEnv.EnvironmentConfig)
	assert.Equal(t, expectedConfig.Environment, bankWithEnv.EnvironmentConfig.Environment)
	assert.Equal(t, expectedConfig.Enabled, bankWithEnv.EnvironmentConfig.Enabled)

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBankDetails_BankNotFound(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBankByID", mock.Anything, "nonexistent-bank").Return(nil, errors.New("bank not found"))

	// Create service with mock
	service := NewBankService(mockRepo)

	// Call the method
	bankDetails, err := service.GetBankDetails(context.Background(), "nonexistent-bank", "")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, bankDetails)
	assert.Contains(t, err.Error(), "bank not found")

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBankDetails_InvalidEnvironment(t *testing.T) {
	// Create mock repository (no expectations because service validates environment first)
	mockRepo := new(MockBankRepository)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Call the method with invalid environment
	bankDetails, err := service.GetBankDetails(context.Background(), "test-bank", "invalid-env")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, bankDetails)
	assert.Contains(t, err.Error(), "invalid environment: invalid-env")

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBankDetails_ValidEnvNoConfig(t *testing.T) {
	// Test data
	expectedBank := &models.Bank{
		BankID:  "test-bank",
		Name:    "Test Bank",
		Country: "ES",
	}
	emptyConfigs := map[string]*models.BankEnvironmentConfig{}

	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBankByID", mock.Anything, "test-bank").Return(expectedBank, nil)
	mockRepo.On("GetBankEnvironmentConfigs", mock.Anything, "test-bank", "sandbox").Return(emptyConfigs, nil)

	// Create service with mock
	service := NewBankService(mockRepo)

	// Call the method - should return error when environment config not found
	bankDetails, err := service.GetBankDetails(context.Background(), "test-bank", "sandbox")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, bankDetails)
	assert.Contains(t, err.Error(), "environment configuration not found for sandbox")

	mockRepo.AssertExpectations(t)
}

func TestBankService_GetBankDetails_RepositoryError(t *testing.T) {
	// Test data
	expectedBank := &models.Bank{
		BankID:  "test-bank",
		Name:    "Test Bank",
		Country: "ES",
	}

	// Create mock repository
	mockRepo := new(MockBankRepository)
	mockRepo.On("GetBankByID", mock.Anything, "test-bank").Return(expectedBank, nil)
	mockRepo.On("GetBankEnvironmentConfigs", mock.Anything, "test-bank", "").Return(nil, errors.New("config fetch failed"))

	// Create service with mock
	service := NewBankService(mockRepo)

	// Call the method
	bankDetails, err := service.GetBankDetails(context.Background(), "test-bank", "")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, bankDetails)
	assert.Contains(t, err.Error(), "config fetch failed")

	mockRepo.AssertExpectations(t)
}
