package secrets

import (
	"errors"
	"os"
	"strings"
)

// JWTEnvProvider retrieves the JWT secret from environment variables.
type JWTEnvProvider struct {
	envVar       string
	defaultValue string
}

// NewJWTEnvProvider creates a new JWT environment provider.
// It reads from the JWT_SECRET environment variable with a fallback default.
func NewJWTEnvProvider() *JWTEnvProvider {
	return &JWTEnvProvider{
		envVar:       "JWT_SECRET",
		defaultValue: "your-super-secret-jwt-key",
	}
}

// GetSecret retrieves the JWT secret from the JWT_SECRET environment variable.
func (p *JWTEnvProvider) GetSecret() (string, error) {
	secret := os.Getenv(p.envVar)
	if secret == "" {
		secret = p.defaultValue
	}

	// Trim any whitespace that might have been accidentally added
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errors.New("JWT secret cannot be empty")
	}

	return secret, nil
}
