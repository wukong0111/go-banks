package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wukong0111/go-banks/internal/auth"
	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/models"
)

// AuthMiddleware handles JWT authentication and authorization
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware creates a new auth middleware with the provided JWT service
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth creates a middleware that requires JWT authentication with specific permissions
func (a *AuthMiddleware) RequireAuth(requiredPermissions ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			if log, ok := logger.GetLogger(c); ok {
				log.Warn("missing authorization header",
					"remote_addr", c.ClientIP(),
					"user_agent", c.GetHeader("User-Agent"),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)
			}
			a.respondUnauthorized(c, "Authorization header is required")
			return
		}

		// Check if the header starts with "Bearer "
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			if log, ok := logger.GetLogger(c); ok {
				log.Warn("invalid authorization header format",
					"remote_addr", c.ClientIP(),
					"user_agent", c.GetHeader("User-Agent"),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"auth_header_prefix", func() string {
						if len(authHeader) > 10 {
							return authHeader[:10] + "..."
						}
						return authHeader
					}(),
				)
			}
			a.respondUnauthorized(c, "Authorization header must use Bearer token")
			return
		}

		// Extract the token
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			if log, ok := logger.GetLogger(c); ok {
				log.Warn("empty bearer token",
					"remote_addr", c.ClientIP(),
					"user_agent", c.GetHeader("User-Agent"),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)
			}
			a.respondUnauthorized(c, "Bearer token is required")
			return
		}

		// Validate the token
		claims, err := a.jwtService.ValidateToken(tokenString)
		if err != nil {
			if log, ok := logger.GetLogger(c); ok {
				log.Warn("token validation failed",
					"error", err.Error(),
					"remote_addr", c.ClientIP(),
					"user_agent", c.GetHeader("User-Agent"),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"token_length", len(tokenString),
				)
			}
			a.respondUnauthorized(c, "Invalid or expired token")
			return
		}

		// Check permissions if required
		if len(requiredPermissions) > 0 {
			if !claims.HasAnyPermission(requiredPermissions) {
				if log, ok := logger.GetLogger(c); ok {
					log.Warn("insufficient permissions",
						"user_id", claims.Subject,
						"required_permissions", requiredPermissions,
						"user_permissions", claims.Permissions,
						"remote_addr", c.ClientIP(),
						"user_agent", c.GetHeader("User-Agent"),
						"path", c.Request.URL.Path,
						"method", c.Request.Method,
					)
				}
				a.respondForbidden(c, "Insufficient permissions")
				return
			}
		}

		// Store claims in context for use in handlers
		c.Set("claims", claims)
		c.Set("user_id", claims.Subject)
		c.Set("permissions", claims.Permissions)

		// Continue to the next handler
		c.Next()
	})
}

// GetClaims retrieves JWT claims from the Gin context
func GetClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	jwtClaims, ok := claims.(*auth.Claims)
	return jwtClaims, ok
}

// GetPermissions retrieves permissions from the Gin context
func GetPermissions(c *gin.Context) ([]string, bool) {
	permissions, exists := c.Get("permissions")
	if !exists {
		return nil, false
	}

	perms, ok := permissions.([]string)
	return perms, ok
}

// respondUnauthorized sends a 401 Unauthorized response
func (a *AuthMiddleware) respondUnauthorized(c *gin.Context, message string) {
	response := models.APIResponse[any]{
		Success: false,
		Error:   &message,
	}
	c.JSON(http.StatusUnauthorized, response)
	c.Abort()
}

// respondForbidden sends a 403 Forbidden response
func (a *AuthMiddleware) respondForbidden(c *gin.Context, message string) {
	response := models.APIResponse[any]{
		Success: false,
		Error:   &message,
	}
	c.JSON(http.StatusForbidden, response)
	c.Abort()
}
