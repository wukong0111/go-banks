package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/models"
)

// MockBankWriter implements the BankWriter interface for testing
type MockBankWriter struct {
	mock.Mock
}

func (m *MockBankWriter) CreateBank(ctx context.Context, bank *models.Bank) error {
	args := m.Called(ctx, bank)
	return args.Error(0)
}

func (m *MockBankWriter) CreateBankWithEnvironments(ctx context.Context, bank *models.Bank, configs []*models.BankEnvironmentConfig) error {
	args := m.Called(ctx, bank, configs)
	return args.Error(0)
}

func TestBankCreatorService_CreateBank_Simple(t *testing.T) {
	mockWriter := new(MockBankWriter)
	service := NewBankCreatorService(mockWriter)

	request := &CreateBankRequest{
		BankID:                 "test_bank_001",
		Name:                   "Test Bank",
		BankCodes:              []string{"0001"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "ES",
		AuthTypeChoiceRequired: true,
	}

	mockWriter.On("CreateBank", mock.Anything, mock.MatchedBy(func(bank *models.Bank) bool {
		return bank.Name == "Test Bank" &&
			bank.API == "berlin_group" &&
			bank.BankID != ""
	})).Return(nil)

	bank, err := service.CreateBank(context.Background(), request)
	require.NoError(t, err)

	assert.Equal(t, "Test Bank", bank.Name)
	assert.Equal(t, []string{"0001"}, bank.BankCodes)
	assert.Equal(t, "berlin_group", bank.API)
	assert.Equal(t, "1.3.6", bank.APIVersion)
	assert.Equal(t, "test_aspsp", bank.ASPSP)
	assert.Equal(t, "ES", bank.Country)
	assert.True(t, bank.AuthTypeChoiceRequired)
	assert.NotEmpty(t, bank.BankID)

	mockWriter.AssertExpectations(t)
}

func TestBankCreatorService_CreateBank_WithPredefinedID(t *testing.T) {
	mockWriter := new(MockBankWriter)
	service := NewBankCreatorService(mockWriter)

	request := &CreateBankRequest{
		BankID:                 "predefined_bank_id",
		Name:                   "Test Bank with ID",
		BankCodes:              []string{"0002"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "DE",
		AuthTypeChoiceRequired: false,
	}

	mockWriter.On("CreateBank", mock.Anything, mock.MatchedBy(func(bank *models.Bank) bool {
		return bank.BankID == "predefined_bank_id"
	})).Return(nil)

	bank, err := service.CreateBank(context.Background(), request)
	require.NoError(t, err)

	assert.Equal(t, "predefined_bank_id", bank.BankID)
	mockWriter.AssertExpectations(t)
}

func TestBankCreatorService_CreateBank_WithOptionalFields(t *testing.T) {
	mockWriter := new(MockBankWriter)
	service := NewBankCreatorService(mockWriter)

	bic := "TESTESMM"
	realName := "Real Bank Name"
	productCode := "PROD001"
	bankGroupID := uuid.New().String()
	logoURL := "https://example.com/logo.png"
	documentation := "Test docs"
	keywords := map[string]any{"key1": "value1"}
	attributes := map[string]any{"attr1": "val1"}

	request := &CreateBankRequest{
		BankID:                 "complete_bank_003",
		Name:                   "Complete Bank",
		BankCodes:              []string{"0003"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "FR",
		AuthTypeChoiceRequired: true,
		BIC:                    &bic,
		RealName:               &realName,
		ProductCode:            &productCode,
		BankGroupID:            &bankGroupID,
		LogoURL:                &logoURL,
		Documentation:          &documentation,
		Keywords:               keywords,
		Attribute:              attributes,
	}

	mockWriter.On("CreateBank", mock.Anything, mock.MatchedBy(func(bank *models.Bank) bool {
		return bank.BIC != nil && *bank.BIC == bic &&
			bank.RealName != nil && *bank.RealName == realName &&
			bank.BankGroupID != nil
	})).Return(nil)

	bank, err := service.CreateBank(context.Background(), request)
	require.NoError(t, err)

	assert.Equal(t, &bic, bank.BIC)
	assert.Equal(t, &realName, bank.RealName)
	assert.Equal(t, &productCode, bank.ProductCode)
	assert.NotNil(t, bank.BankGroupID)
	assert.Equal(t, &logoURL, bank.LogoURL)
	assert.Equal(t, &documentation, bank.Documentation)
	assert.Equal(t, keywords, bank.Keywords)
	assert.Equal(t, attributes, bank.Attribute)

	mockWriter.AssertExpectations(t)
}

func TestBankCreatorService_CreateBank_WithEnvironments(t *testing.T) {
	mockWriter := new(MockBankWriter)
	service := NewBankCreatorService(mockWriter)

	enabled := true
	blocked := false
	riskyMessage := "Risky bank"

	request := &CreateBankRequest{
		BankID:                 "env_bank_004",
		Name:                   "Bank with Environments",
		BankCodes:              []string{"0004"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "IT",
		AuthTypeChoiceRequired: false,
		Environments:           []string{"sandbox", "production"},
		Configuration: &EnvironmentConfig{
			Enabled:      &enabled,
			Blocked:      &blocked,
			RiskyMessage: &riskyMessage,
		},
	}

	mockWriter.On("CreateBankWithEnvironments", mock.Anything,
		mock.MatchedBy(func(bank *models.Bank) bool {
			return bank.Name == "Bank with Environments"
		}),
		mock.MatchedBy(func(configs []*models.BankEnvironmentConfig) bool {
			if len(configs) != 2 {
				return false
			}

			// Check that we have both environments
			hasProduction := false
			hasSandbox := false
			for _, config := range configs {
				if config.Environment == models.EnvironmentProduction {
					hasProduction = true
					assert.Equal(t, enabled, config.Enabled)
					assert.Equal(t, blocked, config.Blocked)
					assert.Equal(t, &riskyMessage, config.RiskyMessage)
				}
				if config.Environment == models.EnvironmentSandbox {
					hasSandbox = true
					assert.Equal(t, enabled, config.Enabled)
					assert.Equal(t, blocked, config.Blocked)
					assert.Equal(t, &riskyMessage, config.RiskyMessage)
				}
			}
			return hasProduction && hasSandbox
		}),
	).Return(nil)

	bank, err := service.CreateBank(context.Background(), request)
	require.NoError(t, err)
	assert.Equal(t, "Bank with Environments", bank.Name)

	mockWriter.AssertExpectations(t)
}

func TestBankCreatorService_CreateBank_WithConfigurations(t *testing.T) {
	mockWriter := new(MockBankWriter)
	service := NewBankCreatorService(mockWriter)

	sandboxEnabled := true
	prodEnabled := false
	appAuthRequired := true

	request := &CreateBankRequest{
		BankID:                 "config_bank_005",
		Name:                   "Bank with Configurations",
		BankCodes:              []string{"0005"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "NL",
		AuthTypeChoiceRequired: false,
		Configurations: map[string]*EnvironmentConfig{
			"sandbox": {
				Enabled:                    &sandboxEnabled,
				AppAuthSetupRequired:       &appAuthRequired,
				OkStatusCodesSimplePayment: []string{"200", "201"},
			},
			"production": {
				Enabled:                     &prodEnabled,
				OkStatusCodesInstantPayment: []string{"200"},
			},
		},
	}

	mockWriter.On("CreateBankWithEnvironments", mock.Anything,
		mock.MatchedBy(func(bank *models.Bank) bool {
			return bank.Name == "Bank with Configurations"
		}),
		mock.MatchedBy(func(configs []*models.BankEnvironmentConfig) bool {
			if len(configs) != 2 {
				return false
			}

			configMap := make(map[models.EnvironmentType]*models.BankEnvironmentConfig)
			for _, config := range configs {
				configMap[config.Environment] = config
			}

			sandboxConfig := configMap[models.EnvironmentSandbox]
			prodConfig := configMap[models.EnvironmentProduction]

			return sandboxConfig != nil &&
				sandboxConfig.Enabled == sandboxEnabled &&
				sandboxConfig.AppAuthSetupRequired == appAuthRequired &&
				len(sandboxConfig.OkStatusCodesSimplePayment) == 2 &&
				prodConfig != nil &&
				prodConfig.Enabled == prodEnabled &&
				len(prodConfig.OkStatusCodesInstantPayment) == 1
		}),
	).Return(nil)

	bank, err := service.CreateBank(context.Background(), request)
	require.NoError(t, err)
	assert.Equal(t, "Bank with Configurations", bank.Name)

	mockWriter.AssertExpectations(t)
}

func TestBankCreatorService_CreateBank_WriterError(t *testing.T) {
	mockWriter := new(MockBankWriter)
	service := NewBankCreatorService(mockWriter)

	request := &CreateBankRequest{
		BankID:                 "error_bank_006",
		Name:                   "Error Bank",
		BankCodes:              []string{"0006"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "BE",
		AuthTypeChoiceRequired: false,
	}

	mockWriter.On("CreateBank", mock.Anything, mock.Anything).
		Return(errors.New("database error"))

	bank, err := service.CreateBank(context.Background(), request)
	require.Error(t, err)
	assert.Nil(t, bank)
	assert.Contains(t, err.Error(), "failed to create bank")

	mockWriter.AssertExpectations(t)
}

func TestBankCreatorService_CreateBank_WithEnvironments_WriterError(t *testing.T) {
	mockWriter := new(MockBankWriter)
	service := NewBankCreatorService(mockWriter)

	enabled := true
	request := &CreateBankRequest{
		BankID:                 "error_env_bank_007",
		Name:                   "Error Bank with Environments",
		BankCodes:              []string{"0007"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "PT",
		AuthTypeChoiceRequired: false,
		Environments:           []string{"sandbox"},
		Configuration: &EnvironmentConfig{
			Enabled: &enabled,
		},
	}

	mockWriter.On("CreateBankWithEnvironments", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("transaction error"))

	bank, err := service.CreateBank(context.Background(), request)
	require.Error(t, err)
	assert.Nil(t, bank)
	assert.Contains(t, err.Error(), "failed to create bank with environments")

	mockWriter.AssertExpectations(t)
}

func TestBankCreatorService_RequestToBank_InvalidBankGroupID(t *testing.T) {
	service := NewBankCreatorService(nil)

	invalidGroupID := "invalid-uuid"
	request := &CreateBankRequest{
		BankID:                 "test_bank_008",
		Name:                   "Test Bank",
		BankCodes:              []string{"0008"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "AT",
		AuthTypeChoiceRequired: false,
		BankGroupID:            &invalidGroupID,
	}

	bank := service.requestToBank(request)

	// Invalid UUID should result in nil BankGroupID
	assert.Nil(t, bank.BankGroupID)
}

func TestBankCreatorService_RequestToBank_ValidBankGroupID(t *testing.T) {
	service := NewBankCreatorService(nil)

	validGroupID := uuid.New().String()
	request := &CreateBankRequest{
		BankID:                 "test_bank_009",
		Name:                   "Test Bank",
		BankCodes:              []string{"0009"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "CH",
		AuthTypeChoiceRequired: false,
		BankGroupID:            &validGroupID,
	}

	bank := service.requestToBank(request)

	require.NotNil(t, bank.BankGroupID)
	assert.Equal(t, validGroupID, bank.BankGroupID.String())
}
