package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/services"
)

type BankUpdaterHandler struct {
	updaterService services.BankUpdater
}

func NewBankUpdaterHandler(updaterService services.BankUpdater) *BankUpdaterHandler {
	return &BankUpdaterHandler{
		updaterService: updaterService,
	}
}

func (h *BankUpdaterHandler) UpdateBank(c *gin.Context) {
	// Get bank ID from path parameter
	bankID := strings.TrimSpace(c.Param("bankId"))
	if bankID == "" {
		if log, ok := logger.GetLogger(c); ok {
			log.Warn("missing bank ID in path parameter",
				"remote_addr", c.ClientIP(),
				"user_agent", c.GetHeader("User-Agent"),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success":   false,
			"error":     "Bank ID is required",
			"timestamp": "2024-01-01T00:00:00Z", // placeholder for now
		})
		return
	}

	// Parse request body
	var request services.UpdateBankRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Warn("invalid JSON request format",
				"error", err.Error(),
				"bank_id", bankID,
				"remote_addr", c.ClientIP(),
				"user_agent", c.GetHeader("User-Agent"),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"content_type", c.GetHeader("Content-Type"),
				"content_length", c.Request.ContentLength,
			)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success":   false,
			"error":     "Invalid request format",
			"details":   err.Error(),
			"timestamp": "2024-01-01T00:00:00Z", // placeholder for now
		})
		return
	}

	// Call service to update bank
	response, err := h.updaterService.UpdateBank(c.Request.Context(), bankID, &request)
	if err != nil {
		if log, ok := logger.GetLogger(c); ok {
			log.Error("failed to update bank",
				"error", err,
				"bank_id", bankID,
				"request", request,
			)
		}

		// Check if it's a not found error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success":   false,
				"error":     "Bank not found",
				"timestamp": "2024-01-01T00:00:00Z", // placeholder for now
			})
			return
		}

		// Check if it's a validation error
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "format") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":   false,
				"error":     "Invalid request parameters",
				"details":   err.Error(),
				"timestamp": "2024-01-01T00:00:00Z", // placeholder for now
			})
			return
		}

		// Generic server error
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":   false,
			"error":     "Failed to update bank",
			"details":   err.Error(),
			"timestamp": "2024-01-01T00:00:00Z", // placeholder for now
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"bank":                response.Bank,
			"environment_configs": response.EnvironmentConfigs,
		},
		"message": "Bank updated successfully",
	})
}
