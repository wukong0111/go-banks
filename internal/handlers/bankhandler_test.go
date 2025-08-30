package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

// MockBankService is a mock implementation of the BankService for testing.
type MockBankService struct {
	mock.Mock
}

func (m *MockBankService) GetBanks(ctx context.Context, filters *repository.BankFilters) ([]models.Bank, *models.Pagination, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.Bank), args.Get(1).(*models.Pagination), args.Error(2)
}

func (m *MockBankService) GetBankDetails(ctx context.Context, bankID, environment string) (models.BankDetails, error) {
	args := m.Called(ctx, bankID, environment)
	// Handle the case where the first argument is nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.BankDetails), args.Error(1)
}

func TestBankHandler_GetBanks(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a mock service
	mockService := new(MockBankService)

	// Create a new handler with the mock service
	handler := NewBankHandler(mockService)

	// Create a new Gin engine
	router := gin.New()
	router.GET("/banks", handler.GetBanks)

	// Define the expected banks and pagination
	expectedBanks := []models.Bank{
		{BankID: "1", Name: "Bank A"},
		{BankID: "2", Name: "Bank B"},
	}
	expectedPagination := &models.Pagination{
		Total: 2,
		Page:  1,
		Limit: 10,
	}

	// Set up the mock service to return the expected banks
	mockService.On("GetBanks", mock.Anything, mock.AnythingOfType("*repository.BankFilters")).Return(expectedBanks, expectedPagination, nil)

	// Create a new HTTP request
	req, _ := http.NewRequest(http.MethodGet, "/banks", http.NoBody)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Serve the HTTP request
	router.ServeHTTP(w, req)

	// Assert the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert the response body
	var response models.APIResponse[[]models.Bank]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, expectedBanks, response.Data)
	assert.Equal(t, expectedPagination, response.Pagination)
	assert.Nil(t, response.Error)

	// Assert that the mock service was called
	mockService.AssertExpectations(t)
}
