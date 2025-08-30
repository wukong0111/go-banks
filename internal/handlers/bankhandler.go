package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
	"github.com/wukong0111/go-banks/internal/services"
)

type BankHandler struct {
	bankService services.BankService
}

func NewBankHandler(bankService services.BankService) *BankHandler {
	return &BankHandler{
		bankService: bankService,
	}
}

func (h *BankHandler) GetBanks(c *gin.Context) {
	// Parse query parameters - no defaults or validation, just HTTP parsing
	filters := &repository.BankFilters{
		Environment: c.Query("env"),
		Name:        c.Query("name"),
		API:         c.Query("api"),
		Country:     c.Query("country"),
	}

	// Parse and validate pagination parameters
	var page, limit int
	var err error

	// Validate page parameter if provided
	if pageStr, exists := c.GetQuery("page"); exists {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			if log, ok := logger.GetLogger(c); ok {
				log.Warn("invalid page parameter",
					"error", err.Error(),
					"page_value", pageStr,
					"remote_addr", c.ClientIP(),
					"user_agent", c.GetHeader("User-Agent"),
					"query_params", c.Request.URL.RawQuery,
				)
			}
			response := models.APIResponse[any]{
				Success: false,
				Error:   stringPtr("Invalid page parameter: must be a number"),
			}
			c.JSON(http.StatusBadRequest, response)
			return
		}
	}

	// Validate limit parameter if provided
	if limitStr, exists := c.GetQuery("limit"); exists {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			if log, ok := logger.GetLogger(c); ok {
				log.Warn("invalid limit parameter",
					"error", err.Error(),
					"limit_value", limitStr,
					"remote_addr", c.ClientIP(),
					"user_agent", c.GetHeader("User-Agent"),
					"query_params", c.Request.URL.RawQuery,
				)
			}
			response := models.APIResponse[any]{
				Success: false,
				Error:   stringPtr("Invalid limit parameter: must be a number"),
			}
			c.JSON(http.StatusBadRequest, response)
			return
		}
	}

	filters.Page = page
	filters.Limit = limit

	// Get banks from service
	banks, pagination, err := h.bankService.GetBanks(c.Request.Context(), filters)
	if err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Error("failed to retrieve banks",
				"error", err,
				"filters", filters,
			)
		}
		response := models.APIResponse[any]{
			Success: false,
			Error:   stringPtr("Failed to retrieve banks"),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Return successful response
	response := models.APIResponse[[]models.Bank]{
		Success:    true,
		Data:       banks,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

func (h *BankHandler) GetBankDetails(c *gin.Context) {
	// Extract bank ID from path parameter
	bankID := c.Param("bankId")
	if bankID == "" {
		if log, ok := logger.GetLogger(c); ok {
			log.Warn("missing bank ID parameter",
				"remote_addr", c.ClientIP(),
				"user_agent", c.GetHeader("User-Agent"),
				"path", c.Request.URL.Path,
				"query_params", c.Request.URL.RawQuery,
			)
		}
		response := models.APIResponse[any]{
			Success: false,
			Error:   stringPtr("Bank ID is required"),
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Extract optional environment parameter
	environment := c.Query("env")

	// Get bank details from service
	bankDetails, err := h.bankService.GetBankDetails(c.Request.Context(), bankID, environment)
	if err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Error("failed to retrieve bank details",
				"error", err,
				"bank_id", bankID,
				"environment", environment,
			)
		}

		// Handle specific error cases
		errorMessage := err.Error()
		switch {
		case strings.Contains(errorMessage, "bank not found"):
			response := models.APIResponse[any]{
				Success: false,
				Error:   stringPtr("Bank not found"),
			}
			c.JSON(http.StatusNotFound, response)
			return
		case strings.Contains(errorMessage, "invalid environment"):
			response := models.APIResponse[any]{
				Success: false,
				Error:   stringPtr(errorMessage),
			}
			c.JSON(http.StatusBadRequest, response)
			return
		case strings.Contains(errorMessage, "environment configuration not found"):
			response := models.APIResponse[any]{
				Success: false,
				Error:   stringPtr(errorMessage),
			}
			c.JSON(http.StatusNotFound, response)
			return
		}

		// Generic server error
		response := models.APIResponse[any]{
			Success: false,
			Error:   stringPtr("Failed to retrieve bank details"),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Return successful response
	response := models.APIResponse[any]{
		Success: true,
		Data:    bankDetails,
	}

	c.JSON(http.StatusOK, response)
}

func stringPtr(s string) *string {
	return &s
}
