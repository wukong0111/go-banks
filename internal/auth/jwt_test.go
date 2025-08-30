package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wukong0111/go-banks/internal/logger"
)

func TestJWTService_GenerateToken(t *testing.T) {
	secret := "test-secret-key"
	expiry := time.Hour
	testLogger := logger.NewDiscardLogger()
	service := NewJWTService(secret, expiry, testLogger)

	permissions := []string{"banks:read", "banks:write"}

	token, err := service.GenerateToken(permissions)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the generated token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)

	// Check claims
	assert.Equal(t, "api-client", claims.Subject)
	assert.Equal(t, "go-banks-api", claims.Issuer)
	assert.Len(t, claims.Permissions, 2)
	assert.Contains(t, claims.Permissions, "banks:read")
	assert.Contains(t, claims.Permissions, "banks:write")
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	testLogger := logger.NewDiscardLogger()
	service := NewJWTService("test-secret", time.Hour, testLogger)

	// Test with invalid token
	_, err := service.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestJWTService_ValidateToken_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	testLogger := logger.NewDiscardLogger()
	service := NewJWTService(secret, -time.Hour, testLogger) // Expired

	permissions := []string{"banks:read"}
	token, err := service.GenerateToken(permissions)
	require.NoError(t, err)

	// Wait a moment to ensure token is expired
	time.Sleep(10 * time.Millisecond)

	// Validate the expired token
	_, err = service.ValidateToken(token)
	assert.Error(t, err)
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	testLogger := logger.NewDiscardLogger()
	service1 := NewJWTService("secret1", time.Hour, testLogger)
	service2 := NewJWTService("secret2", time.Hour, testLogger)

	permissions := []string{"banks:read"}
	token, err := service1.GenerateToken(permissions)
	require.NoError(t, err)

	// Try to validate with different secret
	_, err = service2.ValidateToken(token)
	assert.Error(t, err)
}

func TestClaims_HasPermission(t *testing.T) {
	claims := &Claims{
		Permissions: []string{"banks:read", "banks:write"},
	}

	// Test existing permission
	assert.True(t, claims.HasPermission("banks:read"))
	assert.True(t, claims.HasPermission("banks:write"))

	// Test non-existing permission
	assert.False(t, claims.HasPermission("banks:delete"))
}

func TestClaims_HasAnyPermission(t *testing.T) {
	claims := &Claims{
		Permissions: []string{"banks:read"},
	}

	// Test with some matching permissions
	required := []string{"banks:read", "banks:write"}
	assert.True(t, claims.HasAnyPermission(required))

	// Test with no matching permissions
	required = []string{"banks:write", "banks:delete"}
	assert.False(t, claims.HasAnyPermission(required))

	// Test with empty required permissions
	required = []string{}
	assert.False(t, claims.HasAnyPermission(required))
}

func TestClaims_HasAllPermissions(t *testing.T) {
	claims := &Claims{
		Permissions: []string{"banks:read", "banks:write"},
	}

	// Test with all permissions present
	required := []string{"banks:read", "banks:write"}
	assert.True(t, claims.HasAllPermissions(required))

	// Test with subset of permissions
	required = []string{"banks:read"}
	assert.True(t, claims.HasAllPermissions(required))

	// Test with missing permission
	required = []string{"banks:read", "banks:write", "banks:delete"}
	assert.False(t, claims.HasAllPermissions(required))

	// Test with empty required permissions
	required = []string{}
	assert.True(t, claims.HasAllPermissions(required))
}

func TestJWTService_TokenStructure(t *testing.T) {
	testLogger := logger.NewDiscardLogger()
	service := NewJWTService("test-secret", time.Hour, testLogger)
	permissions := []string{"banks:read"}

	tokenString, err := service.GenerateToken(permissions)
	require.NoError(t, err)

	// Parse token without validation to check structure
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(*Claims)
	require.True(t, ok)

	// Check token structure
	assert.NotEmpty(t, claims.Subject)
	assert.NotEmpty(t, claims.Issuer)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
}
