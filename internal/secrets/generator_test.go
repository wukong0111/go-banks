package secrets

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecureSecret(t *testing.T) {
	tests := []struct {
		name       string
		byteLength int
		shouldFail bool
	}{
		{"Valid 32 bytes", 32, false},
		{"Valid 16 bytes", 16, false},
		{"Valid 64 bytes", 64, false},
		{"Invalid 0 bytes", 0, true},
		{"Invalid negative bytes", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := GenerateSecureSecret(tt.byteLength)

			if tt.shouldFail {
				assert.Error(t, err)
				assert.Empty(t, secret)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, secret)

				// Should be valid base64
				decoded, err := base64.StdEncoding.DecodeString(secret)
				require.NoError(t, err)
				assert.Len(t, decoded, tt.byteLength)
			}
		})
	}
}

func TestGenerateSecureSecret_Uniqueness(t *testing.T) {
	secrets := make(map[string]bool)

	// Generate 100 secrets and ensure they're all unique
	for range 100 {
		secret, err := GenerateSecureSecret(32)
		require.NoError(t, err)

		// Should not have seen this secret before
		assert.False(t, secrets[secret], "Generated duplicate secret")
		secrets[secret] = true
	}
}

func TestEncryptDecryptSecret(t *testing.T) {
	newSecret := "new-secret-value-12345"
	oldSecret := "test-encryption-key-for-testing"

	// Encrypt
	encrypted, err := EncryptSecret(newSecret, oldSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// Should be valid base64
	_, err = base64.StdEncoding.DecodeString(encrypted)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := DecryptSecret(encrypted, oldSecret)
	require.NoError(t, err)
	assert.Equal(t, newSecret, decrypted)
}

func TestEncryptDecryptSecret_WrongKey(t *testing.T) {
	newSecret := "new-secret-value"
	oldSecret := "old-secret-key"
	wrongKey := "wrong-key"

	// Encrypt with old secret
	encrypted, err := EncryptSecret(newSecret, oldSecret)
	require.NoError(t, err)

	// Try to decrypt with wrong key
	_, err = DecryptSecret(encrypted, wrongKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt")
}

func TestEncryptSecret_EmptyValues(t *testing.T) {
	tests := []struct {
		name      string
		newSecret string
		oldSecret string
	}{
		{"Empty new secret", "", "valid-key"},
		{"Empty old secret", "valid-secret", ""},
		{"Both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptSecret(tt.newSecret, tt.oldSecret)
			require.NoError(t, err) // Should not fail, but will encrypt empty string

			decrypted, err := DecryptSecret(encrypted, tt.oldSecret)
			require.NoError(t, err)
			assert.Equal(t, tt.newSecret, decrypted)
		})
	}
}

func TestDecryptSecret_InvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		encrypted string
		key       string
		errorMsg  string
	}{
		{
			"Invalid base64",
			"not-valid-base64!@#",
			"valid-key",
			"failed to decode base64",
		},
		{
			"Too short ciphertext",
			base64.StdEncoding.EncodeToString([]byte("short")),
			"valid-key",
			"ciphertext too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptSecret(tt.encrypted, tt.key)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestEncryptDecrypt_LongSecret(t *testing.T) {
	// Test with a very long secret
	newSecret := strings.Repeat("abcdefghijklmnopqrstuvwxyz", 100) // 2600 characters
	oldSecret := "encryption-key-12345"

	encrypted, err := EncryptSecret(newSecret, oldSecret)
	require.NoError(t, err)

	decrypted, err := DecryptSecret(encrypted, oldSecret)
	require.NoError(t, err)
	assert.Equal(t, newSecret, decrypted)
}

func TestEncryptDecrypt_UnicodeSecret(t *testing.T) {
	// Test with unicode characters
	newData := "test-unicode-content-新しい"
	encryptionKey := "test-unicode-passphrase-старый"

	encrypted, err := EncryptSecret(newData, encryptionKey)
	require.NoError(t, err)

	decrypted, err := DecryptSecret(encrypted, encryptionKey)
	require.NoError(t, err)
	assert.Equal(t, newData, decrypted)
}
