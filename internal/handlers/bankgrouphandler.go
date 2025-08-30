package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/services"
)

type BankGroupHandler struct {
	bankGroupService services.BankGroupService
}

func NewBankGroupHandler(bankGroupService services.BankGroupService) *BankGroupHandler {
	return &BankGroupHandler{
		bankGroupService: bankGroupService,
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
