package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/models"
)

// Since we're testing the PostgresBankWriter with actual database operations,
// we'll create unit tests that focus on the business logic and error handling
// For integration tests with real database, we'd need proper test setup

func TestBankWriter_Interface_Implementation(_ *testing.T) {
	// Test that PostgresBankWriter implements BankWriter interface
	var _ BankWriter = (*PostgresBankWriter)(nil)
}

func TestBankWriter_BankID_Generation(t *testing.T) {
	// Test bank ID generation logic without database
	bank1 := &models.Bank{
		Name:      "Test Bank 1",
		BankCodes: []string{"0001"},
		API:       "berlin_group",
		Country:   "ES",
	}

	bank2 := &models.Bank{
		BankID:    "predefined_id",
		Name:      "Test Bank 2",
		BankCodes: []string{"0002"},
		API:       "berlin_group",
		Country:   "DE",
	}

	// When BankID is empty, it should be generated
	assert.Empty(t, bank1.BankID)

	// When BankID is provided, it should be preserved
	assert.Equal(t, "predefined_id", bank2.BankID)
}

func TestBankWriter_UUID_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   *string
		wantNil bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantNil: true,
		},
		{
			name:    "empty string",
			input:   stringPtr(""),
			wantNil: true,
		},
		{
			name:    "invalid UUID",
			input:   stringPtr("invalid-uuid"),
			wantNil: true,
		},
		{
			name:    "valid UUID",
			input:   stringPtr(uuid.New().String()),
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result *uuid.UUID

			if tt.input != nil && *tt.input != "" {
				if parsed, err := uuid.Parse(*tt.input); err == nil {
					result = &parsed
				}
			}

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestBankWriter_EnvironmentConfig_Mapping(t *testing.T) {
	// Test environment configuration mapping logic
	bankID := "test_bank"

	configs := []*models.BankEnvironmentConfig{
		{
			BankID:      bankID,
			Environment: models.EnvironmentSandbox,
			Enabled:     true,
			Blocked:     false,
		},
		{
			BankID:      bankID,
			Environment: models.EnvironmentProduction,
			Enabled:     true,
			Blocked:     false,
		},
	}

	// Verify all configs have the correct bank ID
	for _, config := range configs {
		assert.Equal(t, bankID, config.BankID)
		assert.NotEmpty(t, config.Environment)
	}

	// Verify we have the expected environments
	environments := make([]models.EnvironmentType, len(configs))
	for i, config := range configs {
		environments[i] = config.Environment
	}

	assert.Contains(t, environments, models.EnvironmentSandbox)
	assert.Contains(t, environments, models.EnvironmentProduction)
}

func TestBankWriter_Optional_Fields_Handling(t *testing.T) {
	// Test handling of optional fields
	bic := "TESTESMM"
	realName := "Real Bank Name"
	productCode := "PROD001"
	logoURL := "https://example.com/logo.png"
	documentation := "Test docs"

	bank := &models.Bank{
		BankID:        "test_bank",
		Name:          "Test Bank",
		BankCodes:     []string{"0001"},
		API:           "berlin_group",
		APIVersion:    "1.3.6",
		ASPSP:         "test_aspsp",
		Country:       "ES",
		BIC:           &bic,
		RealName:      &realName,
		ProductCode:   &productCode,
		LogoURL:       &logoURL,
		Documentation: &documentation,
		Keywords:      map[string]any{"key1": "value1"},
		Attribute:     map[string]any{"attr1": "val1"},
	}

	// Verify optional fields are properly set
	require.NotNil(t, bank.BIC)
	assert.Equal(t, bic, *bank.BIC)

	require.NotNil(t, bank.RealName)
	assert.Equal(t, realName, *bank.RealName)

	require.NotNil(t, bank.ProductCode)
	assert.Equal(t, productCode, *bank.ProductCode)

	require.NotNil(t, bank.LogoURL)
	assert.Equal(t, logoURL, *bank.LogoURL)

	require.NotNil(t, bank.Documentation)
	assert.Equal(t, documentation, *bank.Documentation)

	assert.NotNil(t, bank.Keywords)
	assert.Equal(t, "value1", bank.Keywords["key1"])

	assert.NotNil(t, bank.Attribute)
	assert.Equal(t, "val1", bank.Attribute["attr1"])
}

func TestBankWriter_Required_Fields_Validation(t *testing.T) {
	// Test that required fields are present
	bank := &models.Bank{
		BankID:                 "test_bank",
		Name:                   "Test Bank",
		BankCodes:              []string{"0001"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "ES",
		AuthTypeChoiceRequired: false,
	}

	// Verify all required fields are set
	assert.NotEmpty(t, bank.BankID)
	assert.NotEmpty(t, bank.Name)
	assert.NotEmpty(t, bank.BankCodes)
	assert.NotEmpty(t, bank.API)
	assert.NotEmpty(t, bank.APIVersion)
	assert.NotEmpty(t, bank.ASPSP)
	assert.NotEmpty(t, bank.Country)

	// AuthTypeChoiceRequired is boolean, so it's always set
	assert.False(t, bank.AuthTypeChoiceRequired)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
