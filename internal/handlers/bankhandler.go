package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
	"github.com/wukong0111/go-banks/internal/services"
)

type BankHandler struct {
	bankService *services.BankService
}

func NewBankHandler(bankService *services.BankService) *BankHandler {
	return &BankHandler{
		bankService: bankService,
	}
}

func (h *BankHandler) GetBanks(c *gin.Context) {
	// Parse query parameters - no defaults or validation, just HTTP parsing
	filters := repository.BankFilters{
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
		log.Printf("Error getting banks: %v", err)
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

func stringPtr(s string) *string {
	return &s
}
