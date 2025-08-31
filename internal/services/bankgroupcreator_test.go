package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/wukong0111/go-banks/internal/models"
)

// MockBankGroupWriter is a mock implementation of BankGroupWriter
type MockBankGroupWriter struct {
	mock.Mock
}

func (m *MockBankGroupWriter) CreateBankGroup(ctx context.Context, bankGroup *models.BankGroup) error {
	args := m.Called(ctx, bankGroup)
	return args.Error(0)
}

func (m *MockBankGroupWriter) UpdateBankGroup(ctx context.Context, bankGroup *models.BankGroup) error {
	args := m.Called(ctx, bankGroup)
	return args.Error(0)
}

func TestBankGroupCreatorService_CreateBankGroup_Success(t *testing.T) {
	ctx := context.Background()
	mockWriter := new(MockBankGroupWriter)
	service := NewBankGroupCreatorService(mockWriter)

	validUUID := uuid.New().String()
	request := &CreateBankGroupRequest{
		GroupID:     validUUID,
		Name:        "Test Group",
		Description: stringPtr("Test description"),
		LogoURL:     stringPtr("https://example.com/logo.png"),
		Website:     stringPtr("https://example.com"),
	}

	mockWriter.On("CreateBankGroup", ctx, mock.MatchedBy(func(bg *models.BankGroup) bool {
		return bg.Name == "Test Group" &&
			bg.Description != nil && *bg.Description == "Test description" &&
			bg.LogoURL != nil && *bg.LogoURL == "https://example.com/logo.png" &&
			bg.Website != nil && *bg.Website == "https://example.com"
	})).Return(nil)

	result, err := service.CreateBankGroup(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Group", result.Name)
	assert.NotNil(t, result.Description)
	assert.Equal(t, "Test description", *result.Description)
	mockWriter.AssertExpectations(t)
}

func TestBankGroupCreatorService_CreateBankGroup_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	service := NewBankGroupCreatorService(nil)

	request := &CreateBankGroupRequest{
		GroupID: "invalid-uuid",
		Name:    "Test Group",
	}

	result, err := service.CreateBankGroup(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid group_id")
	assert.Contains(t, err.Error(), "must be a valid UUID")
}

func TestBankGroupCreatorService_CreateBankGroup_EmptyUUID(t *testing.T) {
	ctx := context.Background()
	service := NewBankGroupCreatorService(nil)

	request := &CreateBankGroupRequest{
		GroupID: "",
		Name:    "Test Group",
	}

	result, err := service.CreateBankGroup(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid group_id")
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestBankGroupCreatorService_CreateBankGroup_EmptyName(t *testing.T) {
	ctx := context.Background()
	service := NewBankGroupCreatorService(nil)

	validUUID := uuid.New().String()
	request := &CreateBankGroupRequest{
		GroupID: validUUID,
		Name:    "",
	}

	result, err := service.CreateBankGroup(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestBankGroupCreatorService_CreateBankGroup_WhitespaceName(t *testing.T) {
	ctx := context.Background()
	service := NewBankGroupCreatorService(nil)

	validUUID := uuid.New().String()
	request := &CreateBankGroupRequest{
		GroupID: validUUID,
		Name:    "   ",
	}

	result, err := service.CreateBankGroup(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestBankGroupCreatorService_CreateBankGroup_DuplicateError(t *testing.T) {
	ctx := context.Background()
	mockWriter := new(MockBankGroupWriter)
	service := NewBankGroupCreatorService(mockWriter)

	validUUID := uuid.New().String()
	request := &CreateBankGroupRequest{
		GroupID: validUUID,
		Name:    "Test Group",
	}

	mockWriter.On("CreateBankGroup", ctx, mock.Anything).Return(errors.New("duplicate key value violates unique constraint"))

	result, err := service.CreateBankGroup(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already exists")
	mockWriter.AssertExpectations(t)
}

func TestBankGroupCreatorService_CreateBankGroup_RepositoryError(t *testing.T) {
	ctx := context.Background()
	mockWriter := new(MockBankGroupWriter)
	service := NewBankGroupCreatorService(mockWriter)

	validUUID := uuid.New().String()
	request := &CreateBankGroupRequest{
		GroupID: validUUID,
		Name:    "Test Group",
	}

	mockWriter.On("CreateBankGroup", ctx, mock.Anything).Return(errors.New("database connection failed"))

	result, err := service.CreateBankGroup(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create bank group")
	mockWriter.AssertExpectations(t)
}

func TestBankGroupCreatorService_CreateBankGroup_TrimWhitespace(t *testing.T) {
	ctx := context.Background()
	mockWriter := new(MockBankGroupWriter)
	service := NewBankGroupCreatorService(mockWriter)

	validUUID := uuid.New().String()
	description := "  Test description  "
	logoURL := "  https://example.com/logo.png  "
	website := "  https://example.com  "

	request := &CreateBankGroupRequest{
		GroupID:     "  " + validUUID + "  ",
		Name:        "  Test Group  ",
		Description: &description,
		LogoURL:     &logoURL,
		Website:     &website,
	}

	mockWriter.On("CreateBankGroup", ctx, mock.MatchedBy(func(bg *models.BankGroup) bool {
		return bg.Name == "Test Group" &&
			bg.Description != nil && *bg.Description == "Test description" &&
			bg.LogoURL != nil && *bg.LogoURL == "https://example.com/logo.png" &&
			bg.Website != nil && *bg.Website == "https://example.com"
	})).Return(nil)

	result, err := service.CreateBankGroup(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Group", result.Name)
	mockWriter.AssertExpectations(t)
}

func TestBankGroupCreatorService_CreateBankGroup_NilOptionalFields(t *testing.T) {
	ctx := context.Background()
	mockWriter := new(MockBankGroupWriter)
	service := NewBankGroupCreatorService(mockWriter)

	validUUID := uuid.New().String()
	request := &CreateBankGroupRequest{
		GroupID: validUUID,
		Name:    "Test Group",
	}

	mockWriter.On("CreateBankGroup", ctx, mock.MatchedBy(func(bg *models.BankGroup) bool {
		return bg.Name == "Test Group" &&
			bg.Description == nil &&
			bg.LogoURL == nil &&
			bg.Website == nil
	})).Return(nil)

	result, err := service.CreateBankGroup(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Group", result.Name)
	assert.Nil(t, result.Description)
	assert.Nil(t, result.LogoURL)
	assert.Nil(t, result.Website)
	mockWriter.AssertExpectations(t)
}

func TestBankGroupCreatorService_CreateBankGroup_EmptyOptionalFields(t *testing.T) {
	ctx := context.Background()
	mockWriter := new(MockBankGroupWriter)
	service := NewBankGroupCreatorService(mockWriter)

	validUUID := uuid.New().String()
	emptyString := ""
	whitespaceString := "   "

	request := &CreateBankGroupRequest{
		GroupID:     validUUID,
		Name:        "Test Group",
		Description: &emptyString,
		LogoURL:     &whitespaceString,
		Website:     &emptyString,
	}

	mockWriter.On("CreateBankGroup", ctx, mock.MatchedBy(func(bg *models.BankGroup) bool {
		return bg.Name == "Test Group" &&
			bg.Description == nil &&
			bg.LogoURL == nil &&
			bg.Website == nil
	})).Return(nil)

	result, err := service.CreateBankGroup(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.Description)
	assert.Nil(t, result.LogoURL)
	assert.Nil(t, result.Website)
	mockWriter.AssertExpectations(t)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
