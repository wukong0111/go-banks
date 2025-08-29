package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger creates a middleware that logs HTTP requests
func RequestLogger(log Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.New().String()

		// Enrich logger with request context
		reqLog := log.With(
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Store logger and request ID in context
		c.Set("logger", reqLog)
		c.Set("request_id", requestID)

		// Log request start
		reqLog.Info("request started")

		// Process request
		c.Next()

		// Log request completion
		duration := time.Since(start)
		status := c.Writer.Status()

		logFunc := reqLog.Info
		if status >= 400 && status < 500 {
			logFunc = reqLog.Warn
		} else if status >= 500 {
			logFunc = reqLog.Error
		}

		logFunc("request completed",
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"size", c.Writer.Size(),
		)

		// Log any errors that occurred
		if len(c.Errors) > 0 {
			reqLog.Error("request errors", "errors", c.Errors.String())
		}
	}
}

// GetLogger extracts the logger from gin context
func GetLogger(c *gin.Context) (Logger, bool) {
	if logger, exists := c.Get("logger"); exists {
		if log, ok := logger.(Logger); ok {
			return log, true
		}
	}
	return nil, false
}

// MustGetLogger extracts the logger from gin context, panics if not found
func MustGetLogger(c *gin.Context) Logger {
	log, exists := GetLogger(c)
	if !exists {
		panic("logger not found in context")
	}
	return log
}

// GetRequestID extracts the request ID from gin context
func GetRequestID(c *gin.Context) (string, bool) {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id, true
		}
	}
	return "", false
}
