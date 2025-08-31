package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/services"
)

type MockBankGroupUpdaterService struct {
	mock.Mock
}

func (m *MockBankGroupUpdaterService) UpdateBankGroup(ctx context.Context, groupID string, request *services.UpdateBankGroupRequest) (*models.BankGroup, error) {
	args := m.Called(ctx, groupID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BankGroup), args.Error(1)
}

func setupUpdateTestRouter(updaterService services.BankGroupUpdater) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create handler with mock services (use dummy implementations since we only test UpdateBankGroup)
	mockBankGroupService := &MockTestBankGroupService{}
	mockCreatorService := &MockTestBankGroupCreator{}
	handler := NewBankGroupHandler(mockBankGroupService, mockCreatorService, updaterService)

	router.PUT("/bank-groups/:groupId", handler.UpdateBankGroup)
	return router
}

type MockTestBankGroupService struct{}

func (m *MockTestBankGroupService) GetBankGroups(_ context.Context) ([]models.BankGroup, error) {
	return nil, nil
}

type MockTestBankGroupCreator struct{}

func (m *MockTestBankGroupCreator) CreateBankGroup(_ context.Context, _ *services.CreateBankGroupRequest) (*models.BankGroup, error) {
	return nil, nil
}

func TestBankGroupHandlerUpdateBankGroupSuccess(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	groupID := uuid.New()
	requestBody := services.UpdateBankGroupRequest{
		Name:        stringPtr("Updated Group"),
		Description: stringPtr("Updated description"),
	}

	expectedBankGroup := &models.BankGroup{
		GroupID:     groupID,
		Name:        "Updated Group",
		Description: stringPtr("Updated description"),
	}

	mockUpdaterService.On("UpdateBankGroup", mock.Anything, groupID.String(), mock.MatchedBy(func(req *services.UpdateBankGroupRequest) bool {
		return req.Name != nil && *req.Name == "Updated Group" &&
			req.Description != nil && *req.Description == "Updated description"
	})).Return(expectedBankGroup, nil)

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/bank-groups/"+groupID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse[*models.BankGroup]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.Equal(t, groupID, response.Data.GroupID)
	assert.Equal(t, "Updated Group", response.Data.Name)
	assert.Equal(t, "Updated description", *response.Data.Description)

	mockUpdaterService.AssertExpectations(t)
}

func TestBankGroupHandlerUpdateBankGroupPartialUpdate(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	groupID := uuid.New()
	requestBody := services.UpdateBankGroupRequest{
		Name: stringPtr("Only Name Updated"),
		// Description, LogoURL, Website not provided - should be partial update
	}

	expectedBankGroup := &models.BankGroup{
		GroupID:     groupID,
		Name:        "Only Name Updated",
		Description: stringPtr("Original description"), // Should preserve existing
	}

	mockUpdaterService.On("UpdateBankGroup", mock.Anything, groupID.String(), mock.MatchedBy(func(req *services.UpdateBankGroupRequest) bool {
		return req.Name != nil && *req.Name == "Only Name Updated" &&
			req.Description == nil && req.LogoURL == nil && req.Website == nil
	})).Return(expectedBankGroup, nil)

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/bank-groups/"+groupID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse[*models.BankGroup]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.Equal(t, "Only Name Updated", response.Data.Name)

	mockUpdaterService.AssertExpectations(t)
}

func TestBankGroupHandlerUpdateBankGroupMissingGroupID(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	requestBody := services.UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/bank-groups/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Gin returns 404 for missing path parameter
}

func TestBankGroupHandlerUpdateBankGroupEmptyGroupID(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	requestBody := services.UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/bank-groups/   ", bytes.NewBuffer(jsonBody)) // Whitespace only
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Contains(t, *response.Error, "Group ID is required")
}

func TestBankGroupHandlerUpdateBankGroupInvalidJSON(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	groupID := uuid.New()
	invalidJSON := `{"name": "Updated Group", "description": }` // Invalid JSON

	req := httptest.NewRequest("PUT", "/bank-groups/"+groupID.String(), bytes.NewBuffer([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Contains(t, *response.Error, "Invalid request format")
}

func TestBankGroupHandlerUpdateBankGroupGroupNotFound(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	groupID := uuid.New()
	requestBody := services.UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	mockUpdaterService.On("UpdateBankGroup", mock.Anything, groupID.String(), mock.Anything).
		Return(nil, errors.New("bank group not found"))

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/bank-groups/"+groupID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Contains(t, *response.Error, "Bank group not found")

	mockUpdaterService.AssertExpectations(t)
}

func TestBankGroupHandlerUpdateBankGroupValidationError(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	groupID := uuid.New()
	requestBody := services.UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	mockUpdaterService.On("UpdateBankGroup", mock.Anything, groupID.String(), mock.Anything).
		Return(nil, errors.New("invalid group_id: invalid UUID format"))

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/bank-groups/"+groupID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Contains(t, *response.Error, "Invalid request parameters")

	mockUpdaterService.AssertExpectations(t)
}

func TestBankGroupHandlerUpdateBankGroupEmptyNameValidationError(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	groupID := uuid.New()
	requestBody := services.UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	mockUpdaterService.On("UpdateBankGroup", mock.Anything, groupID.String(), mock.Anything).
		Return(nil, errors.New("name cannot be empty"))

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/bank-groups/"+groupID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Contains(t, *response.Error, "Invalid request parameters")

	mockUpdaterService.AssertExpectations(t)
}

func TestBankGroupHandlerUpdateBankGroupInternalServerError(t *testing.T) {
	// Arrange
	mockUpdaterService := &MockBankGroupUpdaterService{}
	router := setupUpdateTestRouter(mockUpdaterService)

	groupID := uuid.New()
	requestBody := services.UpdateBankGroupRequest{
		Name: stringPtr("Updated Group"),
	}

	mockUpdaterService.On("UpdateBankGroup", mock.Anything, groupID.String(), mock.Anything).
		Return(nil, errors.New("database connection failed"))

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/bank-groups/"+groupID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Contains(t, *response.Error, "Failed to update bank group")

	mockUpdaterService.AssertExpectations(t)
}
