package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

// MockBankRepository implements the BankRepository interface for testing
type MockBankRepository struct {
	GetBanksFunc                  func(ctx context.Context, filters *repository.BankFilters) ([]models.Bank, *models.Pagination, error)
	GetBankByIDFunc               func(ctx context.Context, bankID string) (*models.Bank, error)
	GetBankEnvironmentConfigsFunc func(ctx context.Context, bankID string, environment string) (map[string]*models.BankEnvironmentConfig, error)
}

func (m *MockBankRepository) GetBanks(ctx context.Context, filters *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
	if m.GetBanksFunc != nil {
		return m.GetBanksFunc(ctx, filters)
	}
	return nil, nil, nil
}

func (m *MockBankRepository) GetBankByID(ctx context.Context, bankID string) (*models.Bank, error) {
	if m.GetBankByIDFunc != nil {
		return m.GetBankByIDFunc(ctx, bankID)
	}
	return nil, nil
}

func (m *MockBankRepository) GetBankEnvironmentConfigs(ctx context.Context, bankID, environment string) (map[string]*models.BankEnvironmentConfig, error) {
	if m.GetBankEnvironmentConfigsFunc != nil {
		return m.GetBankEnvironmentConfigsFunc(ctx, bankID, environment)
	}
	return nil, nil
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
	mockRepo := &MockBankRepository{
		GetBanksFunc: func(_ context.Context, _ *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
			return expectedBanks, expectedPagination, nil
		},
	}

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
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(banks) != len(expectedBanks) {
		t.Errorf("Expected %d banks, got %d", len(expectedBanks), len(banks))
	}

	if banks[0].BankID != expectedBanks[0].BankID {
		t.Errorf("Expected bank ID %s, got %s", expectedBanks[0].BankID, banks[0].BankID)
	}

	if pagination.Total != expectedPagination.Total {
		t.Errorf("Expected total %d, got %d", expectedPagination.Total, pagination.Total)
	}
}

func TestBankService_GetBanks_RepositoryError(t *testing.T) {
	// Create mock repository that returns error
	expectedError := errors.New("database connection failed")
	mockRepo := &MockBankRepository{
		GetBanksFunc: func(_ context.Context, _ *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
			return nil, nil, expectedError
		},
	}

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
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}

	if banks != nil {
		t.Errorf("Expected nil banks, got %v", banks)
	}

	if pagination != nil {
		t.Errorf("Expected nil pagination, got %v", pagination)
	}
}

func TestBankService_GetBanks_EmptyResults(t *testing.T) {
	// Create mock repository that returns empty results
	expectedPagination := &models.Pagination{
		Page:       1,
		Limit:      20,
		Total:      0,
		TotalPages: 0,
	}

	mockRepo := &MockBankRepository{
		GetBanksFunc: func(_ context.Context, _ *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
			return []models.Bank{}, expectedPagination, nil
		},
	}

	// Create service with mock
	service := NewBankService(mockRepo)

	// Test filters
	filters := &repository.BankFilters{
		Environment: "production",
		Country:     "XX", // Non-existent country
		Page:        1,
		Limit:       20,
	}

	// Call the method
	banks, pagination, err := service.GetBanks(context.Background(), filters)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(banks) != 0 {
		t.Errorf("Expected 0 banks, got %d", len(banks))
	}

	if pagination.Total != 0 {
		t.Errorf("Expected total 0, got %d", pagination.Total)
	}
}

func TestBankService_GetBanks_FilterValidation(t *testing.T) {
	// Track what filters were passed to repository
	var receivedFilters *repository.BankFilters

	mockRepo := &MockBankRepository{
		GetBanksFunc: func(_ context.Context, filters *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
			receivedFilters = filters
			return []models.Bank{}, &models.Pagination{}, nil
		},
	}

	// Create service with mock
	service := NewBankService(mockRepo)

	// Test filters with specific values
	expectedFilters := &repository.BankFilters{
		Environment: "uat",
		Name:        "Santander",
		API:         "OpenBanking",
		Country:     "ES",
		Page:        2,
		Limit:       50,
	}

	// Call the method
	_, _, err := service.GetBanks(context.Background(), expectedFilters)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify all filters were passed correctly
	if receivedFilters.Environment != expectedFilters.Environment {
		t.Errorf("Expected environment %s, got %s", expectedFilters.Environment, receivedFilters.Environment)
	}

	if receivedFilters.Name != expectedFilters.Name {
		t.Errorf("Expected name %s, got %s", expectedFilters.Name, receivedFilters.Name)
	}

	if receivedFilters.API != expectedFilters.API {
		t.Errorf("Expected API %s, got %s", expectedFilters.API, receivedFilters.API)
	}

	if receivedFilters.Country != expectedFilters.Country {
		t.Errorf("Expected country %s, got %s", expectedFilters.Country, receivedFilters.Country)
	}

	if receivedFilters.Page != expectedFilters.Page {
		t.Errorf("Expected page %d, got %d", expectedFilters.Page, receivedFilters.Page)
	}

	if receivedFilters.Limit != expectedFilters.Limit {
		t.Errorf("Expected limit %d, got %d", expectedFilters.Limit, receivedFilters.Limit)
	}
}

