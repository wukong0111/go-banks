package auth

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims for our API
type Claims struct {
	jwt.RegisteredClaims
	Permissions []string `json:"permissions"`
}

// JWTService handles JWT token generation and validation
type JWTService struct {
	secret []byte
	expiry time.Duration
}

// NewJWTService creates a new JWT service with the provided secret and expiry
func NewJWTService(secret string, expiry time.Duration) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// GenerateToken generates a JWT token with the given permissions
func (j *JWTService) GenerateToken(permissions []string) (string, error) {
	now := time.Now()

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "api-client",
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-banks-api",
		},
		Permissions: permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token, returning the claims
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Ensure the token method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// HasPermission checks if the claims contain the required permission
func (c *Claims) HasPermission(required string) bool {
	return slices.Contains(c.Permissions, required)
}

// HasAnyPermission checks if the claims contain any of the required permissions
func (c *Claims) HasAnyPermission(required []string) bool {
	return slices.ContainsFunc(required, c.HasPermission)
}

// HasAllPermissions checks if the claims contain all of the required permissions
func (c *Claims) HasAllPermissions(required []string) bool {
	return !slices.ContainsFunc(required, func(req string) bool {
		return !c.HasPermission(req)
	})
}
