package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/services"
)

type MockBankGroupService struct {
	mock.Mock
}

func (m *MockBankGroupService) GetBankGroups(ctx context.Context) ([]models.BankGroup, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.BankGroup), args.Error(1)
}

type MockBankGroupCreator struct {
	mock.Mock
}

func (m *MockBankGroupCreator) CreateBankGroup(ctx context.Context, request *services.CreateBankGroupRequest) (*models.BankGroup, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BankGroup), args.Error(1)
}

func TestBankGroupHandler_GetBankGroups_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockService, mockCreatorService)

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
	mockService.On("GetBankGroups", mock.Anything).Return(expectedGroups, nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/bank-groups", http.NoBody)

	// Execute
	handler.GetBankGroups(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse[[]models.BankGroup]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Nil(t, response.Error)
	assert.Equal(t, expectedGroups, response.Data)
	mockService.AssertExpectations(t)
}

func TestBankGroupHandler_GetBankGroups_EmptyResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockService, mockCreatorService)

	// Setup mock expectations
	mockService.On("GetBankGroups", mock.Anything).Return([]models.BankGroup{}, nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/bank-groups", http.NoBody)

	// Execute
	handler.GetBankGroups(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse[[]models.BankGroup]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Nil(t, response.Error)
	assert.Empty(t, response.Data)
	mockService.AssertExpectations(t)
}

func TestBankGroupHandler_GetBankGroups_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockService, mockCreatorService)

	expectedError := errors.New("database connection failed")

	// Setup mock expectations
	mockService.On("GetBankGroups", mock.Anything).Return([]models.BankGroup{}, expectedError)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/bank-groups", http.NoBody)

	// Execute
	handler.GetBankGroups(c)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "Failed to retrieve bank groups", *response.Error)
	mockService.AssertExpectations(t)
}