func TestBankService_GetBanks_BusinessRulesValidation(t *testing.T) {
	// Track what filters were passed to repository after normalization
	var receivedFilters *repository.BankFilters

	mockRepo := &MockBankRepository{
		GetBanksFunc: func(_ context.Context, filters *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
			receivedFilters = filters
			return []models.Bank{}, &models.Pagination{}, nil
		},
	}

	service := NewBankService(mockRepo)

	// Test with invalid/empty values that should be normalized
	inputFilters := &repository.BankFilters{
		Environment: "", // Should become "all"
		Page:        0,  // Should become 1
		Limit:       0,  // Should become 20
	}

	_, _, err := service.GetBanks(context.Background(), inputFilters)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify business rules were applied
	if receivedFilters.Environment != "all" {
		t.Errorf("Expected environment 'all', got '%s'", receivedFilters.Environment)
	}

	if receivedFilters.Page != 1 {
		t.Errorf("Expected page 1, got %d", receivedFilters.Page)
	}

	if receivedFilters.Limit != 20 {
		t.Errorf("Expected limit 20, got %d", receivedFilters.Limit)
	}
}

func TestBankService_GetBanks_MaxLimitEnforcement(t *testing.T) {
	var receivedFilters *repository.BankFilters

	mockRepo := &MockBankRepository{
		GetBanksFunc: func(_ context.Context, filters *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
			receivedFilters = filters
			return []models.Bank{}, &models.Pagination{}, nil
		},
	}

	service := NewBankService(mockRepo)

	// Test with limit exceeding maximum
	inputFilters := &repository.BankFilters{
		Limit: 150, // Should be capped to 100
	}

	_, _, err := service.GetBanks(context.Background(), inputFilters)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify max limit was enforced
	if receivedFilters.Limit != 100 {
		t.Errorf("Expected limit capped to 100, got %d", receivedFilters.Limit)
	}
}

