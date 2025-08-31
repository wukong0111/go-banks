package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SecretEntry represents a single secret with metadata
type SecretEntry struct {
	ID        string     `json:"id"`
	Secret    string     `json:"secret"`
	CreatedAt time.Time  `json:"created_at"`
	RotatedAt *time.Time `json:"rotated_at,omitempty"`
}

// SecretStore represents the structure of the secrets file
type SecretStore struct {
	Current    SecretEntry   `json:"current"`
	Deprecated []SecretEntry `json:"deprecated"`
}

// FileSecretProvider manages secrets stored in a file with rotation capabilities
type FileSecretProvider struct {
	filePath      string
	tcpAddr       string
	store         SecretStore
	mu            sync.RWMutex
	server        *TCPRotationServer
	ctx           context.Context
	cancelFunc    context.CancelFunc
	maxDeprecated int
}

// FileSecretProviderOptions configures the FileSecretProvider
type FileSecretProviderOptions struct {
	FilePath      string
	TCPAddr       string
	MaxDeprecated int // Maximum number of deprecated secrets to keep
}

// NewFileSecretProvider creates a new FileSecretProvider
func NewFileSecretProvider(opts FileSecretProviderOptions) (*FileSecretProvider, error) {
	if opts.FilePath == "" {
		return nil, errors.New("file path cannot be empty")
	}
	if opts.TCPAddr == "" {
		opts.TCPAddr = ":8888" // Default TCP port
	}
	if opts.MaxDeprecated <= 0 {
		opts.MaxDeprecated = 5 // Default max deprecated secrets
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	provider := &FileSecretProvider{
		filePath:      opts.FilePath,
		tcpAddr:       opts.TCPAddr,
		ctx:           ctx,
		cancelFunc:    cancelFunc,
		maxDeprecated: opts.MaxDeprecated,
	}

	// Initialize secrets
	if err := provider.initialize(); err != nil {
		cancelFunc()
		return nil, fmt.Errorf("failed to initialize secrets: %w", err)
	}

	// Start TCP server for rotation commands
	server, err := NewTCPRotationServer(opts.TCPAddr, provider.handleRotateCommand)
	if err != nil {
		cancelFunc()
		return nil, fmt.Errorf("failed to create TCP server: %w", err)
	}
	provider.server = server

	// Start server in background
	go func() {
		if err := server.Start(ctx); err != nil && err != context.Canceled {
			// In a real application, you'd want to log this error properly
			fmt.Printf("TCP server error: %v\n", err)
		}
	}()

	return provider, nil
}

// GetSecret returns the current active secret
func (f *FileSecretProvider) GetSecret() (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.store.Current.Secret == "" {
		return "", errors.New("no active secret available")
	}

	return f.store.Current.Secret, nil
}

// GetAllSecrets returns current and deprecated secrets for validation purposes
func (f *FileSecretProvider) GetAllSecrets() (current string, deprecated []string, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	current = f.store.Current.Secret
	deprecated = make([]string, len(f.store.Deprecated))

	for i := range f.store.Deprecated {
		deprecated[i] = f.store.Deprecated[i].Secret
	}

	return current, deprecated, nil
}

// Close stops the TCP server and cleans up resources
func (f *FileSecretProvider) Close() error {
	f.cancelFunc()
	if f.server != nil {
		return f.server.Stop()
	}
	return nil
}

// initialize loads or creates the secrets file
func (f *FileSecretProvider) initialize() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Try to load existing file
	if err := f.loadFromFile(); err != nil {
		// If file doesn't exist, create new secret
		if os.IsNotExist(err) {
			return f.createInitialSecret()
		}
		return fmt.Errorf("failed to load secrets file: %w", err)
	}

	return nil
}

// loadFromFile loads secrets from the file
func (f *FileSecretProvider) loadFromFile() error {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &f.store); err != nil {
		return fmt.Errorf("failed to parse secrets file: %w", err)
	}

	// Validate that we have a current secret
	if f.store.Current.Secret == "" {
		return errors.New("no current secret found in file")
	}

	return nil
}

// createInitialSecret creates the first secret and saves it
func (f *FileSecretProvider) createInitialSecret() error {
	secret, err := GenerateSecureSecret(32) // 32 bytes = 256 bits
	if err != nil {
		return fmt.Errorf("failed to generate initial secret: %w", err)
	}

	f.store = SecretStore{
		Current: SecretEntry{
			ID:        uuid.New().String(),
			Secret:    secret,
			CreatedAt: time.Now(),
		},
		Deprecated: []SecretEntry{},
	}

	return f.saveToFile()
}

// saveToFile saves the current store to file with appropriate permissions
func (f *FileSecretProvider) saveToFile() error {
	data, err := json.MarshalIndent(f.store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(f.filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}

	return nil
}

// handleRotateCommand handles the rotation command from TCP
func (f *FileSecretProvider) handleRotateCommand() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Generate new secret
	newSecret, err := GenerateSecureSecret(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate new secret: %w", err)
	}

	// Encrypt new secret with current secret
	encryptedResponse, err := EncryptSecret(newSecret, f.store.Current.Secret)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt new secret: %w", err)
	}

	// Mark current secret as rotated and move to deprecated
	now := time.Now()
	f.store.Current.RotatedAt = &now
	f.store.Deprecated = append(f.store.Deprecated, f.store.Current)

	// Set new secret as current
	f.store.Current = SecretEntry{
		ID:        uuid.New().String(),
		Secret:    newSecret,
		CreatedAt: now,
	}

	// Clean up old deprecated secrets if we exceed the limit
	if len(f.store.Deprecated) > f.maxDeprecated {
		// Keep only the most recent ones
		f.store.Deprecated = f.store.Deprecated[len(f.store.Deprecated)-f.maxDeprecated:]
	}

	// Save to file
	if err := f.saveToFile(); err != nil {
		return "", fmt.Errorf("failed to save rotated secrets: %w", err)
	}

	return encryptedResponse, nil
}
