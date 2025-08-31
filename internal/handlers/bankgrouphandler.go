package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/services"
)

type BankGroupHandler struct {
	bankGroupService services.BankGroupService
	creatorService   services.BankGroupCreator
}

func NewBankGroupHandler(bankGroupService services.BankGroupService, creatorService services.BankGroupCreator) *BankGroupHandler {
	return &BankGroupHandler{
		bankGroupService: bankGroupService,
		creatorService:   creatorService,
	}
}

func (h *BankGroupHandler) GetBankGroups(c *gin.Context) {
	// Get bank groups from service
	bankGroups, err := h.bankGroupService.GetBankGroups(c.Request.Context())
	if err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Error("failed to retrieve bank groups",
				"error", err,
			)
		}
		response := models.APIResponse[any]{
			Success: false,
			Error:   stringPtr("Failed to retrieve bank groups"),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Return successful response
	response := models.APIResponse[[]models.BankGroup]{
		Success: true,
		Data:    bankGroups,
	}

	c.JSON(http.StatusOK, response)
}

func (h *BankGroupHandler) CreateBankGroup(c *gin.Context) {
	var request services.CreateBankGroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Warn("invalid JSON request format",
				"error", err.Error(),
				"remote_addr", c.ClientIP(),
				"path", c.Request.URL.Path,
			)
		}
		response := models.APIResponse[any]{
			Success: false,
			Error:   stringPtr("Invalid request format"),
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	bankGroup, err := h.creatorService.CreateBankGroup(c.Request.Context(), &request)
	if err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Error("failed to create bank group",
				"error", err,
				"group_id", request.GroupID,
			)
		}

		// Map service errors to HTTP status codes
		var statusCode int
		var errorMessage string

		switch {
		case strings.Contains(err.Error(), "already exists"):
			statusCode = http.StatusConflict
			errorMessage = "Bank group already exists"
		case strings.Contains(err.Error(), "invalid group_id") ||
			strings.Contains(err.Error(), "name cannot be empty"):
			statusCode = http.StatusBadRequest
			errorMessage = "Invalid request data"
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Failed to create bank group"
		}

		response := models.APIResponse[any]{
			Success: false,
			Error:   &errorMessage,
		}
		c.JSON(statusCode, response)
		return
	}

	response := models.APIResponse[*models.BankGroup]{
		Success: true,
		Data:    bankGroup,
	}
	c.JSON(http.StatusCreated, response)
}