func TestBankService_GetBankDetails_AllEnvironments(t *testing.T) {
	// Test data
	expectedBank := &models.Bank{
		BankID:  "TEST001",
		Name:    "Test Bank",
		Country: "ES",
	}

	expectedConfigs := map[string]*models.BankEnvironmentConfig{
		"production": {
			BankID:      "TEST001",
			Environment: models.EnvironmentProduction,
			Enabled:     true,
		},
		"sandbox": {
			BankID:      "TEST001",
			Environment: models.EnvironmentSandbox,
			Enabled:     true,
		},
	}

	mockRepo := &MockBankRepository{
		GetBankByIDFunc: func(_ context.Context, bankID string) (*models.Bank, error) {
			if bankID == "TEST001" {
				return expectedBank, nil
			}
			return nil, errors.New("bank not found")
		},
		GetBankEnvironmentConfigsFunc: func(_ context.Context, bankID string, environment string) (map[string]*models.BankEnvironmentConfig, error) {
			if bankID == "TEST001" && environment == "" {
				return expectedConfigs, nil
			}
			return nil, errors.New("config not found")
		},
	}

	service := NewBankService(mockRepo)

	result, err := service.GetBankDetails(context.Background(), "TEST001", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	bankWithEnvs, ok := result.(*models.BankWithEnvironments)
	if !ok {
		t.Fatalf("Expected BankWithEnvironments, got %T", result)
	}

	if bankWithEnvs.BankID != expectedBank.BankID {
		t.Errorf("Expected bank ID %s, got %s", expectedBank.BankID, bankWithEnvs.BankID)
	}

	if len(bankWithEnvs.EnvironmentConfigs) != 2 {
		t.Errorf("Expected 2 environment configs, got %d", len(bankWithEnvs.EnvironmentConfigs))
	}
}

func TestBankService_GetBankDetails_SpecificEnvironment(t *testing.T) {
	expectedBank := &models.Bank{
		BankID:  "TEST001",
		Name:    "Test Bank",
		Country: "ES",
	}

	expectedConfig := &models.BankEnvironmentConfig{
		BankID:      "TEST001",
		Environment: models.EnvironmentProduction,
		Enabled:     true,
	}

	expectedConfigs := map[string]*models.BankEnvironmentConfig{
		"production": expectedConfig,
	}

	mockRepo := &MockBankRepository{
		GetBankByIDFunc: func(_ context.Context, _ string) (*models.Bank, error) {
			return expectedBank, nil
		},
		GetBankEnvironmentConfigsFunc: func(_ context.Context, _ string, environment string) (map[string]*models.BankEnvironmentConfig, error) {
			if environment == "production" {
				return expectedConfigs, nil
			}
			return nil, errors.New("config not found")
		},
	}

	service := NewBankService(mockRepo)

	result, err := service.GetBankDetails(context.Background(), "TEST001", "production")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	bankWithEnv, ok := result.(*models.BankWithEnvironment)
	if !ok {
		t.Fatalf("Expected BankWithEnvironment, got %T", result)
	}

	if bankWithEnv.BankID != expectedBank.BankID {
		t.Errorf("Expected bank ID %s, got %s", expectedBank.BankID, bankWithEnv.BankID)
	}

	if bankWithEnv.EnvironmentConfig.Environment != models.EnvironmentProduction {
		t.Errorf("Expected environment production, got %s", bankWithEnv.EnvironmentConfig.Environment)
	}
}

func TestBankService_GetBankDetails_BankNotFound(t *testing.T) {
	mockRepo := &MockBankRepository{
		GetBankByIDFunc: func(_ context.Context, _ string) (*models.Bank, error) {
			return nil, errors.New("no rows in result set")
		},
	}

	service := NewBankService(mockRepo)

	_, err := service.GetBankDetails(context.Background(), "NONEXISTENT", "")
	if err == nil {
		t.Fatal("Expected error for non-existent bank")
	}

	if !strings.Contains(err.Error(), "bank not found") {
		t.Errorf("Expected 'bank not found' error, got %v", err)
	}
}

func TestBankService_GetBankDetails_InvalidEnvironment(t *testing.T) {
	service := NewBankService(&MockBankRepository{})

	_, err := service.GetBankDetails(context.Background(), "TEST001", "invalid")
	if err == nil {
		t.Fatal("Expected error for invalid environment")
	}

	if !strings.Contains(err.Error(), "invalid environment") {
		t.Errorf("Expected 'invalid environment' error, got %v", err)
	}
}

func TestBankService_GetBankDetails_ValidEnvNoConfig(t *testing.T) {
	expectedBank := &models.Bank{
		BankID:  "TEST001",
		Name:    "Test Bank",
		Country: "ES",
	}

	mockRepo := &MockBankRepository{
		GetBankByIDFunc: func(_ context.Context, _ string) (*models.Bank, error) {
			return expectedBank, nil
		},
		GetBankEnvironmentConfigsFunc: func(_ context.Context, _ string, _ string) (map[string]*models.BankEnvironmentConfig, error) {
			// Return empty map - no config for this environment
			return map[string]*models.BankEnvironmentConfig{}, nil
		},
	}

	service := NewBankService(mockRepo)

	_, err := service.GetBankDetails(context.Background(), "TEST001", "production")
	if err == nil {
		t.Fatal("Expected error when environment config not found")
	}

	if !strings.Contains(err.Error(), "environment configuration not found") {
		t.Errorf("Expected 'environment configuration not found' error, got %v", err)
	}
}

func TestBankService_GetBankDetails_RepositoryError(t *testing.T) {
	expectedBank := &models.Bank{
		BankID:  "TEST001",
		Name:    "Test Bank",
		Country: "ES",
	}

	mockRepo := &MockBankRepository{
		GetBankByIDFunc: func(_ context.Context, _ string) (*models.Bank, error) {
			return expectedBank, nil
		},
		GetBankEnvironmentConfigsFunc: func(_ context.Context, _ string, _ string) (map[string]*models.BankEnvironmentConfig, error) {
			return nil, errors.New("database connection error")
		},
	}

	service := NewBankService(mockRepo)

	_, err := service.GetBankDetails(context.Background(), "TEST001", "")
	if err == nil {
		t.Fatal("Expected error from repository")
	}

	if !strings.Contains(err.Error(), "failed to get environment configs") {
		t.Errorf("Expected repository error to be wrapped, got %v", err)
	}
}
