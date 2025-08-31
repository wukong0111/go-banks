package secrets

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTEnvProviderGetSecret(t *testing.T) {
	// Setup
	provider := NewJWTEnvProvider()

	// Set environment variable
	originalValue := os.Getenv("JWT_SECRET")
	defer func() {
		if originalValue != "" {
			_ = os.Setenv("JWT_SECRET", originalValue)
		} else {
			_ = os.Unsetenv("JWT_SECRET")
		}
	}()

	_ = os.Setenv("JWT_SECRET", "test-jwt-secret-123")

	// Act
	secret, err := provider.GetSecret()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "test-jwt-secret-123", secret)
}

func TestJWTEnvProviderGetSecretWithWhitespace(t *testing.T) {
	// Setup
	provider := NewJWTEnvProvider()

	// Set environment variable with whitespace
	originalValue := os.Getenv("JWT_SECRET")
	defer func() {
		if originalValue != "" {
			_ = os.Setenv("JWT_SECRET", originalValue)
		} else {
			_ = os.Unsetenv("JWT_SECRET")
		}
	}()

	_ = os.Setenv("JWT_SECRET", "  test-secret-with-spaces  ")

	// Act
	secret, err := provider.GetSecret()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "test-secret-with-spaces", secret)
}

func TestJWTEnvProviderGetSecretMissingEnvVar(t *testing.T) {
	// Setup
	provider := NewJWTEnvProvider()

	// Unset environment variable
	originalValue := os.Getenv("JWT_SECRET")
	defer func() {
		if originalValue != "" {
			_ = os.Setenv("JWT_SECRET", originalValue)
		} else {
			_ = os.Unsetenv("JWT_SECRET")
		}
	}()

	_ = os.Unsetenv("JWT_SECRET")

	// Act
	secret, err := provider.GetSecret()

	// Assert
	assert.Error(t, err)
	assert.Empty(t, secret)
	assert.Contains(t, err.Error(), "JWT_SECRET environment variable not set")
}

func TestJWTEnvProviderGetSecretEmptyValueError(t *testing.T) {
	// Setup
	provider := NewJWTEnvProvider()

	// Set empty environment variable
	originalValue := os.Getenv("JWT_SECRET")
	defer func() {
		if originalValue != "" {
			_ = os.Setenv("JWT_SECRET", originalValue)
		} else {
			_ = os.Unsetenv("JWT_SECRET")
		}
	}()

	_ = os.Setenv("JWT_SECRET", "")

	// Act
	secret, err := provider.GetSecret()

	// Assert
	assert.Error(t, err)
	assert.Empty(t, secret)
	assert.Contains(t, err.Error(), "JWT_SECRET environment variable not set")
}

func TestJWTEnvProviderGetSecretWhitespaceOnlyError(t *testing.T) {
	// Setup
	provider := NewJWTEnvProvider()

	// Set whitespace-only environment variable
	originalValue := os.Getenv("JWT_SECRET")
	defer func() {
		if originalValue != "" {
			_ = os.Setenv("JWT_SECRET", originalValue)
		} else {
			_ = os.Unsetenv("JWT_SECRET")
		}
	}()

	_ = os.Setenv("JWT_SECRET", "   ")

	// Act
	secret, err := provider.GetSecret()

	// Assert
	assert.Error(t, err)
	assert.Empty(t, secret)
	assert.Contains(t, err.Error(), "JWT_SECRET environment variable not set")
}
