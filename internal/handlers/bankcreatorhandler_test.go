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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/services"
)

// MockBankCreator implements the BankCreator interface for testing
type MockBankCreator struct {
	mock.Mock
}

func (m *MockBankCreator) CreateBank(ctx context.Context, request *services.CreateBankRequest) (*models.Bank, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bank), args.Error(1)
}

func TestBankCreatorHandler_CreateBank_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankCreator)
	handler := NewBankCreatorHandler(mockService)

	expectedBank := &models.Bank{
		BankID:                 "test_bank_001",
		Name:                   "Test Bank",
		BankCodes:              []string{"0001"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "ES",
		AuthTypeChoiceRequired: true,
	}

	mockService.On("CreateBank", mock.Anything, mock.MatchedBy(func(req *services.CreateBankRequest) bool {
		return req.Name == "Test Bank" &&
			req.API == "berlin_group" &&
			len(req.BankCodes) == 1 && req.BankCodes[0] == "0001"
	})).Return(expectedBank, nil)

	requestBody := services.CreateBankRequest{
		Name:                   "Test Bank",
		BankCodes:              []string{"0001"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "ES",
		AuthTypeChoiceRequired: true,
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/banks", handler.CreateBank)

	req, err := http.NewRequest("POST", "/banks", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Bank created successfully", response["message"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]any)
	assert.Equal(t, "test_bank_001", data["bank_id"])
	assert.Equal(t, "Test Bank", data["name"])

	mockService.AssertExpectations(t)
}

func TestBankCreatorHandler_CreateBank_WithEnvironments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankCreator)
	handler := NewBankCreatorHandler(mockService)

	expectedBank := &models.Bank{
		BankID:    "test_bank_002",
		Name:      "Bank with Environments",
		BankCodes: []string{"0002"},
		API:       "berlin_group",
		Country:   "DE",
	}

	mockService.On("CreateBank", mock.Anything, mock.MatchedBy(func(req *services.CreateBankRequest) bool {
		return req.Name == "Bank with Environments" &&
			len(req.Environments) == 2 &&
			req.Configuration != nil &&
			req.Configuration.Enabled != nil &&
			*req.Configuration.Enabled == true
	})).Return(expectedBank, nil)

	enabled := true
	blocked := false
	requestBody := services.CreateBankRequest{
		Name:                   "Bank with Environments",
		BankCodes:              []string{"0002"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "DE",
		AuthTypeChoiceRequired: false,
		Environments:           []string{"sandbox", "production"},
		Configuration: &services.EnvironmentConfig{
			Enabled:                    &enabled,
			Blocked:                    &blocked,
			OkStatusCodesSimplePayment: []string{"200", "201"},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/banks", handler.CreateBank)

	req, err := http.NewRequest("POST", "/banks", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Bank created successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestBankCreatorHandler_CreateBank_WithConfigurations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankCreator)
	handler := NewBankCreatorHandler(mockService)

	expectedBank := &models.Bank{
		BankID:    "test_bank_003",
		Name:      "Bank with Configurations",
		BankCodes: []string{"0003"},
		API:       "berlin_group",
		Country:   "FR",
	}

	mockService.On("CreateBank", mock.Anything, mock.MatchedBy(func(req *services.CreateBankRequest) bool {
		return req.Name == "Bank with Configurations" &&
			len(req.Configurations) == 2 &&
			req.Configurations["sandbox"] != nil &&
			req.Configurations["production"] != nil
	})).Return(expectedBank, nil)

	sandboxEnabled := true
	prodEnabled := false

	requestBody := services.CreateBankRequest{
		Name:                   "Bank with Configurations",
		BankCodes:              []string{"0003"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "FR",
		AuthTypeChoiceRequired: false,
		Configurations: map[string]*services.EnvironmentConfig{
			"sandbox": {
				Enabled:                    &sandboxEnabled,
				OkStatusCodesSimplePayment: []string{"200", "201"},
			},
			"production": {
				Enabled:                     &prodEnabled,
				OkStatusCodesInstantPayment: []string{"200"},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/banks", handler.CreateBank)

	req, err := http.NewRequest("POST", "/banks", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Bank created successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestBankCreatorHandler_CreateBank_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankCreator)
	handler := NewBankCreatorHandler(mockService)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/banks", handler.CreateBank)

	// Send invalid JSON
	req, err := http.NewRequest("POST", "/banks", bytes.NewBufferString("{invalid json"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid request format", response["error"])
	assert.NotNil(t, response["details"])

	// Service should not be called
	mockService.AssertNotCalled(t, "CreateBank")
}

func TestBankCreatorHandler_CreateBank_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankCreator)
	handler := NewBankCreatorHandler(mockService)

	// Missing required fields (name, bank_codes, api, etc.)
	requestBody := services.CreateBankRequest{
		Country: "ES",
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/banks", handler.CreateBank)

	req, err := http.NewRequest("POST", "/banks", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid request format", response["error"])

	// Service should not be called
	mockService.AssertNotCalled(t, "CreateBank")
}

func TestBankCreatorHandler_CreateBank_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankCreator)
	handler := NewBankCreatorHandler(mockService)

	mockService.On("CreateBank", mock.Anything, mock.Anything).
		Return(nil, errors.New("database connection failed"))

	requestBody := services.CreateBankRequest{
		Name:                   "Error Bank",
		BankCodes:              []string{"0004"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "IT",
		AuthTypeChoiceRequired: false,
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/banks", handler.CreateBank)

	req, err := http.NewRequest("POST", "/banks", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to create bank", response["error"])
	assert.Contains(t, response["details"], "database connection failed")

	mockService.AssertExpectations(t)
}

func TestBankCreatorHandler_CreateBank_WithOptionalFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBankCreator)
	handler := NewBankCreatorHandler(mockService)

	expectedBank := &models.Bank{
		BankID:    "test_bank_005",
		Name:      "Complete Bank",
		BankCodes: []string{"0005"},
	}

	mockService.On("CreateBank", mock.Anything, mock.MatchedBy(func(req *services.CreateBankRequest) bool {
		return req.Name == "Complete Bank" &&
			req.BIC != nil && *req.BIC == "TESTESMM" &&
			req.RealName != nil && *req.RealName == "Real Bank Name" &&
			req.Keywords != nil && req.Keywords["key1"] == "value1"
	})).Return(expectedBank, nil)

	bic := "TESTESMM"
	realName := "Real Bank Name"
	productCode := "PROD001"
	logoURL := "https://example.com/logo.png"
	documentation := "Test documentation"

	requestBody := services.CreateBankRequest{
		Name:                   "Complete Bank",
		BankCodes:              []string{"0005"},
		API:                    "berlin_group",
		APIVersion:             "1.3.6",
		ASPSP:                  "test_aspsp",
		Country:                "NL",
		AuthTypeChoiceRequired: true,
		BIC:                    &bic,
		RealName:               &realName,
		ProductCode:            &productCode,
		LogoURL:                &logoURL,
		Documentation:          &documentation,
		Keywords:               map[string]any{"key1": "value1", "key2": 123},
		Attribute:              map[string]any{"attr1": "val1", "attr2": true},
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/banks", handler.CreateBank)

	req, err := http.NewRequest("POST", "/banks", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Bank created successfully", response["message"])

	mockService.AssertExpectations(t)
}
