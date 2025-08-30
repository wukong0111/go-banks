package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/models"
)

func TestBankFiltersService_GetAvailableFilters(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockBankRepository)
		expectedResult *models.BankFilters
		expectedError  error
	}{
		{
			name: "success - returns filters",
			mockSetup: func(mockRepo *MockBankRepository) {
				expectedFilters := &models.BankFilters{
					Countries: []models.CountryFilter{
						{Code: "ES", Name: "ES", Count: 10},
						{Code: "FR", Name: "FR", Count: 5},
					},
					APIs: []models.APIFilter{
						{Type: "berlin_group", Count: 12},
						{Type: "open_banking", Count: 3},
					},
					Environments: []string{"production", "sandbox", "test"},
					BankGroups: []models.BankGroupFilter{
						{GroupID: "group1", Name: "Group 1", Count: 8},
					},
				}
				mockRepo.On("GetAvailableFilters", mock.Anything).Return(expectedFilters, nil)
			},
			expectedResult: &models.BankFilters{
				Countries: []models.CountryFilter{
					{Code: "ES", Name: "ES", Count: 10},
					{Code: "FR", Name: "FR", Count: 5},
				},
				APIs: []models.APIFilter{
					{Type: "berlin_group", Count: 12},
					{Type: "open_banking", Count: 3},
				},
				Environments: []string{"production", "sandbox", "test"},
				BankGroups: []models.BankGroupFilter{
					{GroupID: "group1", Name: "Group 1", Count: 8},
				},
			},
			expectedError: nil,
		},
		{
			name: "error - repository fails",
			mockSetup: func(mockRepo *MockBankRepository) {
				mockRepo.On("GetAvailableFilters", mock.Anything).Return((*models.BankFilters)(nil), errors.New("database error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockBankRepository)
			tt.mockSetup(mockRepo)

			service := NewBankFiltersService(mockRepo)
			ctx := context.Background()

			// Execute
			result, err := service.GetAvailableFilters(ctx)

			// Assert
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
