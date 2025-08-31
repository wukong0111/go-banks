package secrets

// MockSecretProvider is a mock implementation of SecretProvider for testing.
type MockSecretProvider struct {
	secret string
	err    error
}

// NewMockSecretProvider creates a new mock secret provider that returns the specified secret.
func NewMockSecretProvider(secret string) *MockSecretProvider {
	return &MockSecretProvider{
		secret: secret,
		err:    nil,
	}
}

// NewMockSecretProviderWithError creates a new mock secret provider that returns an error.
func NewMockSecretProviderWithError(err error) *MockSecretProvider {
	return &MockSecretProvider{
		secret: "",
		err:    err,
	}
}

// GetSecret returns the configured secret or error.
func (m *MockSecretProvider) GetSecret() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.secret, nil
}
