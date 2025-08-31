package auth

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/wukong0111/go-banks/internal/logger"
	"github.com/wukong0111/go-banks/internal/secrets"
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
	logger logger.Logger
}

// NewJWTService creates a new JWT service with the provided secret provider, expiry and logger
func NewJWTService(provider secrets.SecretProvider, expiry time.Duration, log logger.Logger) (*JWTService, error) {
	// Get the secret from the provider on initialization
	secret, err := provider.GetSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT secret: %w", err)
	}

	return &JWTService{
		secret: []byte(secret),
		expiry: expiry,
		logger: log,
	}, nil
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
		j.logger.Error("failed to sign JWT token",
			"error", err.Error(),
			"subject", claims.Subject,
			"permissions", claims.Permissions,
			"expires_at", claims.ExpiresAt.Time,
			"signing_method", "HS256",
		)
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
		j.logger.Warn("failed to parse JWT token",
			"error", err.Error(),
			"token_length", len(tokenString),
		)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		j.logger.Warn("token validation failed - token is invalid",
			"token_length", len(tokenString),
		)
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		j.logger.Warn("token validation failed - invalid claims structure",
			"token_length", len(tokenString),
			"claims_type", fmt.Sprintf("%T", token.Claims),
		)
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
