package secrets

import (
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullRotationFlow tests the complete secret rotation flow end-to-end
func TestFullRotationFlow(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	// Create FileSecretProvider
	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath:      secretsFile,
		TCPAddr:       ":0", // Use random port
		MaxDeprecated: 3,
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	// Step 1: Verify initial secret exists
	initialSecret, err := provider.GetSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, initialSecret)
	fmt.Printf("Initial secret: %s\n", initialSecret[:8]+"...")

	// Step 2: Wait for server to be ready and get actual address
	time.Sleep(200 * time.Millisecond)
	assert.True(t, provider.server.IsRunning())

	serverAddr := provider.server.GetAddress()
	fmt.Printf("TCP server listening on: %s\n", serverAddr)

	// Step 3: Connect via TCP and send rotation command
	conn, err := net.Dial("tcp", serverAddr)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// Send rotation command as JSON
	rotateCmd := RotationCommand{Action: "rotate"}
	cmdBytes, err := json.Marshal(rotateCmd)
	require.NoError(t, err)

	_, err = conn.Write(append(cmdBytes, '\n'))
	require.NoError(t, err)
	fmt.Println("Sent rotation command via TCP")

	// Step 4: Read encrypted response
	buffer := make([]byte, 2048)
	n, err := conn.Read(buffer)
	require.NoError(t, err)

	var response RotationResponse
	err = json.Unmarshal(buffer[:n], &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotEmpty(t, response.EncryptedData)
	assert.Empty(t, response.Error)
	fmt.Printf("Received encrypted response: %s\n", response.EncryptedData[:16]+"...")

	// Step 5: Verify new secret is different
	newSecret, err := provider.GetSecret()
	require.NoError(t, err)
	assert.NotEqual(t, initialSecret, newSecret)
	fmt.Printf("New secret: %s\n", newSecret[:8]+"...")

	// Step 6: Decrypt response using old secret to verify it contains new secret
	decryptedSecret, err := DecryptSecret(response.EncryptedData, initialSecret)
	require.NoError(t, err)
	assert.Equal(t, newSecret, decryptedSecret)
	fmt.Println("Successfully decrypted response - matches new secret")

	// Step 7: Verify deprecated secrets are tracked
	current, deprecated, err := provider.GetAllSecrets()
	require.NoError(t, err)
	assert.Equal(t, newSecret, current)
	assert.Len(t, deprecated, 1)
	assert.Equal(t, initialSecret, deprecated[0])
	fmt.Printf("Deprecated secrets count: %d\n", len(deprecated))

	// Step 8: Perform multiple rotations to test limit enforcement
	for i := 0; i < 5; i++ {
		// Send another rotation command
		_, err = conn.Write(append(cmdBytes, '\n'))
		require.NoError(t, err)

		// Read response
		n, err = conn.Read(buffer)
		require.NoError(t, err)

		err = json.Unmarshal(buffer[:n], &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	}

	// Step 9: Verify max deprecated limit is enforced
	_, deprecated, err = provider.GetAllSecrets()
	require.NoError(t, err)
	assert.LessOrEqual(t, len(deprecated), 3, "Should not exceed max deprecated limit")
	fmt.Printf("Final deprecated secrets count: %d (max: 3)\n", len(deprecated))

	fmt.Println("✅ Full rotation flow test completed successfully!")
}

// TestMultipleClientsRotation tests handling multiple simultaneous rotation requests
func TestMultipleClientsRotation(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	provider, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0",
	})
	require.NoError(t, err)
	defer func() { _ = provider.Close() }()

	initialSecret, err := provider.GetSecret()
	require.NoError(t, err)

	// Wait for server
	time.Sleep(100 * time.Millisecond)
	serverAddr := provider.server.GetAddress()

	// Simulate multiple clients trying to rotate simultaneously
	numClients := 3
	responses := make(chan RotationResponse, numClients)
	errors := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				errors <- fmt.Errorf("client %d dial error: %w", clientID, err)
				return
			}
			defer func() { _ = conn.Close() }()

			rotateCmd := RotationCommand{Action: "rotate"}
			cmdBytes, err := json.Marshal(rotateCmd)
			if err != nil {
				errors <- fmt.Errorf("client %d marshal error: %w", clientID, err)
				return
			}

			_, err = conn.Write(append(cmdBytes, '\n'))
			if err != nil {
				errors <- fmt.Errorf("client %d write error: %w", clientID, err)
				return
			}

			buffer := make([]byte, 2048)
			n, err := conn.Read(buffer)
			if err != nil {
				errors <- fmt.Errorf("client %d read error: %w", clientID, err)
				return
			}

			var response RotationResponse
			err = json.Unmarshal(buffer[:n], &response)
			if err != nil {
				errors <- fmt.Errorf("client %d unmarshal error: %w", clientID, err)
				return
			}

			responses <- response
		}(i)
	}

	// Collect responses
	successCount := 0
	for i := 0; i < numClients; i++ {
		select {
		case response := <-responses:
			if response.Success {
				successCount++
				assert.NotEmpty(t, response.EncryptedData)
			} else {
				t.Logf("Client received error: %s", response.Error)
			}
		case err := <-errors:
			t.Errorf("Client error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for client responses")
		}
	}

	// At least one rotation should have succeeded
	assert.Greater(t, successCount, 0, "At least one rotation should succeed")

	// Verify final secret is different from initial
	finalSecret, err := provider.GetSecret()
	require.NoError(t, err)
	assert.NotEqual(t, initialSecret, finalSecret)

	fmt.Printf("Multiple clients test: %d/%d successful rotations\n", successCount, numClients)
}

// TestSecretPersistence tests that secrets persist across provider restarts
func TestSecretPersistence(t *testing.T) {
	tempDir := t.TempDir()
	secretsFile := filepath.Join(tempDir, "secrets.json")

	// Create first provider instance
	provider1, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0",
	})
	require.NoError(t, err)

	// Get initial secret and rotate
	secret1, err := provider1.GetSecret()
	require.NoError(t, err)

	_, err = provider1.handleRotateCommand()
	require.NoError(t, err)

	secret2, err := provider1.GetSecret()
	require.NoError(t, err)
	assert.NotEqual(t, secret1, secret2)

	// Close first provider
	_ = provider1.Close()

	// Create second provider instance with same file
	provider2, err := NewFileSecretProvider(FileSecretProviderOptions{
		FilePath: secretsFile,
		TCPAddr:  ":0",
	})
	require.NoError(t, err)
	defer func() { _ = provider2.Close() }()

	// Verify secret persistence
	persistedSecret, err := provider2.GetSecret()
	require.NoError(t, err)
	assert.Equal(t, secret2, persistedSecret)

	// Verify deprecated secrets are also persisted
	_, deprecated, err := provider2.GetAllSecrets()
	require.NoError(t, err)
	assert.Len(t, deprecated, 1)
	assert.Equal(t, secret1, deprecated[0])

	fmt.Println("✅ Secret persistence test completed successfully!")
}
