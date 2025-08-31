package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/services"
)

// MockBankGroupCreator is defined in bankgrouphandler_test.go

func TestBankGroupHandler_CreateBankGroup_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGroupService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockGroupService, mockCreatorService)

	validUUID := uuid.New()
	request := services.CreateBankGroupRequest{
		GroupID: validUUID.String(),
		Name:    "Test Group",
	}

	expectedGroup := &models.BankGroup{
		GroupID: validUUID,
		Name:    "Test Group",
	}

	mockCreatorService.On("CreateBankGroup", mock.Anything, mock.MatchedBy(func(req *services.CreateBankGroupRequest) bool {
		return req.GroupID == validUUID.String() && req.Name == "Test Group"
	})).Return(expectedGroup, nil)

	// Create request
	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/bank-groups", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	handler.CreateBankGroup(c)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.APIResponse[*models.BankGroup]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.Equal(t, "Test Group", response.Data.Name)
	assert.Equal(t, validUUID, response.Data.GroupID)

	mockCreatorService.AssertExpectations(t)
}

func TestBankGroupHandler_CreateBankGroup_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGroupService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockGroupService, mockCreatorService)

	// Create request with invalid JSON
	req, _ := http.NewRequest("POST", "/api/bank-groups", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateBankGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "Invalid request format", *response.Error)
}

func TestBankGroupHandler_CreateBankGroup_MissingRequiredField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGroupService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockGroupService, mockCreatorService)

	// Request missing required field (name)
	request := map[string]any{
		"group_id": uuid.New().String(),
	}

	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/bank-groups", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateBankGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "Invalid request format", *response.Error)
}

func TestBankGroupHandler_CreateBankGroup_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGroupService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockGroupService, mockCreatorService)

	request := services.CreateBankGroupRequest{
		GroupID: "invalid-uuid",
		Name:    "Test Group",
	}

	mockCreatorService.On("CreateBankGroup", mock.Anything, mock.Anything).Return(nil, errors.New("invalid group_id: must be a valid UUID"))

	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/bank-groups", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateBankGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "Invalid request data", *response.Error)

	mockCreatorService.AssertExpectations(t)
}

func TestBankGroupHandler_CreateBankGroup_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGroupService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockGroupService, mockCreatorService)

	request := services.CreateBankGroupRequest{
		GroupID: uuid.New().String(),
		Name:    "   ", // Whitespace name that will pass JSON binding but fail service validation
	}

	mockCreatorService.On("CreateBankGroup", mock.Anything, mock.Anything).Return(nil, errors.New("name cannot be empty"))

	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/bank-groups", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateBankGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "Invalid request data", *response.Error)

	mockCreatorService.AssertExpectations(t)
}

func TestBankGroupHandler_CreateBankGroup_DuplicateGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGroupService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockGroupService, mockCreatorService)

	validUUID := uuid.New().String()
	request := services.CreateBankGroupRequest{
		GroupID: validUUID,
		Name:    "Test Group",
	}

	mockCreatorService.On("CreateBankGroup", mock.Anything, mock.Anything).Return(nil, errors.New("bank group with ID '"+validUUID+"' already exists"))

	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/bank-groups", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateBankGroup(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "Bank group already exists", *response.Error)

	mockCreatorService.AssertExpectations(t)
}

func TestBankGroupHandler_CreateBankGroup_InternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockGroupService := new(MockBankGroupService)
	mockCreatorService := new(MockBankGroupCreator)
	handler := NewBankGroupHandler(mockGroupService, mockCreatorService)

	request := services.CreateBankGroupRequest{
		GroupID: uuid.New().String(),
		Name:    "Test Group",
	}

	mockCreatorService.On("CreateBankGroup", mock.Anything, mock.Anything).Return(nil, errors.New("database connection failed"))

	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/bank-groups", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateBankGroup(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "Failed to create bank group", *response.Error)

	mockCreatorService.AssertExpectations(t)
}
