package secrets

// SecretProvider defines the interface for obtaining secrets from various sources.
// Each implementation is responsible for retrieving a specific secret from its configured source.
type SecretProvider interface {
	// GetSecret retrieves the secret from the configured source.
	// The implementation determines which specific secret to return.
	GetSecret() (string, error)
}
