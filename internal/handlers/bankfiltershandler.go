package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/services"
)

type BankFiltersHandler struct {
	filtersService services.BankFiltersService
}

func NewBankFiltersHandler(filtersService services.BankFiltersService) *BankFiltersHandler {
	return &BankFiltersHandler{
		filtersService: filtersService,
	}
}

func (h *BankFiltersHandler) GetFilters(c *gin.Context) {
	ctx := c.Request.Context()

	filters, err := h.filtersService.GetAvailableFilters(ctx)
	if err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Error("failed to get available filters",
				"error", err,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"user_agent", c.Request.UserAgent(),
			)
		}
		errorMsg := "Internal server error"
		c.JSON(http.StatusInternalServerError, models.APIResponse[any]{
			Success: false,
			Error:   &errorMsg,
		})
		return
	}

	response := models.APIResponse[*models.BankFilters]{
		Success: true,
		Data:    filters,
	}

	c.JSON(http.StatusOK, response)
}
