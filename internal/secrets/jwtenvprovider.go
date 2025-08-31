package secrets

import (
	"errors"
	"os"
	"strings"
)

// JWTEnvProvider retrieves the JWT secret from environment variables.
type JWTEnvProvider struct {
	envVar string
}

// NewJWTEnvProvider creates a new JWT environment provider.
// It reads from the JWT_SECRET environment variable.
func NewJWTEnvProvider() *JWTEnvProvider {
	return &JWTEnvProvider{
		envVar: "JWT_SECRET",
	}
}

// GetSecret retrieves the JWT secret from the JWT_SECRET environment variable.
func (p *JWTEnvProvider) GetSecret() (string, error) {
	secret := strings.TrimSpace(os.Getenv(p.envVar))
	if secret == "" {
		return "", errors.New("JWT_SECRET environment variable not set")
	}
	return secret, nil
}
