package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/auth"
	"github.com/wukong0111/go-banks/internal/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_RequireAuth_ValidToken(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Generate valid token
	token, err := jwtService.GenerateToken([]string{"banks:read"})
	require.NoError(t, err)

	// Create gin router
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with valid token
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RequireAuth_MissingAuthorizationHeader(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Create gin router
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request without Authorization header
	req, _ := http.NewRequest("GET", "/test", http.NoBody)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_InvalidBearerFormat(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Create gin router
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test cases for invalid bearer format
	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "token-without-bearer"},
		{"Wrong prefix", "Basic dGVzdA=="},
		{"Empty Bearer", "Bearer "},
		{"Only Bearer", "Bearer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", http.NoBody)
			req.Header.Set("Authorization", tc.header)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Create gin router
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with invalid token
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_ExpiredToken(t *testing.T) {
	// Setup with very short expiry
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Nanosecond, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Generate token (will be expired immediately)
	token, err := jwtService.GenerateToken([]string{"banks:read"})
	require.NoError(t, err)

	// Wait to ensure token is expired
	time.Sleep(10 * time.Millisecond)

	// Create gin router
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with expired token
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireAuth_InsufficientPermissions(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Generate token with only read permission
	token, err := jwtService.GenerateToken([]string{"banks:read"})
	require.NoError(t, err)

	// Create gin router that requires write permission
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:write"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with insufficient permissions
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthMiddleware_RequireAuth_MultiplePermissions(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Generate token with read permission
	token, err := jwtService.GenerateToken([]string{"banks:read"})
	require.NoError(t, err)

	// Create gin router that requires either read OR write permission
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:read", "banks:write"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test - should pass because token has read permission
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RequireAuth_NoPermissionsRequired(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Generate token with any permission
	token, err := jwtService.GenerateToken([]string{"banks:read"})
	require.NoError(t, err)

	// Create gin router with no specific permission requirements
	router := gin.New()
	router.GET("/test", middleware.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test - should pass with valid token regardless of permissions
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_GetClaims(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Generate valid token
	token, err := jwtService.GenerateToken([]string{"banks:read", "banks:write"})
	require.NoError(t, err)

	// Create gin router
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:read"), func(c *gin.Context) {
		// Test GetClaims helper
		claims, exists := GetClaims(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "claims not found"})
			return
		}

		if claims.Subject != "api-client" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong subject"})
			return
		}

		if len(claims.Permissions) != 2 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong permissions count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"subject": claims.Subject, "permissions": claims.Permissions})
	})

	// Create request with valid token
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_GetPermissions(t *testing.T) {
	// Setup
	testLogger := logger.NewDiscardLogger()
	jwtService := auth.NewJWTService("test-secret", time.Hour, testLogger)
	middleware := NewAuthMiddleware(jwtService)

	// Generate valid token
	expectedPermissions := []string{"banks:read", "banks:write"}
	token, err := jwtService.GenerateToken(expectedPermissions)
	require.NoError(t, err)

	// Create gin router
	router := gin.New()
	router.GET("/test", middleware.RequireAuth("banks:read"), func(c *gin.Context) {
		// Test GetPermissions helper
		permissions, exists := GetPermissions(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permissions not found"})
			return
		}

		if len(permissions) != 2 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong permissions count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"permissions": permissions})
	})

	// Create request with valid token
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}
