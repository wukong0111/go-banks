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
	// Parse query parameters
	filters := repository.BankFilters{
		Environment: c.DefaultQuery("env", "all"),
		Name:        c.Query("name"),
		API:         c.Query("api"),
		Country:     c.Query("country"),
	}

	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	filters.Page = page

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
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
