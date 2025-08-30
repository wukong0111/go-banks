package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/services"
)

type BankCreatorHandler struct {
	creatorService services.BankCreator
}

func NewBankCreatorHandler(creatorService services.BankCreator) *BankCreatorHandler {
	return &BankCreatorHandler{
		creatorService: creatorService,
	}
}

func (h *BankCreatorHandler) CreateBank(c *gin.Context) {
	var request services.CreateBankRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Warn("invalid JSON request format",
				"error", err.Error(),
				"remote_addr", c.ClientIP(),
				"user_agent", c.GetHeader("User-Agent"),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"content_type", c.GetHeader("Content-Type"),
				"content_length", c.Request.ContentLength,
			)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	bank, err := h.creatorService.CreateBank(c.Request.Context(), &request)
	if err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Error("failed to create bank",
				"error", err,
				"request", request,
			)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create bank",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Bank created successfully",
		"data":    bank,
	})
}
