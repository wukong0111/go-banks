package auth

import (
	"slices"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTService_GenerateToken(t *testing.T) {
	secret := "test-secret-key"
	expiry := time.Hour
	service := NewJWTService(secret, expiry)

	permissions := []string{"banks:read", "banks:write"}

	token, err := service.GenerateToken(permissions)
	if err != nil {
		t.Fatalf("Expected no error generating token, got %v", err)
	}

	if token == "" {
		t.Fatal("Expected non-empty token")
	}

	// Validate the generated token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected no error validating token, got %v", err)
	}

	// Check claims
	if claims.Subject != "api-client" {
		t.Errorf("Expected subject 'api-client', got %s", claims.Subject)
	}

	if claims.Issuer != "go-banks-api" {
		t.Errorf("Expected issuer 'go-banks-api', got %s", claims.Issuer)
	}

	if len(claims.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(claims.Permissions))
	}

	if !contains(claims.Permissions, "banks:read") {
		t.Error("Expected permissions to contain 'banks:read'")
	}

	if !contains(claims.Permissions, "banks:write") {
		t.Error("Expected permissions to contain 'banks:write'")
	}
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour)

	// Test with invalid token
	_, err := service.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}
}

func TestJWTService_ValidateToken_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret, -time.Hour) // Expired

	permissions := []string{"banks:read"}
	token, err := service.GenerateToken(permissions)
	if err != nil {
		t.Fatalf("Expected no error generating token, got %v", err)
	}

	// Wait a moment to ensure token is expired
	time.Sleep(10 * time.Millisecond)

	// Validate the expired token
	_, err = service.ValidateToken(token)
	if err == nil {
		t.Fatal("Expected error for expired token")
	}
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	service1 := NewJWTService("secret1", time.Hour)
	service2 := NewJWTService("secret2", time.Hour)

	permissions := []string{"banks:read"}
	token, err := service1.GenerateToken(permissions)
	if err != nil {
		t.Fatalf("Expected no error generating token, got %v", err)
	}

	// Try to validate with different secret
	_, err = service2.ValidateToken(token)
	if err == nil {
		t.Fatal("Expected error when validating with wrong secret")
	}
}

func TestClaims_HasPermission(t *testing.T) {
	claims := &Claims{
		Permissions: []string{"banks:read", "banks:write"},
	}

	// Test existing permission
	if !claims.HasPermission("banks:read") {
		t.Error("Expected HasPermission to return true for 'banks:read'")
	}

	if !claims.HasPermission("banks:write") {
		t.Error("Expected HasPermission to return true for 'banks:write'")
	}

	// Test non-existing permission
	if claims.HasPermission("banks:delete") {
		t.Error("Expected HasPermission to return false for 'banks:delete'")
	}
}

func TestClaims_HasAnyPermission(t *testing.T) {
	claims := &Claims{
		Permissions: []string{"banks:read"},
	}

	// Test with some matching permissions
	required := []string{"banks:read", "banks:write"}
	if !claims.HasAnyPermission(required) {
		t.Error("Expected HasAnyPermission to return true when one permission matches")
	}

	// Test with no matching permissions
	required = []string{"banks:write", "banks:delete"}
	if claims.HasAnyPermission(required) {
		t.Error("Expected HasAnyPermission to return false when no permissions match")
	}

	// Test with empty required permissions
	required = []string{}
	if claims.HasAnyPermission(required) {
		t.Error("Expected HasAnyPermission to return false for empty required permissions")
	}
}

func TestClaims_HasAllPermissions(t *testing.T) {
	claims := &Claims{
		Permissions: []string{"banks:read", "banks:write"},
	}

	// Test with all permissions present
	required := []string{"banks:read", "banks:write"}
	if !claims.HasAllPermissions(required) {
		t.Error("Expected HasAllPermissions to return true when all permissions are present")
	}

	// Test with subset of permissions
	required = []string{"banks:read"}
	if !claims.HasAllPermissions(required) {
		t.Error("Expected HasAllPermissions to return true for subset of permissions")
	}

	// Test with missing permission
	required = []string{"banks:read", "banks:write", "banks:delete"}
	if claims.HasAllPermissions(required) {
		t.Error("Expected HasAllPermissions to return false when missing a permission")
	}

	// Test with empty required permissions
	required = []string{}
	if !claims.HasAllPermissions(required) {
		t.Error("Expected HasAllPermissions to return true for empty required permissions")
	}
}

func TestJWTService_TokenStructure(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour)
	permissions := []string{"banks:read"}

	tokenString, err := service.GenerateToken(permissions)
	if err != nil {
		t.Fatalf("Expected no error generating token, got %v", err)
	}

	// Parse token without validation to check structure
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		t.Fatalf("Expected no error parsing token structure, got %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		t.Fatal("Expected Claims type")
	}

	// Check token structure
	if claims.Subject == "" {
		t.Error("Expected non-empty subject")
	}

	if claims.Issuer == "" {
		t.Error("Expected non-empty issuer")
	}

	if claims.ExpiresAt == nil {
		t.Error("Expected ExpiresAt to be set")
	}

	if claims.IssuedAt == nil {
		t.Error("Expected IssuedAt to be set")
	}

	if claims.NotBefore == nil {
		t.Error("Expected NotBefore to be set")
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
