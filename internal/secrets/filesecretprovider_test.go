package secrets

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSecretProvider_NewFileSecretProvider(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0", // Use random available port
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	// Should create initial secret
	secret, err := provider.GetSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, secret)

	// File should exist
	assert.FileExists(t, secretsFile)
}

func TestFileSecretProvider_GetSecret_WithExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	// Create initial secrets file
	initialStore := SecretStore{
		Current: SecretEntry{
			ID:        "test-id",
			Secret:    "test-secret",
			CreatedAt: time.Now(),
		},
		Deprecated: []SecretEntry{},
	}

	data, err := json.MarshalIndent(initialStore, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(secretsFile, data, 0o600))

	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0",
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	secret, err := provider.GetSecret()
	require.NoError(t, err)
	assert.Equal(t, "test-secret", secret)
}

func TestFileSecretProvider_GetAllSecrets(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0",
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	current, deprecated, err := provider.GetAllSecrets()
	require.NoError(t, err)
	assert.NotEmpty(t, current)
	assert.Empty(t, deprecated) // No deprecated secrets initially
}

func TestFileSecretProvider_SecretRotation(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0",
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	// Get initial secret
	initialSecret, err := provider.GetSecret()
	require.NoError(t, err)

	// Trigger rotation
	encryptedResponse, err := provider.handleRotateCommand()
	require.NoError(t, err)
	assert.NotEmpty(t, encryptedResponse)

	// Get new secret - should be different
	newSecret, err := provider.GetSecret()
	require.NoError(t, err)
	assert.NotEqual(t, initialSecret, newSecret)

	// Check deprecated secrets
	current, deprecated, err := provider.GetAllSecrets()
	require.NoError(t, err)
	assert.Equal(t, newSecret, current)
	assert.Len(t, deprecated, 1)
	assert.Equal(t, initialSecret, deprecated[0])

	// Decrypt response to verify it contains the new secret
	decryptedSecret, err := DecryptSecret(encryptedResponse, initialSecret)
	require.NoError(t, err)
	assert.Equal(t, newSecret, decryptedSecret)
}

func TestFileSecretProvider_MaxDeprecatedSecrets(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath:      secretsFile,
		TCPAddr:       ":0",
		MaxDeprecated: 2, // Only keep 2 deprecated secrets
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	// Perform multiple rotations
	for i := 0; i < 5; i++ {
		_, err := provider.handleRotateCommand()
		require.NoError(t, err)
	}

	// Should only have 2 deprecated secrets (max limit)
	_, deprecated, err := provider.GetAllSecrets()
	require.NoError(t, err)
	assert.Len(t, deprecated, 2)
}

func TestFileSecretProvider_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0",
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, err := provider.GetSecret()
				if err != nil {
					errors <- err
					return
				}
			}
		}()
	}

	// Concurrent rotations
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := provider.handleRotateCommand()
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestFileSecretProvider_InvalidFilePath(t *testing.T) {
	_, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file path cannot be empty")
}

func TestFileSecretProvider_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0",
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	// Check file permissions
	info, err := os.Stat(secretsFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestTCPRotationServer_StartStop(t *testing.T) {
	handler := func() (string, error) {
		return "encrypted-response", nil
	}

	server, err := NewTCPRotationServer(":0", handler)
	require.NoError(t, err)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	assert.True(t, server.IsRunning())

	// Stop server
	err = server.Stop()
	assert.NoError(t, err)
	assert.False(t, server.IsRunning())
}

func TestTCPRotationServer_RotateCommand(t *testing.T) {
	expectedResponse := "encrypted-new-secret"
	handler := func() (string, error) {
		return expectedResponse, nil
	}

	server, err := NewTCPRotationServer(":0", handler)
	require.NoError(t, err)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Start(ctx)
	}()
	defer func() { _ = server.Stop() }()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Connect and send rotate command
	conn, err := net.Dial("tcp", server.GetAddress())
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// Send rotate command
	_, err = conn.Write([]byte("rotate\n"))
	require.NoError(t, err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	require.NoError(t, err)

	var response RotationResponse
	err = json.Unmarshal(buffer[:n], &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, expectedResponse, response.EncryptedData)
	assert.Empty(t, response.Error)
}

func TestTCPRotationServer_InvalidCommand(t *testing.T) {
	handler := func() (string, error) {
		return "should-not-be-called", nil
	}

	server, err := NewTCPRotationServer(":0", handler)
	require.NoError(t, err)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Start(ctx)
	}()
	defer func() { _ = server.Stop() }()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Connect and send invalid command
	conn, err := net.Dial("tcp", server.GetAddress())
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// Send invalid command
	_, err = conn.Write([]byte("{\"action\":\"invalid\"}\n"))
	require.NoError(t, err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	require.NoError(t, err)

	var response RotationResponse
	err = json.Unmarshal(buffer[:n], &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Empty(t, response.EncryptedData)
	assert.Contains(t, response.Error, "Unknown command")
}
